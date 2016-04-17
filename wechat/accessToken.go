package wechat

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func GetAccessToken(appID, appSecret string) (string, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", appID, appSecret)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	decoded := make(map[string]interface{})

	err = json.Unmarshal(content, &decoded)
	if err != nil {
		return "", err
	}

	// response example: {"access_token":"ACCESS_TOKEN","expires_in":7200}
	token := decoded["access_token"]
	if token != nil {
		return token.(string), nil
	}

	// response example: {"errcode":40013,"errmsg":"invalid appid"}
	code := decoded["errcode"].(int)
	msg := decoded["errmsg"].(string)
	return "", NewError(code, msg)
}
