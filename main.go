package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/haowang1013/wechat-server/wechat"
	"os"
)

const (
	port        = 8080
	webLoginUrl = "/wechat/weblogin"
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
	router := gin.Default()

	server := wechat.NewServer(appID, appSecret, appToken)
	server.SetHandler(new(handler))
	server.SetLogger(new(logger))

	server.SetupRouter(router, "/wechat")

	router.GET(webLoginUrl, func(c *gin.Context) {
		server.RouteWebLogin(c)
	})

	router.GET("/qrcode/:str", func(c *gin.Context) {
		str := c.Param("str")
		generateQRCode(str, c)
	})

	router.GET("/logintest/:state", func(c *gin.Context) {
		state := c.Param("state")
		loginTestHandler(state, c)
	})

	router.GET("/testlogin/:state", func(c *gin.Context) {
		state := c.Param("state")
		testLoginHandler(state, c)
	})

	log.Debugf("listen on port %d", port)
	router.Run(fmt.Sprintf(":%d", port))
}
