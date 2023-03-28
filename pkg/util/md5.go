package util

import (
	"GoChat/config"
	"crypto/md5"
	"fmt"
)

// GetMD5 加盐生成 md5
func GetMD5(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s+config.GlobalConfig.App.Salt)))
}
