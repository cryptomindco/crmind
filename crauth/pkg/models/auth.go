package models

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

type AuthClaims struct {
	jwt.StandardClaims
	Id          int64  `json:"id"`
	Username    string `json:"username"`
	Expire      int64  `json:"expire"`
	Role        int    `json:"role"`
	Createdt    int64  `json:"createdt"`
	LastLogindt int64  `json:"lastLogindt"`
}

type UserInfo struct {
	Id       int64  `json:"id"`
	Username string `json:"username"`
	Role     int    `json:"role"`
}

func (c AuthClaims) Valid() error {
	timestamp := time.Now().Unix()
	if timestamp >= c.Expire {
		return fmt.Errorf("%s", "The credential is expired")
	}
	return nil
}
