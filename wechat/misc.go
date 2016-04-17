package wechat

import (
	"crypto/sha1"
	"encoding/hex"
	"sort"
	"strings"
)

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
