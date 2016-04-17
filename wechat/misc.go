package wechat

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

type WeChatError struct {
	Code    int    `json:"errcode"`
	Message string `json:"errmsg"`
}

func (this *WeChatError) Error() string {
	return fmt.Sprintf("Code: %d, Message: %s", this.Code, this.Message)
}

func NewError(code int, message string) *WeChatError {
	err := new(WeChatError)
	err.Code = code
	err.Message = message
	return err
}

func ValidateLogin(timestamp, nonce, appToken, signature string) bool {
	a := []string{
		timestamp,
		nonce,
		appToken,
	}
	sort.Strings(a)
	combined := strings.Join(a, "")

	hash := sha1.Sum([]byte(combined))
	h := hex.EncodeToString(hash[:])
	return h == signature
}
