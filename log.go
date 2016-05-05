package main

import (
	"github.com/haowang1013/wechat-server/wechat"
	"github.com/op/go-logging"
	"os"
)

var (
	log = logging.MustGetLogger("")
)

func init() {
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
