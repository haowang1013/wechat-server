package wechat

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type UserInfo struct {
	Subscribed    int    `json:"subscribe"`
	OpenID        string `json:"openid"`
	NickName      string `json:"nickname"`
	Gender        int    `json:"sex"`
	City          string `json:"city"`
	Country       string `json:"country"`
	Province      string `json:"province"`
	Language      string `json:"language"`
	IconUrl       string `json:"headimgurl"`
	SubscribeTime int    `json:"subscribe_time"`
	UnionID       string `json:"unionid"`
	Remark        string `json:"remark"`
	GroupID       int    `json:"groupid"`
}

func GetUserInfo(accessToken, openID string) (*UserInfo, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/user/info?access_token=%s&openid=%s&lang=zh_CN", accessToken, openID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	user := new(UserInfo)
	err = json.Unmarshal(content, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}
