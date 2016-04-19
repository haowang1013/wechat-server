package main

import (
	"encoding/json"
	"fmt"
	"github.com/haowang1013/wechat-server/wechat"
	"github.com/op/go-logging"
	"io"
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

	server *wechat.Server

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

type handler struct {
}

func (h *handler) HandleText(m *wechat.UserTextMessage, result io.Writer) {
	m.ReplyText(result, fmt.Sprintf("You said '%s'", m.Content))
}

func (h *handler) HandleImage(m *wechat.UserImageMessage, result io.Writer) {
	m.ReplyText(result, fmt.Sprintf("Image uploaded to %s", m.PicUrl))
}

func (h *handler) HandleVoice(m *wechat.UserVoiceMessage, result io.Writer) {
	m.ReplyText(result, "Thank you for sending a voice message")
}

func (h *handler) HandleVideo(m *wechat.UserVideoMessage, result io.Writer) {
	m.ReplyText(result, "Thank you for sending a video message")
}

func (h *handler) HandleLink(m *wechat.UserLinkMessage, result io.Writer) {
	m.ReplyText(result, "Thank you for sending a link message")
}

func (h *handler) HandleEvent(event wechat.UserEvent, result io.Writer) {
	et := event.EventType()
	switch et {
	case "subscribe":
		log.Debugf("new follower: %s", event.From())
		event.ReplyText(result, "Welcome!")
	case "unsubscribe":
		log.Debugf("%s unsubscribed", event.From())
		fmt.Fprint(result, "")
	default:
		log.Errorf("unknown event type: %s", et)
		fmt.Fprint(result, "")
	}
}

func (h *handler) HandleWebLogin(u *wechat.UserInfo, state string, result io.Writer) {
	data := make(map[string]interface{})
	data["user"] = u
	data["state"] = state
	content, _ := json.MarshalIndent(data, "", "  ")
	fmt.Fprint(result, string(content))
}

func main() {
	server := wechat.NewServer(appID, appSecret, appToken)
	server.SetHandler(new(handler))

	http.HandleFunc("/wechat", func(rw http.ResponseWriter, req *http.Request) {
		server.RouteRequest(rw, req)
	})

	http.HandleFunc("/wechat/weblogin", func(rw http.ResponseWriter, req *http.Request) {
		server.RouteWebLogin(rw, req)
	})

	log.Debugf("listen on port %d", port)
	log.Critical(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
