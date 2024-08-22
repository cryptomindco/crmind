package models

type ChatMsg struct {
	Id       int64  `json:"id" gorm:"primaryKey"`
	FromName string `json:"fromName"`
	ToName   string `json:"toName"`
	PinMsg   string `json:"pinMsg"`
	Createdt int64  `json:"createdt"`
	Updatedt int64  `json:"updatedt"`
}

type ChatContent struct {
	Id       int64  `json:"id" gorm:"primaryKey"`
	ChatId   int64  `json:"chatId"`
	UserName string `json:"userName"`
	Content  string `json:"content"`
	IsHello  bool   `json:"isHello"`
	Seen     bool   `json:"seen"`
	Createdt int64  `json:"createdt"`
}

type ChatDisplay struct {
	*ChatMsg
	HasContent      bool           `json:"hasContent"`
	TargetUser      string         `json:"targetUser"`
	LastContent     ChatContent    `json:"lastContent"`
	ChatContentList []*ChatContent `json:"chatContentList"`
	UnreadNum       int            `json:"unreadNum"`
}
