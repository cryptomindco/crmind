package utils

import (
	"crchat/models"
	"fmt"

	"github.com/beego/beego/v2/client/orm"
)

func CheckExistChat(fromId int64, toId int64) (bool, error) {
	var chatMsg models.ChatMsg
	queryBuilder := fmt.Sprintf("SELECT * FROM chat_msg WHERE (from_id=%d AND to_id=%d) OR (from_id=%d AND to_id=%d)", fromId, toId, toId, fromId)
	o := orm.NewOrm()
	getChatMsgErr := o.Raw(queryBuilder).QueryRow(&chatMsg)
	if getChatMsgErr != nil && getChatMsgErr != orm.ErrNoRows {
		return false, getChatMsgErr
	}
	if getChatMsgErr == orm.ErrNoRows {
		return false, nil
	}
	return true, nil
}
