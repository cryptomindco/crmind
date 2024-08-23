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
		"FROM (SELECT cm.*, cc.maxdate as last_contentdt FROM (SELECT * FROM chat_msgs WHERE from_name = '%s' OR to_name = '%s') cm "+
		"JOIN (SELECT chat_id,MAX(createdt) AS maxdate FROM chat_contents GROUP BY chat_id) cc ON cm.id = cc.chat_id ORDER BY cc.maxdate DESC) tb", username, username)

	err := h.DB.Raw(queryBuilder).Scan(&chatMsgList).Error
	if err != nil {
		return make([]*models.ChatMsg, 0), err
	}
	return chatMsgList, nil
}

func (h *Handler) GetChatContentList(chatId int64) ([]*models.ChatContent, error) {
	chatContentList := make([]*models.ChatContent, 0)
	queryBuilder := fmt.Sprintf("SELECT * FROM chat_contents WHERE chat_id = %d ORDER BY createdt", chatId)
	err := h.DB.Raw(queryBuilder).Scan(&chatContentList).Error
	if err != nil {
		return chatContentList, err
	}
	return chatContentList, nil
}

func (h *Handler) GetChatMsgFromId(chatId int64) (*models.ChatMsg, error) {
	var chatMsg models.ChatMsg
	err := h.DB.Where(&models.ChatMsg{Id: chatId}).First(&chatMsg).Error
	if err != nil {
		return nil, err
	}
	return &chatMsg, nil
}

func (h *Handler) CheckExistChat(fromName, toName string) (bool, error) {
	var count int64
	queryBuilder := fmt.Sprintf("SELECT COUNT(*) FROM chat_msgs WHERE (from_name='%s' AND to_name='%s') OR (from_name='%s' AND to_name='%s')", fromName, toName, toName, fromName)
	countMsgErr := h.DB.Raw(queryBuilder).Scan(&count).Error
	if countMsgErr != nil {
		return false, countMsgErr
	}
	return count > 0, nil
}

func (h *Handler) GetUnreadChatContentNumber(excludeUsername string, chatId int64) int {
	//Get unread number of chat
	var unreadCount int64
	queryBuilder := fmt.Sprintf("SELECT COUNT(*) FROM chat_contents WHERE chat_id = %d AND seen = false AND user_name <> '%s'", chatId, excludeUsername)
	err := h.DB.Raw(queryBuilder).Scan(&unreadCount).Error
	if err != nil {
		return 0
	}
	return int(unreadCount)
}

func (h *Handler) GetUnreadChatCount(username string) int {
	//Get unread number of chat
	queryBuilder := fmt.Sprintf("SELECT COUNT(*) FROM (SELECT chat_id, count(*) FROM "+
		"chat_contents WHERE seen = false AND user_name <> '%s' AND chat_id IN (SELECT id FROM chat_msgs WHERE from_name = '%s' OR to_name = '%s') GROUP BY chat_id) cc", username, username, username)
	var count int64
	err := h.DB.Raw(queryBuilder).Scan(&count).Error
	if err != nil {
		return 0
	}
	return int(count)
}
