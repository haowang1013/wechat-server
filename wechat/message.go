package wechat

import (
	"encoding/xml"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	textResponseTemplate = "<xml><ToUserName><![CDATA[%s]]></ToUserName><FromUserName><![CDATA[%s]]></FromUserName><CreateTime>%d</CreateTime><MsgType><![CDATA[text]]></MsgType><Content><![CDATA[%s]]></Content></xml>"
)

var (
	messageFactory = make(map[string]func() UserMessage)
)

func init() {
	messageFactory["text"] = func() UserMessage {
		return new(UserTextMessage)
	}

	messageFactory["image"] = func() UserMessage {
		return new(UserImageMessage)
	}

	messageFactory["voice"] = func() UserMessage {
		return new(UserVoiceMessage)
	}

	messageFactory["video"] = func() UserMessage {
		return new(UserVideoMessage)
	}

	messageFactory["shortvideo"] = func() UserMessage {
		return new(UserVideoMessage)
	}

	messageFactory["link"] = func() UserMessage {
		return new(UserLinkMessage)
	}

	messageFactory["event"] = func() UserMessage {
		return new(BaseEvent)
	}
}

type UserMessage interface {
	MessageType() string
	To() string
	From() string
	ReplyText(c *gin.Context, content string)
}

type UserEvent interface {
	UserMessage
	EventType() string
}

type BaseMessage struct {
	ToUserName   string
	FromUserName string
	CreateTime   int64
	MsgType      string
	MsgId        int64
}

func (this *BaseMessage) MessageType() string {
	return this.MsgType
}

func (this *BaseMessage) To() string {
	return this.ToUserName
}

func (this *BaseMessage) From() string {
	return this.FromUserName
}

func (this *BaseMessage) ReplyText(c *gin.Context, content string) {
	text := fmt.Sprintf(textResponseTemplate, this.FromUserName, this.ToUserName, 0, content)
	c.String(http.StatusOK, text)
}

type UserTextMessage struct {
	BaseMessage
	Content string
}

type UserImageMessage struct {
	BaseMessage
	PicUrl  string
	MediaId string
}

type UserVoiceMessage struct {
	BaseMessage
	MediaId string
	Format  string
}

type UserVideoMessage struct {
	BaseMessage
	MediaID      string
	ThumbMediaId string
}

type UserLinkMessage struct {
	BaseMessage
	Title       string
	Description string
	Url         string
}

type BaseEvent struct {
	BaseMessage
	Event string
}

func (this *BaseEvent) EventType() string {
	return this.Event
}

func LoadUserMessage(content []byte) (UserMessage, error) {
	var base BaseMessage
	err := xml.Unmarshal(content, &base)
	if err != nil {
		return nil, err
	}

	factory := messageFactory[base.MsgType]
	if factory == nil {
		return nil, fmt.Errorf("Unknown message type: %s", base.MsgType)
	}

	m := factory()
	err = xml.Unmarshal(content, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}
