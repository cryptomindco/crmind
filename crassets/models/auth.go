package models

type AuthClaims struct {
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

type ContactItem struct {
	UserId   int64  `json:"userId"`
	UserName string `json:"userName"`
	Addeddt  int64  `json:"addeddt"`
}
