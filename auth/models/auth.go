package models

import (
	"fmt"
	"time"
)

type AuthClaims struct {
	Id       int64  `json:"id"`
	Username string `json:"username"`
	Expire   int64  `json:"expire"`
	Role     int    `json:"role"`
	Token    string `json:"token"`
	Contacts string `json:"contacts"`
	Createdt int64  `json:"createdt"`
}

type UserInfo struct {
	Id       int64  `json:"id"`
	Username string `json:"username"`
	Token    string `json:"token"`
	Contacts string `json:"contacts"`
}

func (c AuthClaims) Valid() error {
	timestamp := time.Now().Unix()
	if timestamp >= c.Expire {
		return fmt.Errorf("%s", "The credential is expired")
	}
	return nil
}
