package models

type AssetDisplay struct {
	Type               string  `json:"type"`
	TypeDisplay        string  `json:"typeDisplay"`
	Balance            float64 `json:"balance"`
	DaemonBalance      float64 `json:"daemonBalance"`
	SpendableFund      float64 `json:"spendableFund"`
	TotalChainReceived float64 `json:"totalChainReceived"`
	TotalChainSent     float64 `json:"totalChainSent"`
}

// Assets for wallets
type Asset struct {
	Id             int64   `orm:"column(id);auto;size(11)" json:"id"`
	DisplayName    string  `json:"displayName"`
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

type TxHistory struct {
	Id            int64   `orm:"column(id);auto;size(11)" json:"id"`
	TransType     int     `json:"transType"`
	SenderId      int64   `json:"senderId"`
	Sender        string  `json:"sender"`
	ReceiverId    int64   `json:"receiverId"`
	Receiver      string  `json:"receiver"`
	Currency      string  `json:"currency"`
	Amount        float64 `json:"amount"`
	Rate          float64 `json:"rate"`
	Status        int     `json:"status"`
	ToAddress     string  `json:"toAddress"`
	Txid          string  `json:"txid"`
	Fee           float64 `json:"fee"`
	IsTrading     bool    `json:"isTrading"`
	TradingType   string  `json:"tradingType"`
	PaymentType   string  `json:"paymentType"`
	Confirmed     bool    `json:"confirmed"`
	Confirmations int     `orm:"default(0)" json:"confirmations"`
	Description   string  `json:"description"`
	Createdt      int64   `orm:"size(10);default(0)" json:"createdt"`
}

type TxHistoryDisplay struct {
	TxHistory
	RateValue            float64           `json:"rateValue"`
	IsSender             bool              `json:"isSender"`
	TypeDisplay          string            `json:"typeDisplay"`
	ConfirmationNeed     int               `json:"confirmationNeed"`
	Transaction          TransactionResult `json:"transaction"`
	IsOffChain           bool              `json:"isOffChain"`
	CreatedtDisp         string            `json:"createdtDisp"`
	TradingPaymentAmount float64           `json:"tradingPaymentAmount"`
}

type TransactionResult struct {
	Amount          float64                       `json:"amount"`
	Fee             float64                       `json:"fee,omitempty"`
	Confirmations   int64                         `json:"confirmations"`
	BlockHash       string                        `json:"blockhash"`
	BlockIndex      int64                         `json:"blockindex"`
	BlockTime       int64                         `json:"blocktime"`
	TxID            string                        `json:"txid"`
	WalletConflicts []string                      `json:"walletconflicts"`
	Time            int64                         `json:"time"`
	TimeReceived    int64                         `json:"timereceived"`
	Details         []GetTransactionDetailsResult `json:"details"`
	Hex             string                        `json:"hex"`
}

type GetTransactionDetailsResult struct {
	Account           string   `json:"account"`
	Address           string   `json:"address,omitempty"`
	Amount            float64  `json:"amount"`
	Category          string   `json:"category"`
	InvolvesWatchOnly bool     `json:"involveswatchonly,omitempty"`
	Fee               *float64 `json:"fee,omitempty"`
	Vout              uint32   `json:"vout"`
}

// Assets for wallets
type TxCode struct {
	Id        int64   `json:"id" gorm:"primaryKey"`
	Asset     string  `json:"asset"`
	Code      string  `json:"code"`
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
