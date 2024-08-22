package models

type Rates struct {
	Id            int64  `json:"id" gorm:"primaryKey"`
	UsdRate       string `json:"usdRate"`
	AllRate       string `json:"allRate"`
	YesterdayRate string `json:"yesterdayRate"`
	LastMonthRate string `json:"lastMonthRate"`
	Updatedt      int64  `json:"updatedt"`
}

type RateObject struct {
	UsdRates map[string]float64 `json:"usdRates"`
	AllRates map[string]float64 `json:"allRates"`
}
