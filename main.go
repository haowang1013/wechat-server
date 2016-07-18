package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/haowang1013/wechat-server/wechat"
	"net/http"
	"os"
	"time"
)

const (
	port        = 8080
	wechatUrl   = "/wechat"
	webLoginUrl = "/wechat/weblogin"
	qrcodeUrl   = "/qrcode/:str"
	loginUrl    = "/login/:uuid"
)

var (
	appID        string
	appSecret    string
	appToken     string
	redisAddress string

	server *wechat.Server

	cache kvCache
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

	redisAddress = os.Getenv("REDIS_SERVER_ADDRESS")
}

func main() {
	if len(redisAddress) == 0 {
		log.Warning("redis server address not configured via environment variable 'REDIS_SERVER_ADDRESS', using in-memory cache")
		cache = newMemCache()
	} else {
		log.Infof("using redis server at: %s", redisAddress)
		cache = newRedisCache(redisAddress, "wechat-login", time.Hour)
	}
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	// create the server and setup router
	server := wechat.NewServer(appID, appSecret, appToken)
	server.SetHandler(new(handler))
	server.SetLogger(new(logger))

	server.SetupRouter(router, wechatUrl)

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

	router.GET("/", func(c *gin.Context) {
		resp := map[string]string{
			"wechat_url":   makeSimpleUrl("http", c.Request.Host, wechatUrl).String(),
			"weblogin_url": makeSimpleUrl("http", c.Request.Host, webLoginUrl).String(),
			"qrcode_url":   makeSimpleUrl("http", c.Request.Host, qrcodeUrl).String(),
			"login_url":    makeSimpleUrl("http", c.Request.Host, loginUrl).String(),
		}
		c.IndentedJSON(http.StatusOK, resp)
	})

	log.Debugf("listen on port %d", port)
	router.Run(fmt.Sprintf(":%d", port))
}
