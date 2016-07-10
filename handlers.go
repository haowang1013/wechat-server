package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/haowang1013/wechat-server/wechat"
	"github.com/skip2/go-qrcode"
	"net/http"
	"net/url"
)

var (
	cache = newCache()
)

type handler struct {
}

func (h *handler) HandleText(m *wechat.UserTextMessage, c *gin.Context) {
	m.ReplyText(c, fmt.Sprintf("You said '%s'", m.Content))
}

func (h *handler) HandleImage(m *wechat.UserImageMessage, c *gin.Context) {
	m.ReplyText(c, fmt.Sprintf("Image uploaded to %s", m.PicUrl))
}

func (h *handler) HandleVoice(m *wechat.UserVoiceMessage, c *gin.Context) {
	m.ReplyText(c, "Thank you for sending a voice message")
}

func (h *handler) HandleVideo(m *wechat.UserVideoMessage, c *gin.Context) {
	m.ReplyText(c, "Thank you for sending a video message")
}

func (h *handler) HandleLink(m *wechat.UserLinkMessage, c *gin.Context) {
	m.ReplyText(c, "Thank you for sending a link message")
}

func (h *handler) HandleEvent(event wechat.UserEvent, c *gin.Context) {
	et := event.EventType()
	switch et {
	case "subscribe":
		log.Debugf("new follower: %s", event.From())
		event.ReplyText(c, "Welcome!")
	case "unsubscribe":
		log.Debugf("%s unsubscribed", event.From())
		c.String(http.StatusOK, "")
	default:
		log.Errorf("unknown event type: %s", et)
		c.String(http.StatusOK, "")
	}
}

func printUserInfo(c *gin.Context, u *wechat.UserInfo, state string) {
	data := make(map[string]interface{})
	data["user"] = u
	data["state"] = state
	c.IndentedJSON(http.StatusOK, data)
}

func (h *handler) HandleWebLogin(u *wechat.UserInfo, state string, c *gin.Context) {
	log.Debugf("%+v logged in with state '%s'", u, state)
	cache.set(state, u)
	printUserInfo(c, u, state)
}

func generateQRCode(str string, c *gin.Context) {
	data, err := qrcode.Encode(str, qrcode.Medium, 256)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Data(http.StatusOK, "image/png", data)
}

func loginTestHandler(state string, c *gin.Context) {
	redirectUrl := makeUrl("http", c.Request.Host, webLoginUrl, nil, "").String()

	loginUrl := makeUrl("https", "open.weixin.qq.com", "/connect/oauth2/authorize", map[string]string{
		"appid":         appID,
		"redirect_uri":  redirectUrl,
		"response_type": "code",
		"scope":         "snsapi_userinfo",
		"state":         state,
	},
		"wechat_redirect").String()
	generateQRCode(loginUrl, c)
}

func testLoginHandler(state string, c *gin.Context) {
	o, ok := cache.get(state)
	if !ok {
		c.String(http.StatusNotFound, "No user logged in with state '%s'", state)
		return
	}

	u, ok := o.(*wechat.UserInfo)
	if !ok {
		c.String(http.StatusInternalServerError, "Invalid user info")
		return
	}

	printUserInfo(c, u, state)
}

func makeUrl(scheme, host, path string, queries map[string]string, fragment string) *url.URL {
	u := new(url.URL)
	u.Scheme = scheme
	u.Host = host
	u.Path = path
	u.Fragment = fragment

	if queries != nil {
		params := url.Values{}
		for k, v := range queries {
			params.Add(k, v)
		}
		u.RawQuery = params.Encode()
	}

	return u
}
