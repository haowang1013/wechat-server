package main

import (
	"encoding/json"
	"fmt"
	"github.com/haowang1013/wechat-server/wechat"
	"github.com/op/go-logging"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	port = 8080
)

var (
	appID     string
	appSecret string
	appToken  string

	accessToken *wechat.BaseAccessToken

	log = logging.MustGetLogger("")
)

func init() {
	appID = os.Getenv("WECHAT_APP_ID")
	if len(appID) == 0 {
		panic("Failed to get app id from env variable 'WECHAT_APP_ID'")
	}

	appSecret = os.Getenv("WECHAT_APP_SECRET")
	if len(appSecret) == 0 {
		panic("Failed to get app secret from env variable 'WECHAT_APP_SECRET'")
	}

	appToken = os.Getenv("WECHAT_APP_TOKEN")
	if len(appToken) == 0 {
		panic("Failed to get app token from env variable 'WECHAT_APP_TOKEN'")
	}

	format := logging.MustStringFormatter(
		"%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}",
	)

	backend := logging.NewLogBackend(os.Stdout, "", 0)
	formtter := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(formtter)
}

func logUserInfo(openID string) {
	if accessToken == nil {
		return
	}

	user, err := wechat.GetUserInfo(accessToken, openID)
	if err != nil {
		log.Errorf("failed to get user info: %s", err.Error())
		return
	}
	log.Debugf("%+v", user)
}

func loginHandler(rw http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	signature := q.Get("signature")
	timestamp := q.Get("timestamp")
	nonce := q.Get("nonce")
	echostr := q.Get("echostr")
	if wechat.ValidateLogin(timestamp, nonce, appToken, signature) {
		fmt.Fprint(rw, echostr)
		log.Debug("validated wechat login request")
	} else {
		http.Error(rw, "Signature doesn't match", http.StatusBadRequest)
		log.Error("failed to validate wechat login request")
	}
}

func messageHandler(rw http.ResponseWriter, req *http.Request) {
	content, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	m, err := wechat.LoadUserMessage(content)
	if err == nil {
		log.Debugf("message received: %+v", m)

		go logUserInfo(m.From())

		if event, ok := m.(wechat.UserEvent); ok {
			eventHandler(rw, req, event)
			return
		}

		switch v := m.(type) {
		case *wechat.UserTextMessage:
			v.ReplyText(rw, fmt.Sprintf("You said '%s'", v.Content))
		case *wechat.UserImageMessage:
			v.ReplyText(rw, fmt.Sprintf("Image uploaded to %s", v.PicUrl))
		case *wechat.UserVoiceMessage:
			v.ReplyText(rw, "Thank you for uploading voice")
		case *wechat.UserVideoMessage:
			v.ReplyText(rw, "Thank you for uploading video")
		case *wechat.UserLinkMessage:
			v.ReplyText(rw, "Thank you for sending a link")
		}
	} else {
		log.Errorf("failed to load user message: %s", err.Error())
		fmt.Fprint(rw, "")
		return
	}
}

func eventHandler(rw http.ResponseWriter, req *http.Request, event wechat.UserEvent) {
	et := event.EventType()
	switch et {
	case "subscribe":
		log.Debugf("new follower: %s", event.From())
		event.ReplyText(rw, "Welcome!")
	case "unsubscribe":
		log.Debugf("%s unsubscribed", event.From())
		fmt.Fprint(rw, "")
	default:
		log.Errorf("unknown event type: %s", et)
		fmt.Fprint(rw, "")
	}
}

func rootHandler(rw http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		if len(req.URL.Query().Get("signature")) > 0 {
			loginHandler(rw, req)
		} else {
			// regular handler
			fmt.Fprint(rw, "Hello")
		}
	} else if req.Method == "POST" {
		messageHandler(rw, req)
	}
}

func webLoginHandler(rw http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	code := q.Get("code")
	state := q.Get("state")
	log.Debugf("web login: code=%s, state=%s", code, state)

	if len(code) == 0 {
		fmt.Fprint(rw, "")
		return
	}

	token, err := wechat.GetWebAccessToken(appID, appSecret, code)
	if err != nil {
		log.Errorf("failed to get web access token: %s", err.Error())
		fmt.Fprint(rw, "")
		return
	}

	user, err := wechat.GetUserInfoWithWebToken(token)
	if err != nil {
		log.Errorf("failed to user info with web access token: %s", err.Error())
		fmt.Fprint(rw, "")
		return
	}

	content, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		log.Error(err)
		fmt.Fprint(rw, "")
		return
	}

	fmt.Fprint(rw, string(content))
}

func main() {
	t, err := wechat.GetAccessToken(appID, appSecret)
	if err == nil {
		accessToken = t
		log.Debugf("access token acquired: %+v", accessToken)
	} else {
		log.Errorf("failed to get access token: %s", err.Error())
	}
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/weblogin", webLoginHandler)
	log.Debugf("listen on port %d", port)
	log.Critical(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
