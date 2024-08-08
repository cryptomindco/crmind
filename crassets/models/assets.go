package models

//Assets for wallets
type Asset struct {
	Id             int64   `orm:"column(id);auto;size(11)" json:"id"`
	DisplayName    string  `json:"displayName"`
	UserId         int64   `json:"userId"`
	UserName       string  `json:"userName"`
	IsAdmin        bool    `json:"isAdmin"`
	Type           string  `json:"type"`
	Sort           int     `json:"sort"`
	Balance        float64 `orm:"default(0)" json:"balance"`
	OnChainBalance float64 `orm:"default(0)" json:"onChainBalance"`
	LocalReceived  float64 `orm:"default(0)" json:"localReceived"`
	LocalSent      float64 `orm:"default(0)" json:"localSent"`
	ChainReceived  float64 `orm:"default(0)" json:"chainReceived"`
	ChainSent      float64 `orm:"default(0)" json:"chainSent"`
	TotalFee       float64 `orm:"default(0)" json:"totalFee"`
	Status         int     `json:"status"`
	Createdt       int64   `orm:"size(10);default(0)" json:"createdt"`
	Updatedt       int64   `orm:"size(10);default(0)" json:"updatedt"`
}

type AssetDisplay struct {
	Type               string  `json:"type"`
	TypeDisplay        string  `json:"typeDisplay"`
	Balance            float64 `json:"balance"`
	DaemonBalance      float64 `json:"daemonBalance"`
	SpendableFund      float64 `json:"spendableFund"`
	TotalChainReceived float64 `json:"totalChainReceived"`
	TotalChainSent     float64 `json:"totalChainSent"`
}

type Addresses struct {
	Id            int64   `orm:"column(id);auto;size(11)" json:"id"`
	AssetId       int64   `json:"assetId"`
	Address       string  `json:"address"`
	Label         string  `json:"label"`
	LocalReceived float64 `orm:"default(0)" json:"localReceived"`
	ChainReceived float64 `orm:"default(0)" json:"chainReceived"`
	Transactions  int     `orm:"default(0)" json:"transactions"`
	Archived      bool    `json:"archived"`
	Createdt      int64   `orm:"size(10);default(0)" json:"createdt"`
}

type AddressesDisplay struct {
	Addresses
	CreatedtDisplay string  `json:"createdtDisplay"`
	TotalReceived   float64 `json:"totalReceived"`
	LabelMainPart   string  `json:"labelMainPart"`
}

type Summary struct {
	TotalTransaction int     `json:"totalTransaction"`
	TotalSpent       float64 `json:"totalSpent"`
	TotalReceived    float64 `json:"totalReceived"`
	TotalFees        float64 `json:"totalFees"`
}
