package main

import (
	"github.com/gin-gonic/gin"
	"github.com/haowang1013/wechat-server/wechat"
	"github.com/satori/go.uuid"
	"github.com/skip2/go-qrcode"
	"net/http"
	"net/url"
)

func printUserInfo(c *gin.Context, u *wechat.UserInfo, state string) {
	data := make(map[string]interface{})
	data["user"] = u
	data["state"] = state
	c.IndentedJSON(http.StatusOK, data)
}

func generateQRCode(str string, c *gin.Context, unescape bool) {
	if unescape {
		str, _ = url.QueryUnescape(str)
	}
	data, err := qrcode.Encode(str, qrcode.Medium, 256)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Data(http.StatusOK, "image/png", data)
}

func newUUID() string {
	return uuid.NewV4().String()
}

func makeUrl(scheme, host, path string, queries map[string]string, fragment string) *url.URL {
	u := new(url.URL)
	u.Scheme = scheme
	u.Host = host
	u.Path = path
	u.Fragment = fragment

	if queries != nil {
		params := url.Values{}
		for k, v := range queries {
			params.Add(k, v)
		}
		u.RawQuery = params.Encode()
	}

	return u
}

func makeSimpleUrl(scheme, host, path string) *url.URL {
	return makeUrl(scheme, host, path, nil, "")
}
