package services

import (
	"context"
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

func (s *Server) UpdateUnreadForChat(ctx context.Context, reqData *pb.UpdateUnreadForChatRequest) (*pb.ResponseData, error) {
	chatId := reqData.ChatId
	userName := reqData.UserName
	if chatId < 1 || utils.IsEmpty(userName) {
		return ResponseLoginError(reqData.Common.LoginName, "Param not found", utils.GetFuncName(), nil)
	}
	//get all chat content with chat Id and exclude loginUser
	chatContentList := make([]models.ChatContent, 0)
	listErr := s.H.DB.Where("chat_id = ? AND user_name <> ?", chatId, userName).Find(&chatContentList).Error
	if listErr != nil {
		return ResponseLoginError(reqData.Common.LoginName, "Get chat content list to update unread count failed!", utils.GetFuncName(), nil)
	}
	//create tx
	tx := s.H.DB.Begin()
	//update all chat content, set seen to true
	for _, chatContent := range chatContentList {
		chatContent.Seen = true
		err := tx.Save(&chatContent).Error
		if err != nil {
			tx.Rollback()
			return ResponseLoginError(reqData.Common.LoginName, "Update seen for chat content failed", utils.GetFuncName(), err)
		}
	}
	tx.Commit()
	return ResponseSuccessfully(reqData.Common.LoginName, "Update seen for chat contents successfully", utils.GetFuncName())
}

func (s *Server) DeleteChat(ctx context.Context, reqData *pb.DeleteChatRequest) (*pb.ResponseData, error) {
	chatId := reqData.ChatId
	if chatId < 1 {
		return ResponseLoginError(reqData.Common.LoginName, "Chat ID param not found", utils.GetFuncName(), nil)
	}
	//create tx
	tx := s.H.DB.Begin()
	delQuery := fmt.Sprintf("DELETE FROM chat_content WHERE chat_id = %d", chatId)
	//delete chat content
	err := tx.Exec(delQuery).Error
	if err != nil {
		tx.Rollback()
		return ResponseLoginError(reqData.Common.LoginName, "Delete chat content failed. Please try again!", utils.GetFuncName(), err)
	}

	//delete chat msg
	delQuery = fmt.Sprintf("DELETE FROM chat_msg WHERE id = %d", chatId)
	//delete chat content
	err = tx.Exec(delQuery).Error
	if err != nil {
		tx.Rollback()
		return ResponseLoginError(reqData.Common.LoginName, "Delete chat msg failed. Please try again!", utils.GetFuncName(), err)
	}
	tx.Commit()
	return ResponseSuccessfully(reqData.Common.LoginName, "Delete chat successfully!", utils.GetFuncName())
}

func (s *Server) CheckAndCreateChat(ctx context.Context, reqData *pb.CheckAndCreateChatRequest) (*pb.ResponseData, error) {
	toName := reqData.ToName
	if utils.IsEmpty(toName) {
		return ResponseLoginError(reqData.Common.LoginName, "Param not found", utils.GetFuncName(), nil)
	}
	//check if exist chat conversion
	chatExist, err := s.H.CheckExistChat(reqData.Common.LoginName, toName)
	if err != nil {
		return ResponseLoginError(reqData.Common.LoginName, "Check chat exist failed", utils.GetFuncName(), err)
	}
	if err == nil && !chatExist {
		//Create new chat
		newChatMsg := &models.ChatMsg{
			FromName: reqData.Common.LoginName,
			ToName:   toName,
			Createdt: time.Now().Unix(),
			Updatedt: time.Now().Unix(),
		}
		tx := s.H.DB.Begin()
		//insert new ChatMsg
		chatInsertErr := tx.Create(newChatMsg).Error
		if chatInsertErr != nil {
			tx.Rollback()
			return ResponseLoginError(reqData.Common.LoginName, "Create new chat failed", utils.GetFuncName(), chatInsertErr)
		} else {
			helloChat := &models.ChatContent{
				ChatId:   newChatMsg.Id,
				UserName: reqData.Common.LoginName,
				Content:  fmt.Sprintf("%s has added %s to contacts. Start chatting now", reqData.Common.LoginName, toName),
				IsHello:  true,
				Createdt: time.Now().Unix(),
			}
			//insert to chat content
			err := tx.Create(helloChat).Error
			if err != nil {
				tx.Rollback()
				return ResponseLoginError(reqData.Common.LoginName, "Create hello chat content failed", utils.GetFuncName(), err)
			}
		}
		tx.Commit()
	}
	return ResponseSuccessfully(reqData.Common.LoginName, "Create hello chat successfully", utils.GetFuncName())
}

func (s *Server) SendChatMessage(ctx context.Context, reqData *pb.SendChatMessageRequest) (*pb.ResponseData, error) {
	chatId := reqData.ChatId
	fromName := reqData.FromName
	toName := reqData.ToName
	newMsg := reqData.NewMsg
	if chatId < 1 || utils.IsEmpty(fromName) || utils.IsEmpty(toName) || utils.IsEmpty(newMsg) {
		return ResponseError("Param failed", utils.GetFuncName(), nil)
	}

	if reqData.Common.LoginName != fromName && reqData.Common.LoginName != toName {
		return ResponseLoginError(reqData.Common.LoginName, "Don't have access to this feature", utils.GetFuncName(), nil)
	}
	var chatMsgId int64
	var newChatMsg *models.ChatMsg
	var otherName string
	if reqData.Common.LoginName == fromName {
		otherName = toName
	} else if reqData.Common.LoginName == toName {
		otherName = fromName
	}
	//create tx
	tx := s.H.DB.Begin()
	//if chat ID is empty, create new msg
	if chatId <= 0 {
		//check msg object Exist on DB
		var chatMsg models.ChatMsg
		queryBuilder := fmt.Sprintf("SELECT * FROM chat_msg WHERE (from_name = '%s' AND to_name = '%s') OR (from_name= '%s' AND to_name = '%s')", fromName, toName, toName, fromName)
		getChatMsgErr := s.H.DB.Raw(queryBuilder).Scan(&chatMsg).Error
		if getChatMsgErr != nil && getChatMsgErr != gorm.ErrRecordNotFound {
			tx.Rollback()
			return ResponseLoginError(reqData.Common.LoginName, "Check chat msg from DB failed", utils.GetFuncName(), getChatMsgErr)
		}
		//if exist
		if getChatMsgErr != gorm.ErrRecordNotFound {
			chatMsgId = chatMsg.Id
			newChatMsg = &chatMsg
		} else {
			//create new
			newChatMsg = &models.ChatMsg{
				FromName: fromName,
				ToName:   toName,
				Createdt: time.Now().Unix(),
				Updatedt: time.Now().Unix(),
			}
			var newErr error
			newErr = tx.Create(newChatMsg).Error
			if newErr != nil {
				return ResponseLoginRollbackError(reqData.Common.LoginName, tx, "Create new chat message failed. Please try again!", utils.GetFuncName(), newErr)
			}
		}
	} else {
		chatMsgId = chatId
	}

	//insert new Msg content
	newContent := &models.ChatContent{
		ChatId:   chatMsgId,
		UserName: reqData.Common.LoginName,
		Content:  newMsg,
		Createdt: time.Now().Unix(),
	}
	//insert
	newContentErr := tx.Create(newContent).Error
	if newContentErr != nil {
		return ResponseLoginRollbackError(reqData.Common.LoginName, tx, "Create new message failed. Please try again!", utils.GetFuncName(), newContentErr)
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
	return ResponseSuccessfullyWithAnyData(reqData.Common.LoginName, "Send message successfully", utils.GetFuncName(), resultObject)
}

func (s *Server) CheckChatExist(ctx context.Context, reqData *pb.CheckChatExistRequest) (*pb.ResponseData, error) {
	fromName := reqData.FromName
	toName := reqData.ToName
	if utils.IsEmpty(fromName) || utils.IsEmpty(toName) {
		return ResponseLoginError(reqData.Common.LoginName, "Param not found", utils.GetFuncName(), nil)
	}
	if reqData.Common.LoginName != fromName && reqData.Common.LoginName != toName {
		return ResponseLoginError(reqData.Common.LoginName, "No permission for this feature", utils.GetFuncName(), fmt.Errorf("No permission for this feature"))
	}
	exist, err := s.H.CheckExistChat(fromName, toName)
	if err != nil {
		return ResponseLoginError(reqData.Common.LoginName, "Check chat exist failed", utils.GetFuncName(), err)
	}
	return ResponseSuccessfullyWithAnyData(reqData.Common.LoginName, "Check chat exist successfully", utils.GetFuncName(), exist)
}

func (s *Server) GetChatMsgDisplayList(ctx context.Context, reqData *pb.CommonRequest) (*pb.ResponseData, error) {
	chatMsgList, _ := s.H.GetChatMsgList(reqData.LoginName)
	result := make([]*models.ChatDisplay, 0)
	for _, chatMsg := range chatMsgList {
		chatDisplay := &models.ChatDisplay{
			ChatMsg: chatMsg,
		}
		chatContentList, _ := s.H.GetChatContentList(chatMsg.Id)
		if len(chatContentList) > 0 {
			chatDisplay.LastContent = *chatContentList[len(chatContentList)-1]
		}
		if chatMsg.FromName == reqData.LoginName {
			chatDisplay.TargetUser = chatMsg.ToName
		} else {
			chatDisplay.TargetUser = chatMsg.FromName
		}
		chatDisplay.ChatContentList = chatContentList
		chatDisplay.HasContent = len(chatContentList) > 0
		chatDisplay.UnreadNum = s.H.GetUnreadChatContentNumber(reqData.LoginName, chatMsg.Id)
		result = append(result, chatDisplay)
	}
	//get unread chat count
	responseData := struct {
		ChatList    []*models.ChatDisplay `json:"chatList"`
		UnreadCount int                   `json:"unreadCount"`
	}{
		ChatList:    result,
		UnreadCount: s.H.GetUnreadChatCount(reqData.LoginName),
	}
	return ResponseSuccessfullyWithAnyData(reqData.LoginName, "Get msg display list successfully", utils.GetFuncName(), responseData)
}
