package models

type AuthClaims struct {
	Id           int64  `json:"id"`
	Username     string `json:"username"`
	Expire       int64  `json:"expire"`
	Role         int    `json:"role"`
	Token        string `json:"token"`
	Contacts     string `json:"contacts"`
	CredsArrJson string `json:"credsArrJson"`
}
