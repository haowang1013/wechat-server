package main

import (
	"encoding/json"
	"fmt"
	"github.com/haowang1013/wechat-server/wechat"
	"github.com/skip2/go-qrcode"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

var (
	cache = newCache()
)

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
	cache.set(state, u)
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
	str := uri[len(qrCodeUrl):]
	generateQRCode(str, rw)
}

func loginTestHandler(rw http.ResponseWriter, req *http.Request) {
	uri := req.RequestURI
	state := uri[len(loginTestUrl):]
	redirectUrl := "http://wechattest.ngrok.natapp.cn/wechat/weblogin"
	loginUrl := fmt.Sprintf("https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_userinfo&state=%s#wechat_redirect", appID, url.QueryEscape(redirectUrl), state)
	generateQRCode(loginUrl, rw)
}

func testLoginHandler(rw http.ResponseWriter, req *http.Request) {
	uri := req.RequestURI
	state := uri[len(testLoginUrl):]
	o, ok := cache.get(state)
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
