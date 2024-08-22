package models

type AuthClaims struct {
	Id          int64  `json:"id"`
	Username    string `json:"username"`
	Expire      int64  `json:"expire"`
	Role        int    `json:"role"`
	Createdt    int64  `json:"createdt"`
	LastLogindt int64  `json:"lastLogindt"`
}

type User struct {
	Id           int64  `orm:"column(id);auto;size(11)" json:"id"`
	Username     string `json:"username"`
	Name         string `json:"name"`
	Role         int    `json:"role"`
	Status       int    `json:"status"`
	Createdt     int64  `orm:"size(10);default(0)" json:"createdt"`
	Updatedt     int64  `orm:"size(10);default(0)" json:"updatedt"`
	LastLogindt  int64  `orm:"size(10);default(0)" json:"lastLogindt"`
	CredsArrJson string `json:"credsArrJson"`
}

type UserInfo struct {
	Id       int64  `json:"id"`
	UserName string `json:"username"`
	Role     int    `json:"role"`
}
