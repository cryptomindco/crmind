package services

import (
	"crchat/pkg/db"
	"crchat/pkg/models"
	"crchat/pkg/pb"
	"crchat/pkg/utils"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Server struct {
	H db.Handler // handler
	pb.UnimplementedChatServiceServer
}

func (s *Server) UpdateUnreadForChat(reqData *pb.RequestData) *pb.ResponseData {
	chatIDAny, isExist := reqData.DataMap["chatId"]
	userIDAny, userIdExist := reqData.DataMap["userId"]
	if !isExist || !userIdExist {
		return pb.ResponseError("Param not found", utils.GetFuncName(), nil)
	}
	chatId := chatIDAny.(int64)
	userId := userIDAny.(int64)
	//get all chat content with chat Id and exclude loginUser
	chatContentList := make([]models.ChatContent, 0)
	listErr := s.H.DB.Where("chat_id = ? AND user_id <> ?", chatId, userId).Find(&chatContentList).Error
	if listErr != nil {
		return pb.ResponseError("Get chat content list to update unread count failed!", utils.GetFuncName(), nil)
	}
	//create tx
	tx := s.H.DB.Begin()
	//update all chat content, set seen to true
	for _, chatContent := range chatContentList {
		chatContent.Seen = true
		err := tx.Save(&chatContent).Error
		if err != nil {
			tx.Rollback()
			return pb.ResponseError("Update seen for chat content failed", utils.GetFuncName(), err)
		}
	}
	tx.Commit()
	return pb.ResponseSuccessfully(0, "Update seen for chat contents successfully", utils.GetFuncName())
}

func (s *Server) DeleteChat(reqData *pb.RequestData) *pb.ResponseData {
	chatIDAny, isExist := reqData.DataMap["chatId"]
	if !isExist {
		return pb.ResponseError("Chat ID param not found", utils.GetFuncName(), nil)
	}
	chatId := chatIDAny.(int64)
	//create tx
	tx := s.H.DB.Begin()
	delQuery := fmt.Sprintf("DELETE FROM chat_content WHERE chat_id = %d", chatId)
	//delete chat content
	err := tx.Exec(delQuery).Error
	if err != nil {
		tx.Rollback()
		return pb.ResponseError("Delete chat content failed. Please try again!", utils.GetFuncName(), err)
	}

	//delete chat msg
	delQuery = fmt.Sprintf("DELETE FROM chat_msg WHERE id = %d", chatId)
	//delete chat content
	err = tx.Exec(delQuery).Error
	if err != nil {
		tx.Rollback()
		return pb.ResponseError("Delete chat msg failed. Please try again!", utils.GetFuncName(), err)
	}
	tx.Commit()
	return pb.ResponseSuccessfully(0, "Delete chat successfully!", utils.GetFuncName())
}

func (s *Server) CheckAndCreateChat(reqData *pb.RequestData) *pb.ResponseData {
	toIDAny, toIdExist := reqData.DataMap["toId"]
	toNameAny, toNameExist := reqData.DataMap["toName"]
	if !toIdExist || !toNameExist || reqData.LoginId < 1 || utils.IsEmpty(reqData.LoginName) {
		return pb.ResponseError("Param not found", utils.GetFuncName(), nil)
	}
	toId := toIDAny.(int64)
	toName := toNameAny.(string)
	//check if exist chat conversion
	chatExist, err := s.H.CheckExistChat(reqData.LoginId, toId)
	if err != nil {
		return pb.ResponseError("Check chat exist failed", utils.GetFuncName(), err)
	}
	if err == nil && !chatExist {
		//Create new chat
		newChatMsg := &models.ChatMsg{
			FromId:   reqData.LoginId,
			FromName: reqData.LoginName,
			ToId:     toId,
			ToName:   toName,
			Createdt: time.Now().Unix(),
			Updatedt: time.Now().Unix(),
		}
		tx := s.H.DB.Begin()
		//insert new ChatMsg
		chatInsertErr := tx.Create(newChatMsg).Error
		if chatInsertErr != nil {
			return pb.ResponseError("Create new chat failed", utils.GetFuncName(), chatInsertErr)
		} else {
			helloChat := &models.ChatContent{
				ChatId:   newChatMsg.Id,
				UserId:   reqData.LoginId,
				UserName: reqData.LoginName,
				Content:  fmt.Sprintf("%s has added %s to contacts. Start chatting now", reqData.LoginName, toName),
				IsHello:  true,
				Createdt: time.Now().Unix(),
			}
			//insert to chat content
			err := tx.Create(helloChat).Error
			if err != nil {
				return pb.ResponseError("Create hello chat content failed", utils.GetFuncName(), err)
			}
		}
	}
	return pb.ResponseSuccessfully(0, "Create hello chat successfully", utils.GetFuncName())
}

func (s *Server) SendChatMessage(reqData *pb.RequestData) *pb.ResponseData {
	chatIdAny, chatIdExist := reqData.DataMap["chatId"]
	fromNameAny, fromNameExist := reqData.DataMap["fromName"]
	fromIdAny, fromIdExist := reqData.DataMap["fromId"]
	toNameAny, toNameExist := reqData.DataMap["toName"]
	toIdAny, toIdExist := reqData.DataMap["toId"]
	newMsgAny, newMsgExist := reqData.DataMap["newMsg"]
	if !chatIdExist || !fromNameExist || !fromIdExist || !toNameExist || !toIdExist || !newMsgExist || reqData.LoginId < 1 || utils.IsEmpty(reqData.LoginName) {
		return pb.ResponseError("Param not found", utils.GetFuncName(), nil)
	}
	chatId := chatIdAny.(int64)
	fromName := fromNameAny.(string)
	fromId := fromIdAny.(int64)
	toName := toNameAny.(string)
	toId := toIdAny.(int64)
	newMsg := newMsgAny.(string)

	if reqData.LoginId != fromId && reqData.LoginId != toId {
		return pb.ResponseLoginError(reqData.LoginId, "Don't have access to this feature", utils.GetFuncName(), nil)
	}
	//create tx
	tx := s.H.DB.Begin()
	var chatMsgId int64
	var newChatMsg *models.ChatMsg
	var otherName string
	if reqData.LoginName == fromName {
		otherName = toName
	} else if reqData.LoginName == toName {
		otherName = fromName
	}
	//if chat ID is empty, create new msg
	if chatId <= 0 {
		//check msg object Exist on DB
		var chatMsg models.ChatMsg
		queryBuilder := fmt.Sprintf("SELECT * FROM chat_msg WHERE (from_name = '%s' AND to_name = '%s') OR (from_name= '%s' AND to_name = '%s')", fromName, toName, toName, fromName)
		getChatMsgErr := s.H.DB.Raw(queryBuilder).Scan(&chatMsg).Error
		if getChatMsgErr != nil && getChatMsgErr != gorm.ErrRecordNotFound {
			return pb.ResponseLoginError(reqData.LoginId, "Check chat msg from DB failed", utils.GetFuncName(), getChatMsgErr)
		}
		//if exist
		if getChatMsgErr != gorm.ErrRecordNotFound {
			chatMsgId = chatMsg.Id
			newChatMsg = &chatMsg
		} else {
			//create new
			newChatMsg = &models.ChatMsg{
				FromId:   fromId,
				FromName: fromName,
				ToId:     toId,
				ToName:   toName,
				Createdt: time.Now().Unix(),
				Updatedt: time.Now().Unix(),
			}
			var newErr error
			newErr = tx.Create(newChatMsg).Error
			if newErr != nil {
				return pb.ResponseLoginRollbackError(reqData.LoginId, tx, "Create new chat message failed. Please try again!", utils.GetFuncName(), newErr)
			}
		}
	} else {
		chatMsgId = chatId
	}

	//insert new Msg content
	newContent := &models.ChatContent{
		ChatId:   chatMsgId,
		UserId:   reqData.LoginId,
		UserName: reqData.LoginName,
		Content:  newMsg,
		Createdt: time.Now().Unix(),
	}
	//insert
	newContentErr := tx.Create(newContent).Error
	if newContentErr != nil {
		return pb.ResponseLoginRollbackError(reqData.LoginId, tx, "Create new message failed. Please try again!", utils.GetFuncName(), newContentErr)
	}
	tx.Commit()
	var newMsgObject models.ChatDisplay
	if newChatMsg != nil {
		newMsgObject = models.ChatDisplay{
			ChatMsg:         newChatMsg,
			HasContent:      true,
			TargetUser:      otherName,
			LastContent:     *newContent,
			ChatContentList: []*models.ChatContent{newContent},
		}
	}
	//Return result
	var resultObject = struct {
		NewMsg     models.ChatDisplay `json:"newMsg"`
		NewContent models.ChatContent `json:"newContent"`
	}{
		NewMsg:     newMsgObject,
		NewContent: *newContent,
	}
	return pb.ResponseSuccessfullyWithAnyData(reqData.LoginId, "Send message successfully", utils.GetFuncName(), resultObject)
}
func (s *Server) CheckChatExist(reqData *pb.RequestData) *pb.ResponseData {
	fromIdAny, toIdExist := reqData.DataMap["fromId"]
	toIdAny, toIdExist := reqData.DataMap["toId"]
	if !toIdExist || !toIdExist || reqData.LoginId < 1 || utils.IsEmpty(reqData.LoginName) {
		return pb.ResponseError("Param not found", utils.GetFuncName(), nil)
	}
	fromId := fromIdAny.(int64)
	toId := toIdAny.(int64)
	if reqData.LoginId != fromId && reqData.LoginId != toId {
		return pb.ResponseError("No permission for this feature", utils.GetFuncName(), fmt.Errorf("No permission for this feature"))
	}
	exist, err := s.H.CheckExistChat(fromId, toId)
	if err != nil {
		return pb.ResponseError("Check chat exist failed", utils.GetFuncName(), err)
	}
	return pb.ResponseSuccessfullyWithAnyData(reqData.LoginId, "Check chat exist successfully", utils.GetFuncName(), exist)
}

func (s *Server) GetChatMsgDisplayList(reqData *pb.RequestData) *pb.ResponseData {
	if reqData.LoginId < 1 {
		return pb.ResponseError("Get login Id failed", utils.GetFuncName(), fmt.Errorf("Get login Id failed"))
	}
	chatMsgList, _ := s.H.GetChatMsgList(reqData.LoginId)
	result := make([]*models.ChatDisplay, 0)
	for _, chatMsg := range chatMsgList {
		chatDisplay := &models.ChatDisplay{
			ChatMsg: chatMsg,
		}
		chatContentList, _ := s.H.GetChatContentList(chatMsg.Id)
		if len(chatContentList) > 0 {
			chatDisplay.LastContent = *chatContentList[len(chatContentList)-1]
		}
		if chatMsg.FromId == reqData.LoginId {
			chatDisplay.TargetUser = chatMsg.ToName
		} else {
			chatDisplay.TargetUser = chatMsg.FromName
		}
		chatDisplay.ChatContentList = chatContentList
		chatDisplay.HasContent = len(chatContentList) > 0
		chatDisplay.UnreadNum = s.H.GetUnreadChatContentNumber(reqData.LoginId, chatMsg.Id)
		result = append(result, chatDisplay)
	}
	//get unread chat count
	responseData := struct {
		ChatList    []*models.ChatDisplay `json:"chatList"`
		UnreadCount int                   `json:"unreadCount"`
	}{
		ChatList:    result,
		UnreadCount: s.H.GetUnreadChatCount(reqData.LoginId),
	}
	return pb.ResponseSuccessfullyWithAnyData(reqData.LoginId, "Get msg display list successfully", utils.GetFuncName(), responseData)
}
