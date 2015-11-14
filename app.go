package app

import (
    "net/http"
    "appengine"
    "appengine/urlfetch"
    "net/url"
    "encoding/json"
    "regexp"
    chatwork "github.com/eiel/go-chatwork"
    "appengine/taskqueue"
    "appengine/memcache"
    "strconv"
)

type LastTime struct {
    Time int64
}

func init() {
    http.HandleFunc("/", handler)
    http.HandleFunc("/task", taskHandler)
}

type Body struct {
    ResponseData ResponseData `json:"responseData"`
}

type ResponseData struct {
    Results []Result `json"results"`
}

type Result struct {
    UnescapedUrl string `json:"unescapedUrl"`
}

var ROOM_ID = "xxxxxxxxxxxxxxxxxxxxx"
var TOKEN = "xxxxxxxxxxxxxxxxxxxxxxxx"
var BOT_ACCOUNT_ID = 1111111111111111111

func handler(w http.ResponseWriter, r *http.Request) {

    c := appengine.NewContext(r)

    client := urlfetch.Client(c)
    roomId := ROOM_ID
    cwClient := chatwork.NewClient(TOKEN)
    cwClient.HttpClient = client
    messages := cwClient.RoomMessages(roomId,map[string]string {
        "force": "1",
    })
    //contents, _ := ioutil.ReadAll(resp.Body)
    //fmt.Fprint(w, string(contents))

    //if err != nil {
//        w.WriteHeader(http.StatusInternalServerError)
//        fmt.Fprint(w, err.Error())
//        return
//    }

    item := &memcache.Item{
        Key:   "lastTime",
        Value: []byte("0"),
    }
    memcache.Add(c, item)
    lastTimeItem, _ := memcache.Get(c, "lastTime")
    lastTime, _ := strconv.ParseInt(string(lastTimeItem.Value[:]), 10, 64)

    for i,_ := range messages {
        message := messages[i]
        if lastTime != 0 && message.SendTime > lastTime {
            // message := messages[len(messages) - 1]
            body, _ := domain(message, "^画像 (.*)", client, handlerSearchGoogle)
            if body != nil {
                cwClient.PostRoomMessage(roomId, *body)
            }
        }
    }
    lastMessage := messages[len(messages) - 1]
    item.Value = []byte(strconv.FormatInt(lastMessage.SendTime, 10))
    memcache.Set(c, item)
}


func taskHandler(w http.ResponseWriter, r *http.Request) {
    max := 12
    var tasks []*taskqueue.Task
    for i := 0; i < max; i++ {
        tasks = append(tasks, taskqueue.NewPOSTTask("/", nil))
    }
    c := appengine.NewContext(r)
    taskqueue.AddMulti(c, tasks, "read")
}

func handlerSearchGoogle(match []string, message chatwork.Message, client *http.Client) (*string, error) {
    text := match[1]
    n := int(message.SendTime % 10)
    res, err := SearchGoogle(text, n, client)
    var body Body
    json.NewDecoder(res.Body).Decode(&body)
    imageUrl := body.ResponseData.Results[0].UnescapedUrl
    return &imageUrl, err
}

func domain(message chatwork.Message, regexpString string, client *http.Client, handler func(match []string, message chatwork.Message, client *http.Client) (*string, error)) (*string, error){
    if message.Account.AccountID != BOT_ACCOUNT_ID {
        r, _ := regexp.Compile(regexpString)
        match := r.FindStringSubmatch(message.Body)
        if len(match) > 1 {
            return handler(match, message, client)
        }
    }
    return nil, nil
}

func SearchGoogle(text string, number int,  client *http.Client) (*http.Response, error){
    url := "http://ajax.googleapis.com/ajax/services/search/images?v=1.0&rsz=1&imgsz=large&imgtype=face&q="+ url.QueryEscape(text) +"&as_filetype=jpg&start=" + strconv.Itoa(number)
    req, _ := http.NewRequest("GET", url, nil)
    return client.Do(req)
}
