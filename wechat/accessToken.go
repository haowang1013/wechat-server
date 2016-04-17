package wechat

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	we := new(WechatError)
	err = json.Unmarshal(content, we)
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
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	/**
	on success:
	{
		"access_token":"ACCESS_TOKEN",
		"expires_in":7200
	}
	*/
	token := new(BaseAccessToken)
	err = json.Unmarshal(content, token)
	if err != nil {
		return nil, err
	}

	if len(token.Token) > 0 {
		return token, nil
	}

	/**
	on failure:
	{
		"errcode":40013,
		"errmsg":"invalid appid"
	}
	*/
	we := new(WechatError)
	err = json.Unmarshal(content, we)
	if err != nil {
		return nil, err
	}

	return nil, we
}

func GetWebAccessToken(appID, appSecret, code string) (*WebAccessToken, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code", appID, appSecret, code)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	/**
	on success:
		{
	   		"access_token":"ACCESS_TOKEN",
	   		"expires_in":7200,
	   		"refresh_token":"REFRESH_TOKEN",
	   		"openid":"OPENID",
	   		"scope":"SCOPE",
	   		"unionid": "o6_bmasdasdsad6_2sgVt7hMZOPfL"
	}
	*/
	token := new(WebAccessToken)
	err = json.Unmarshal(content, token)
	if err != nil {
		return nil, err
	}

	if len(token.Token) > 0 {
		return token, nil
	}

	/**
	on failure:
	{
		"errcode":40029,
		"errmsg":"invalid code"
	}
	*/
	we := new(WechatError)
	err = json.Unmarshal(content, we)
	if err != nil {
		return nil, err
	}
	return nil, we
}
