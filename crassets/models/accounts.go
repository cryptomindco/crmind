package models

type Accounts struct {
	Id       int64  `orm:"column(id);auto;size(11)" json:"id"`
	UserId   int64  `json:"userId"`
	Username string `json:"username"`
	Role     int    `json:"role"`
	Contacts string `json:"contacts"`
	Token    string `json:"token"`
}
