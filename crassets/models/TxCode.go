package models

//Assets for wallets
type TxCode struct {
	Id        int64   `orm:"column(id);auto;size(11)" json:"id"`
	Asset     string  `json:"asset"`
	Code      string  `json:"code"`
	OwnerId   int64   `json:"ownerId"`
	OwnerName string  `json:"ownerName"`
	Amount    float64 `json:"amount"`
	Txid      string  `json:"txid"` //save when confirm
	HistoryId int64   `json:"historyId"`
	Status    int     `json:"status"`
	Note      string  `json:"note"`
	Createdt  int64   `orm:"size(10);default(0)" json:"createdt"`
	Confirmdt int64   `orm:"size(10);default(0)" json:"confirmdt"`
}

type TxCodeDisplay struct {
	TxCode
	StatusDisplay    string     `json:"statusDisplay"`
	CreatedtDisplay  string     `json:"createdtDisplay"`
	ConfirmdtDisplay string     `json:"confirmdtDisplay"`
	IsCancelled      bool       `json:"isCancelled"`
	IsConfirmed      bool       `json:"isConfirmed"`
	IsCreatedt       bool       `json:"isCreatedt"`
	TxHistory        *TxHistory `json:"txHistory"`
}
