package models

type User struct {
	Id           int64  `orm:"column(id);auto;size(11)" json:"id"`
	Username     string `json:"username"`
	Name         string `json:"name"`
	Password     string `json:"password"`
	Role         int    `json:"role"`
	Status       int    `json:"status"`
	Token        string `json:"token"`
	Contacts     string `json:"contacts"`
	Createdt     int64  `orm:"size(10);default(0)" json:"createdt"`
	Updatedt     int64  `orm:"size(10);default(0)" json:"updatedt"`
	LastLogindt  int64  `orm:"size(10);default(0)" json:"lastLogindt"`
	CredsArrJson string `json:"credsArrJson"`
}

type SessionUser struct {
	*User
	OtherUsers []string `json:"otherUsers"`
}

type ContactItem struct {
	UserId   int64  `json:"userId"`
	UserName string `json:"userName"`
	Addeddt  int64  `json:"addeddt"`
}

type UserDisplay struct {
	*User
	UsdBalance      float64 `json:"usdBalance"`
	CreatedtDisplay string  `json:"createdtDisplay"`
}

type UserInfo struct {
	Id       int64  `json:"id"`
	UserName string `json:"username"`
}
