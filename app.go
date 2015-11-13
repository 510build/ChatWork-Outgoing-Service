package app

import (
    "net/http"
    "appengine"
    "appengine/urlfetch"
    "net/url"
    "encoding/json"
    "regexp"
    chatwork "github.com/eiel/go-chatwork"
)

func init() {
    http.HandleFunc("/", handler)
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
    //for i,_ := range messages {
        // message := messages[i]
        message := messages[len(messages) - 1]
        body, _ := domain(message, "^画像 (.*)", client, handlerSearchGoogle)
        if body != nil {
            cwClient.PostRoomMessage(roomId, *body)
        }
    //}
}


func handlerSearchGoogle(match []string, message chatwork.Message, client *http.Client) (*string, error) {
    text := match[1]
    res, err := SearchGoogle(text, client)
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

func SearchGoogle(text string, client *http.Client) (*http.Response, error){
    url := "http://ajax.googleapis.com/ajax/services/search/images?v=1.0&rsz=1&imgsz=large&imgtype=face&q="+ url.QueryEscape(text) +"&as_filetype=jpg&start=1"
    req, _ := http.NewRequest("GET", url, nil)
    return client.Do(req)
}
