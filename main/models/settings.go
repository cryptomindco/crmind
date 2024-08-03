package models

type Settings struct {
	Id           int64  `orm:"column(id);auto;size(11)" json:"id"`
	ActiveAssets string `json:"activeAssets"`
}
