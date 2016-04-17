package wechat

import (
	"encoding/json"
	"fmt"
	"github.com/levigross/grequests"
)

type BaseAccessToken struct {
	Token     string `json:"access_token"`
	ExpiresIn int    `json:"expires_in"`
}

type WebAccessToken struct {
	BaseAccessToken
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid"`
}

func (this *WebAccessToken) Validate() error {
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/auth?access_token=%s&openid=%s", this.Token, this.OpenID)
	resp, err := grequests.Get(url, nil)
	if err != nil {
		return err
	}

	we := new(WeChatError)
	err = resp.JSON(we)

	if err != nil {
		return err
	}

	if we.Code == 0 {
		return nil
	}

	return we
}

func GetAccessToken(appID, appSecret string) (*BaseAccessToken, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", appID, appSecret)
	resp, err := grequests.Get(url, nil)
	if err != nil {
		return nil, err
	}

	b := resp.Bytes()
	token := new(BaseAccessToken)
	err = json.Unmarshal(b, token)
	if err != nil {
		return nil, err
	}

	if len(token.Token) > 0 {
		return token, nil
	}

	we := new(WeChatError)
	err = json.Unmarshal(b, we)
	if err != nil {
		return nil, err
	}

	return nil, we
}

func GetWebAccessToken(appID, appSecret, code string) (*WebAccessToken, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code", appID, appSecret, code)
	resp, err := grequests.Get(url, nil)
	if err != nil {
		return nil, err
	}

	b := resp.Bytes()
	token := new(WebAccessToken)
	err = json.Unmarshal(b, token)
	if err != nil {
		return nil, err
	}

	if len(token.Token) > 0 {
		return token, nil
	}

	we := new(WeChatError)
	err = json.Unmarshal(b, we)
	if err != nil {
		return nil, err
	}
	return nil, we
}
