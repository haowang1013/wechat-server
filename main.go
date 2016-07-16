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
	qrcodeUrl   = "/qrcode/:str"
	loginUrl    = "/login/:uuid"
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
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	// create the server and setup router
	server := wechat.NewServer(appID, appSecret, appToken)
	server.SetHandler(new(handler))
	server.SetLogger(new(logger))

	server.SetupRouter(router, "/wechat")

	// web login endpoint
	router.GET(webLoginUrl, func(c *gin.Context) {
		server.HandleWebLogin(c)
	})

	// qr code endpoint
	router.GET(qrcodeUrl, func(c *gin.Context) {
		str := c.Param("str")
		unescape := c.DefaultQuery("unescape", "false")
		generateQRCode(str, c, unescape == "true")
	})

	// client facing login endpoint
	router.POST("/login", func(c *gin.Context) {
		loginRequestHandler(c)
	})

	router.GET(loginUrl, func(c *gin.Context) {
		uuid := c.Param("uuid")
		loginQueryHandler(uuid, c)
	})

	log.Debugf("listen on port %d", port)
	router.Run(fmt.Sprintf(":%d", port))
}
