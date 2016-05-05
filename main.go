package main

import (
	"fmt"
	"github.com/haowang1013/wechat-server/wechat"
	"net/http"
	"os"
)

const (
	port         = 8080
	loginTestUrl = "/logintest/"
	testLoginUrl = "/testlogin/"
	qrCodeUrl    = "/qrcode/"
)

var (
	appID     string
	appSecret string
	appToken  string

	server *wechat.Server
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

	http.HandleFunc(qrCodeUrl, qrHandler)

	http.HandleFunc(loginTestUrl, loginTestHandler)
	http.HandleFunc(testLoginUrl, testLoginHandler)

	log.Debugf("listen on port %d", port)
	log.Critical(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
