package wechat

type LogType int

const (
	Debug LogType = iota
	Notice
	Info
	Warning
	Error
	Fatal
	Panic
)

type Logger interface {
	Log(t LogType, text string)
	Logf(t LogType, format string, v ...interface{})
}
