package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/haowang1013/wechat-server/wechat"
	"net/http"
	"net/url"
	"strings"
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

func (h *handler) HandleWebLogin(u *wechat.UserInfo, uuid string, c *gin.Context) {
	log.Debugf("%+v logged in with uuid '%s'", u, uuid)
	record, ok := cache.get(uuid)
	if !ok {
		log.Errorf("invalid uuid from web login: '%s'", uuid)
		c.String(http.StatusBadRequest, "Invalid UUID")
		return
	}

	existing, _ := record.(*wechat.UserInfo)
	if existing != nil && existing.OpenID != u.OpenID {
		log.Errorf("user '%s' has logged in with uuid '%s', current user '%s' is rejected", existing.OpenID, uuid, u.OpenID)
		c.String(http.StatusBadRequest, "UUID expired")
		return
	}

	cache.set(uuid, u)
	c.HTML(http.StatusOK, "wechat_welcome.html", gin.H{
		"title": "Main website",
	})
}

func loginRequestHandler(c *gin.Context) {
	uid := newUUID()
	for ; cache.exists(uid); uid = newUUID() {
	}
	cache.set(uid, nil)

	redirectUrl := makeUrl(
		"http",
		c.Request.Host,
		webLoginUrl,
		nil,
		"").String()

	wechatUrl := makeUrl(
		"https",
		"open.weixin.qq.com",
		"/connect/oauth2/authorize",
		map[string]string{
			"appid":         appID,
			"redirect_uri":  redirectUrl,
			"response_type": "code",
			"scope":         "snsapi_userinfo",
			"state":         uid,
		},
		"wechat_redirect").String()

	qrUrl := makeUrl(
		"http",
		c.Request.Host,
		strings.Replace(qrcodeUrl, ":str", url.QueryEscape(wechatUrl), 1),
		map[string]string{
			"unescape": "true",
		},
		"").String()

	queryUrl := makeUrl(
		"http",
		c.Request.Host,
		strings.Replace(loginUrl, ":uuid", uid, 1),
		nil,
		"").String()

	resp := map[string]string{
		"uuid":       uid,
		"query_url":  queryUrl,
		"qrcode_url": qrUrl,
	}

	c.IndentedJSON(http.StatusCreated, resp)
}

func loginQueryHandler(uuid string, c *gin.Context) {
	o, ok := cache.get(uuid)
	if !ok {
		c.String(http.StatusNotFound, "uuid not found")
		return
	}

	if o == nil {
		c.String(http.StatusNotFound, "uuid not logged in")
		return
	}

	u, ok := o.(*wechat.UserInfo)
	if !ok {
		c.String(http.StatusInternalServerError, "Invalid user info")
		return
	}

	data := make(map[string]interface{})
	data["user"] = u
	data["uuid"] = uuid
	c.IndentedJSON(http.StatusOK, data)
}
