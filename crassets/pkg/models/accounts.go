package models

type Accounts struct {
	Id       int64  `json:"id" gorm:"primaryKey"`
	Username string `json:"username"`
	Role     int    `json:"role"`
	Contacts string `json:"contacts"`
	Token    string `json:"token"`
}

type ContactItem struct {
	UserName string `json:"userName"`
	Addeddt  int64  `json:"addeddt"`
}

type UserInfo struct {
	Username string `json:"username"`
	Role     int    `json:"role"`
}
