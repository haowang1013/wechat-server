package main

import (
	"encoding/json"
	"fmt"
	"github.com/haowang1013/wechat-server/wechat"
	"github.com/op/go-logging"
	"github.com/skip2/go-qrcode"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
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

	cache = newCache()
)

type Cache struct {
	data map[string]interface{}
	m    sync.Mutex
}

func newCache() *Cache {
	c := new(Cache)
	c.data = make(map[string]interface{})
	return c
}

func (this *Cache) Get(key string) (interface{}, bool) {
	this.m.Lock()
	defer this.m.Unlock()
	v, ok := this.data[key]
	return v, ok
}

func (this *Cache) Set(key string, value interface{}) {
	this.m.Lock()
	defer this.m.Unlock()
	this.data[key] = value
}

func (this *Cache) Exists(key string) bool {
	this.m.Lock()
	defer this.m.Unlock()
	_, ok := this.data[key]
	return ok
}

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

type logger struct {
}

func (l *logger) Log(t wechat.LogType, text string) {
	switch t {
	case wechat.Debug:
		log.Debug(text)

	case wechat.Notice:
		log.Notice(text)

	case wechat.Info:
		log.Info(text)

	case wechat.Warning:
		log.Warning(text)

	case wechat.Error:
		log.Error(text)

	case wechat.Fatal:
		log.Fatal(text)

	case wechat.Panic:
		log.Panic(text)

	default:
		panic("log type not supported")
	}
}

func (l *logger) Logf(t wechat.LogType, format string, v ...interface{}) {
	switch t {
	case wechat.Debug:
		log.Debugf(format, v...)

	case wechat.Notice:
		log.Noticef(format, v...)

	case wechat.Info:
		log.Infof(format, v...)

	case wechat.Warning:
		log.Warningf(format, v...)

	case wechat.Error:
		log.Errorf(format, v...)

	case wechat.Fatal:
		log.Fatalf(format, v...)

	case wechat.Panic:
		log.Panicf(format, v...)

	default:
		panic("log type not supported")
	}
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

func printUserInfo(result io.Writer, u *wechat.UserInfo, state string) {
	data := make(map[string]interface{})
	data["user"] = u
	data["state"] = state
	content, _ := json.MarshalIndent(data, "", "  ")
	fmt.Fprint(result, string(content))
}

func (h *handler) HandleWebLogin(u *wechat.UserInfo, state string, result io.Writer) {
	log.Debugf("%+v logged in with state '%s'", u, state)
	cache.Set(state, u)
	printUserInfo(result, u, state)
}

func generateQRCode(str string, rw http.ResponseWriter) {
	data, err := qrcode.Encode(str, qrcode.Medium, 256)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rw.Header().Add("Content-Length", strconv.Itoa(len(data)))
	rw.Header().Add("Content-Type", "image/png")
	rw.Write(data)
}

func qrHandler(rw http.ResponseWriter, req *http.Request) {
	uri := req.RequestURI
	str := uri[len("/qrcode/"):]
	generateQRCode(str, rw)
}

func loginTestHandler(rw http.ResponseWriter, req *http.Request) {
	uri := req.RequestURI
	state := uri[len("/logintest/"):]
	redirectUrl := "http://wechattest.ngrok.natapp.cn/wechat/weblogin"
	loginUrl := fmt.Sprintf("https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_userinfo&state=%s#wechat_redirect", appID, url.QueryEscape(redirectUrl), state)
	generateQRCode(loginUrl, rw)
}

func testLoginHandler(rw http.ResponseWriter, req *http.Request) {
	uri := req.RequestURI
	state := uri[len("/logintest/"):]
	o, ok := cache.Get(state)
	if !ok {
		fmt.Fprintf(rw, "No user logged in with state '%s'", state)
		return
	}

	u, ok := o.(*wechat.UserInfo)
	if !ok {
		http.Error(rw, "Invalid user info", http.StatusInternalServerError)
		return
	}

	printUserInfo(rw, u, state)
}

func main() {
	server := wechat.NewServer(appID, appSecret, appToken)
	server.SetHandler(new(handler))
	server.SetLogger(new(logger))

	http.HandleFunc("/wechat", func(rw http.ResponseWriter, req *http.Request) {
		server.RouteRequest(rw, req)
	})

	http.HandleFunc("/wechat/weblogin", func(rw http.ResponseWriter, req *http.Request) {
		server.RouteWebLogin(rw, req)
	})

	http.HandleFunc("/qrcode/", qrHandler)
	http.HandleFunc("/logintest/", loginTestHandler)

	http.HandleFunc("/testlogin/", testLoginHandler)

	log.Debugf("listen on port %d", port)
	log.Critical(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
