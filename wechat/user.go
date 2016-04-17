package wechat

import (
	"fmt"
	"github.com/levigross/grequests"
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

func GetUserInfo(token *BaseAccessToken, openID string) (*UserInfo, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/user/info?access_token=%s&openid=%s&lang=zh_CN", token.Token, openID)
	resp, err := grequests.Get(url, nil)
	if err != nil {
		return nil, err
	}

	user := new(UserInfo)
	err = resp.JSON(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func GetUserInfoWithWebToken(token *WebAccessToken) (*UserInfo, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s&lang=zh_CN", token.Token, token.OpenID)
	resp, err := grequests.Get(url, nil)
	if err != nil {
		return nil, err
	}

	user := new(UserInfo)
	err = resp.JSON(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}
