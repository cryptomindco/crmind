package db

import (
	"crchat/pkg/config"
	"crchat/pkg/models"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

func Init(config config.Config) Handler {
	db, err := gorm.Open(postgres.Open(config.DBUrl), &gorm.Config{})

	if err != nil {
		log.Fatalln(err)
	}
	db.AutoMigrate(
		&models.ChatMsg{},
		&models.ChatContent{},
	)
	return Handler{DB: db}
}

func (h *Handler) GetChatMsgList(username string) ([]*models.ChatMsg, error) {
	chatMsgList := make([]*models.ChatMsg, 0)
	queryBuilder := fmt.Sprintf("SELECT tb.id,tb.from_name,tb.to_name,tb.createdt,tb.pin_msg,tb.updatedt "+
		"FROM (SELECT cm.*, cc.maxdate as last_contentdt FROM (SELECT * FROM chat_msg WHERE from_name = '%s' OR to_name = '%s') cm "+
		"JOIN (SELECT chat_id,MAX(createdt) AS maxdate FROM chat_content GROUP BY chat_id) cc ON cm.id = cc.chat_id ORDER BY cc.maxdate DESC) tb", username, username)

	err := h.DB.Raw(queryBuilder).Scan(&chatMsgList).Error
	if err != nil {
		return make([]*models.ChatMsg, 0), err
	}
	return chatMsgList, nil
}

func (h *Handler) GetChatContentList(chatId int64) ([]*models.ChatContent, error) {
	chatContentList := make([]*models.ChatContent, 0)
	queryBuilder := fmt.Sprintf("SELECT * FROM chat_content WHERE chat_id = %d ORDER BY createdt", chatId)
	err := h.DB.Raw(queryBuilder).Scan(&chatContentList).Error
	if err != nil {
		return chatContentList, err
	}
	return chatContentList, nil
}

func (h *Handler) CheckExistChat(fromName, toName string) (bool, error) {
	var chatMsg models.ChatMsg
	queryBuilder := fmt.Sprintf("SELECT * FROM chat_msg WHERE (from_name='%s' AND to_name='%s') OR (from_name='%s' AND to_name='%s')", fromName, toName, toName, fromName)
	getChatMsgErr := h.DB.Raw(queryBuilder).Scan(&chatMsg).Error
	if getChatMsgErr != nil && getChatMsgErr != gorm.ErrRecordNotFound {
		return false, getChatMsgErr
	}
	if getChatMsgErr == gorm.ErrRecordNotFound {
		return false, nil
	}
	return true, nil
}

func (h *Handler) GetUnreadChatContentNumber(excludeUsername string, chatId int64) int {
	//Get unread number of chat
	var unreadCount int64
	queryBuilder := fmt.Sprintf("SELECT COUNT(*) FROM chat_content WHERE chat_id = %d AND seen AND user_name <> '%s'", chatId, excludeUsername)
	err := h.DB.Raw(queryBuilder).Scan(&unreadCount)
	if err != nil {
		return 0
	}
	return int(unreadCount)
}

func (h *Handler) GetUnreadChatCount(username string) int {
	//Get unread number of chat
	queryBuilder := fmt.Sprintf("SELECT COUNT(*) FROM (SELECT chat_id, count(*) FROM "+
		"chat_content WHERE seen = false AND user_name <> '%s' AND chat_id IN (SELECT id FROM chat_msg WHERE from_name = '%s' OR to_name = '%s') GROUP BY chat_id) cc", username, username, username)
	var count int
	err := h.DB.Raw(queryBuilder).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}
