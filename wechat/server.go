package wechat

import (
	"errors"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

type Server struct {
	appID     string
	appSecret string
	token     string
	handler   ServerHandler
	logger    Logger
}

type ServerHandler interface {
	HandleText(m *UserTextMessage, c *gin.Context)
	HandleImage(m *UserImageMessage, c *gin.Context)
	HandleVoice(m *UserVoiceMessage, c *gin.Context)
	HandleVideo(m *UserVideoMessage, c *gin.Context)
	HandleLink(m *UserLinkMessage, c *gin.Context)
	HandleEvent(e UserEvent, c *gin.Context)
	HandleWebLogin(u *UserInfo, state string, c *gin.Context)
}

func (s *Server) SetHandler(h ServerHandler) {
	s.handler = h
}

func (s *Server) SetLogger(logger Logger) {
	s.logger = logger
}

func (s *Server) SetupRouter(router *gin.Engine, url string) {
	router.GET(url, func(c *gin.Context) {
		signature := c.Query("signature")
		if signature != "" {
			timestamp := c.Query("timestamp")
			nonce := c.Query("nonce")
			echostr := c.Query("echostr")
			if ValidateLogin(timestamp, nonce, s.token, signature) {
				s.log(Debug, "validated wechat login request")
				c.String(http.StatusOK, echostr)
			} else {
				s.log(Error, "failed to validate wechat login request")
				c.AbortWithError(http.StatusBadRequest, errors.New("Signature doesn't match"))
				return
			}
		} else {
			c.String(http.StatusOK, "Hello World")
		}

	})

	router.POST(url, func(c *gin.Context) {
		s.handleMessage(c)
	})

}

func (s *Server) HandleWebLogin(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	s.logf(Debug, "handling web login, code=%s, state=%s", code, state)

	token, err := GetWebAccessToken(s.appID, s.appSecret, code)
	if err != nil {
		s.logf(Error, "failed to get web access token with code '%s': %s", code, err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	user, err := GetUserInfoWithWebToken(token)
	if err != nil {
		s.logf(Error, "failed to user info with web access token: %s", err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if s.handler != nil {
		s.handler.HandleWebLogin(user, state, c)
	} else {
		c.String(http.StatusOK, "login succeed")
		return
	}
}

func (s *Server) handleMessage(c *gin.Context) {
	content, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if s.handler == nil {
		c.String(http.StatusOK, "")
		return
	}

	m, err := LoadUserMessage(content)
	if err == nil {
		s.logf(Debug, "message received: %+v", m)
		if event, ok := m.(UserEvent); ok {
			s.handler.HandleEvent(event, c)
			return
		}

		switch v := m.(type) {
		case *UserTextMessage:
			s.handler.HandleText(v, c)

		case *UserImageMessage:
			s.handler.HandleImage(v, c)

		case *UserVoiceMessage:
			s.handler.HandleVoice(v, c)

		case *UserVideoMessage:
			s.handler.HandleVideo(v, c)

		case *UserLinkMessage:
			s.handler.HandleLink(v, c)
		}
	} else {
		s.logf(Error, "failed to load user message: %s", err.Error())
		c.String(http.StatusOK, "")
		return
	}
}

func (s *Server) log(t LogType, text string) {
	if s.logger != nil {
		s.logger.Log(t, text)
	}
}

func (s *Server) logf(t LogType, format string, v ...interface{}) {
	if s.logger != nil {
		s.logger.Logf(t, format, v...)
	}
}

func NewServer(appID, appSecret, token string) *Server {
	s := new(Server)
	s.appID = appID
	s.appSecret = appSecret
	s.token = token
	return s
}
