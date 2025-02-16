package models

type Settings struct {
	Id             int64   `orm:"column(id);auto;size(11)" json:"id"`
	ActiveAssets   string  `json:"activeAssets"`
	ActiveServices string  `json:"activeServices"`
	RateServer     string  `json:"rateServer"`
	PriceSpread    float64 `json:"priceSpread"`
}
