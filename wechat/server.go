package wechat

import (
	"fmt"
	"io"
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
	HandleText(m *UserTextMessage, result io.Writer)
	HandleImage(m *UserImageMessage, result io.Writer)
	HandleVoice(m *UserVoiceMessage, result io.Writer)
	HandleVideo(m *UserVideoMessage, result io.Writer)
	HandleLink(m *UserLinkMessage, result io.Writer)
	HandleEvent(e UserEvent, result io.Writer)
	HandleWebLogin(u *UserInfo, state string, result io.Writer)
}

func (s *Server) SetHandler(h ServerHandler) {
	s.handler = h
}

func (s *Server) SetLogger(logger Logger) {
	s.logger = logger
}

func (s *Server) RouteRequest(rw http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		if len(req.URL.Query().Get("signature")) > 0 {
			s.handleLogin(rw, req)
		} else {
			fmt.Fprint(rw, "Hello World")
		}
	} else if req.Method == "POST" {
		s.handleMessage(rw, req)
	}
}

func (s *Server) RouteWebLogin(rw http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	code := q.Get("code")
	state := q.Get("state")
	s.logf(Debug, "web login: code=%s, state=%s", code, state)

	if len(code) == 0 {
		fmt.Fprint(rw, "")
		return
	}

	token, err := GetWebAccessToken(s.appID, s.appSecret, code)
	if err != nil {
		s.logf(Error, "failed to get web access token: %s", err.Error())
		fmt.Fprint(rw, "")
		return
	}

	user, err := GetUserInfoWithWebToken(token)
	if err != nil {
		s.logf(Error, "failed to user info with web access token: %s", err.Error())
		fmt.Fprint(rw, "")
		return
	}

	if s.handler != nil {
		s.handler.HandleWebLogin(user, state, rw)
	} else {
		fmt.Fprint(rw, "")
		return
	}
}

func (s *Server) handleLogin(rw http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	signature := q.Get("signature")
	timestamp := q.Get("timestamp")
	nonce := q.Get("nonce")
	echostr := q.Get("echostr")
	if ValidateLogin(timestamp, nonce, s.token, signature) {
		fmt.Fprint(rw, echostr)
		s.log(Debug, "validated wechat login request")
	} else {
		http.Error(rw, "Signature doesn't match", http.StatusBadRequest)
		s.log(Error, "failed to validate wechat login request")
	}
}

func (s *Server) handleMessage(rw http.ResponseWriter, req *http.Request) {
	content, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if s.handler == nil {
		fmt.Fprint(rw, "")
		return
	}

	m, err := LoadUserMessage(content)
	if err == nil {
		s.logf(Debug, "message received: %+v", m)
		if event, ok := m.(UserEvent); ok {
			s.handler.HandleEvent(event, rw)
			return
		}

		switch v := m.(type) {
		case *UserTextMessage:
			s.handler.HandleText(v, rw)

		case *UserImageMessage:
			s.handler.HandleImage(v, rw)

		case *UserVoiceMessage:
			s.handler.HandleVoice(v, rw)

		case *UserVideoMessage:
			s.handler.HandleVideo(v, rw)

		case *UserLinkMessage:
			s.handler.HandleLink(v, rw)
		}
	} else {
		s.logf(Error, "failed to load user message: %s", err.Error())
		fmt.Fprint(rw, "")
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
