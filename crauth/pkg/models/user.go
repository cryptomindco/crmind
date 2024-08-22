package models

type User struct {
	Id           int64  `json:"id" gorm:"primaryKey"`
	Username     string `json:"username" gorm:"unique"`
	Name         string `json:"name"`
	Role         int    `json:"role"`
	Status       int    `json:"status"`
	Createdt     int64  `json:"createdt"`
	Updatedt     int64  `json:"updatedt"`
	LastLogindt  int64  `json:"lastLogindt"`
	CredsArrJson string `json:"credsArrJson"`
}
