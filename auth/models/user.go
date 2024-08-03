package models

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
