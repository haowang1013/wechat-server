package main

import (
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
	appToken string
	log      = logging.MustGetLogger("")
)

func init() {
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

func loginHandler(rw http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	signature := q.Get("signature")
	timestamp := q.Get("timestamp")
	nonce := q.Get("nonce")
	echostr := q.Get("echostr")
	if wechat.ValidateLogin(timestamp, nonce, appToken, signature) {
		fmt.Fprintf(rw, echostr)
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
		log.Debugf("user message received: %+v", m)
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
		fmt.Fprintf(rw, "")
		return
	}
}

func rootHandler(rw http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		if len(req.URL.Query().Get("signature")) > 0 {
			loginHandler(rw, req)
		} else {
			// regular handler
			fmt.Fprintf(rw, "Hello")
		}
	} else if req.Method == "POST" {
		messageHandler(rw, req)
	}
}

func main() {
	http.HandleFunc("/", rootHandler)
	log.Debugf("listen on port %d", port)
	log.Critical(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
