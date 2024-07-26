package controllers

import (
	"crchat/models"
	"crchat/utils"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type ChatController struct {
	BaseController
}

func (this *ChatController) UpdateUnreadForChat() {
	authToken := this.GetString("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	chatIdStr := strings.TrimSpace(this.GetString("chatId"))
	chatId, parseErr := strconv.ParseInt(chatIdStr, 0, 32)
	if parseErr != nil {
		this.ResponseLoginError(loginUser.Id, "Param failed. Please try again!", utils.GetFuncName(), nil)
		return
	}

	//get all chat content with chat Id and exclude loginUser
	o := orm.NewOrm()
	chatContentList := make([]models.ChatContent, 0)
	_, listErr := o.QueryTable(chatContentModel).Filter("chat_id", chatId).Exclude("user_id", loginUser.Id).All(&chatContentList)
	if listErr != nil {
		this.ResponseLoginError(loginUser.Id, "Get chat content list to update unread count failed!", utils.GetFuncName(), nil)
		return
	}
	//create tx
	tx, beginErr := o.Begin()
	if beginErr != nil {
		this.ResponseLoginError(loginUser.Id, "An error has occurred. Please try again!", utils.GetFuncName(), beginErr)
		return
	}
	//update all chat content, set seen to true
	for _, chatContent := range chatContentList {
		chatContent.Seen = true
		_, err := tx.Update(&chatContent)
		if err != nil {
			this.ResponseLoginError(loginUser.Id, "Update seen for chat content failed", utils.GetFuncName(), err)
			tx.Rollback()
			return
		}
	}
	tx.Commit()
	this.ResponseSuccessfully(loginUser.Id, "Update seen for chat contents successfully", utils.GetFuncName())
}

func (this *ChatController) GetChatMsgList(userId int64) ([]*models.ChatMsg, error) {
	chatMsgList := make([]*models.ChatMsg, 0)
	o := orm.NewOrm()
	queryBuilder := fmt.Sprintf("SELECT tb.id,tb.from_id,tb.from_name,tb.to_id,tb.to_name,tb.createdt,tb.pin_msg,tb.updatedt "+
		"FROM (SELECT cm.*, cc.maxdate as last_contentdt FROM (SELECT * FROM chat_msg WHERE from_id = %d OR to_id = %d) cm "+
		"JOIN (SELECT chat_id,MAX(createdt) AS maxdate FROM chat_content GROUP BY chat_id) cc ON cm.id = cc.chat_id ORDER BY cc.maxdate DESC) tb", userId, userId)
	_, err := o.Raw(queryBuilder).QueryRows(&chatMsgList)
	if err != nil {
		return make([]*models.ChatMsg, 0), err
	}
	return chatMsgList, nil
}

func (this *ChatController) DeleteChat() {
	authToken := this.GetString("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	chatIdStr := strings.TrimSpace(this.GetString("chatId"))
	chatId, parseErr := strconv.ParseInt(chatIdStr, 0, 32)
	if parseErr != nil {
		this.ResponseLoginError(loginUser.Id, "Param failed. Please try again!", utils.GetFuncName(), nil)
		return
	}
	o := orm.NewOrm()
	//create tx
	tx, beginErr := o.Begin()
	if beginErr != nil {
		this.ResponseLoginError(loginUser.Id, "An error has occurred. Please try again!", utils.GetFuncName(), beginErr)
		return
	}

	//delete chat content
	_, delChatErr := tx.QueryTable(chatContentModel).Filter("chat_id", chatId).Delete()
	if delChatErr != nil {
		this.ResponseLoginError(loginUser.Id, "Delete chat content failed. Please try again!", utils.GetFuncName(), delChatErr)
		return
	}

	//delete chat msg
	_, delMsgErr := tx.QueryTable(chatMsgModel).Filter("id", chatId).Delete()
	if delMsgErr != nil {
		this.ResponseLoginError(loginUser.Id, "Delete chat msg failed. Please try again!", utils.GetFuncName(), delChatErr)
		return
	}

	tx.Commit()
	this.ResponseSuccessfully(loginUser.Id, "Delete chat successfully!", utils.GetFuncName())
}

func (this *ChatController) SendChatMessage() {
	authToken := this.GetString("authorization")
	//check login
	loginUser, err := this.AuthTokenCheck(authToken)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	chatIdStr := strings.TrimSpace(this.GetString("chatId"))
	fromName := strings.TrimSpace(this.GetString("fromName"))
	fromId, fromErr := this.GetInt64("fromId")
	toName := strings.TrimSpace(this.GetString("toName"))
	toId, toErr := this.GetInt64("toId")
	newMsg := strings.TrimSpace(this.GetString("newMsg"))
	//parse chatId
	chatId, parseErr := strconv.ParseInt(chatIdStr, 0, 32)
	if parseErr != nil || utils.IsEmpty(fromName) || utils.IsEmpty(toName) || utils.IsEmpty(newMsg) || fromErr != nil || toErr != nil {
		this.ResponseLoginError(loginUser.Id, "Param failed. Please try again!", utils.GetFuncName(), nil)
		return
	}
	if loginUser.Id != fromId && loginUser.Id != toId {
		this.ResponseLoginError(loginUser.Id, "Don't have access to this feature", utils.GetFuncName(), nil)
		return
	}
	o := orm.NewOrm()
	//create tx
	tx, beginErr := o.Begin()
	if beginErr != nil {
		this.ResponseLoginError(loginUser.Id, "An error has occurred. Please try again!", utils.GetFuncName(), beginErr)
		return
	}

	var chatMsgId int64
	var newChatMsg *models.ChatMsg
	var otherName string
	if loginUser.Username == fromName {
		otherName = toName
	} else if loginUser.Username == toName {
		otherName = fromName
	}
	//if chat ID is empty, create new msg
	if chatId <= 0 {
		//check msg object Exist on DB
		var chatMsg models.ChatMsg
		queryBuilder := fmt.Sprintf("SELECT * FROM chat_msg WHERE (from_name = '%s' AND to_name = '%s') OR (from_name= '%s' AND to_name = '%s')", fromName, toName, toName, fromName)
		getChatMsgErr := o.Raw(queryBuilder).QueryRow(&chatMsg)
		if getChatMsgErr != nil && getChatMsgErr != orm.ErrNoRows {
			this.ResponseLoginError(loginUser.Id, "Check chat msg from DB failed", utils.GetFuncName(), getChatMsgErr)
			return
		}
		//if exist
		if getChatMsgErr != orm.ErrNoRows {
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
			chatMsgId, newErr = tx.Insert(newChatMsg)
			if newErr != nil {
				this.ResponseLoginRollbackError(loginUser.Id, tx, "Create new chat message failed. Please try again!", utils.GetFuncName(), newErr)
				return
			}
		}
	} else {
		chatMsgId = chatId
	}

	//insert new Msg content
	newContent := &models.ChatContent{
		ChatId:   chatMsgId,
		UserId:   loginUser.Id,
		UserName: loginUser.Username,
		Content:  newMsg,
		Createdt: time.Now().Unix(),
	}
	//insert
	_, newContentErr := tx.Insert(newContent)
	if newContentErr != nil {
		this.ResponseLoginRollbackError(loginUser.Id, tx, "Create new message failed. Please try again!", utils.GetFuncName(), newContentErr)
		return
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
	this.ResponseSuccessfullyWithAnyData(loginUser.Id, "Send message successfully", utils.GetFuncName(), resultObject)
}

func (this *ChatController) GetChatContentList(chatId int64) ([]*models.ChatContent, error) {
	chatContentList := make([]*models.ChatContent, 0)
	o := orm.NewOrm()
	queryBuilder := fmt.Sprintf("SELECT * FROM chat_content WHERE chat_id = %d ORDER BY createdt", chatId)
	_, err := o.Raw(queryBuilder).QueryRows(&chatContentList)
	if err != nil {
		return make([]*models.ChatContent, 0), err
	}
	return chatContentList, nil
}

func (this *ChatController) GetChatMsgDisplayList() {
	userIdStr := this.Ctx.Request.Header.Get("UserId")
	userId, intErr := strconv.ParseInt(userIdStr, 0, 32)
	if intErr != nil {
		this.ResponseError(intErr.Error(), utils.GetFuncName(), intErr)
		return
	}
	//check login
	loginUser, err := this.AuthCheck()
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	if loginUser.Id != userId {
		this.ResponseError("User not match", utils.GetFuncName(), fmt.Errorf("User not match"))
		return
	}
	chatMsgList, _ := this.GetChatMsgList(userId)
	result := make([]*models.ChatDisplay, 0)
	for _, chatMsg := range chatMsgList {
		chatDisplay := &models.ChatDisplay{
			ChatMsg: chatMsg,
		}
		chatContentList, _ := this.GetChatContentList(chatMsg.Id)
		if len(chatContentList) > 0 {
			chatDisplay.LastContent = *chatContentList[len(chatContentList)-1]
		}
		if chatMsg.FromId == userId {
			chatDisplay.TargetUser = chatMsg.ToName
		} else {
			chatDisplay.TargetUser = chatMsg.FromName
		}
		chatDisplay.ChatContentList = chatContentList
		chatDisplay.HasContent = len(chatContentList) > 0
		chatDisplay.UnreadNum = this.GetUnreadChatContentNumber(userId, chatMsg.Id)
		result = append(result, chatDisplay)
	}
	//get unread chat count
	responseData := struct {
		ChatList    []*models.ChatDisplay `json:"chatList"`
		UnreadCount int                   `json:"unreadCount"`
	}{
		ChatList:    result,
		UnreadCount: this.GetUnreadChatCount(userId),
	}
	this.ResponseSuccessfullyWithAnyData(loginUser.Id, "Get msg display list successfully", utils.GetFuncName(), responseData)
}

func (this *ChatController) GetUnreadChatCount(userId int64) int {
	//Get unread number of chat
	o := orm.NewOrm()
	queryBuilder := fmt.Sprintf("SELECT COUNT(*) FROM (SELECT chat_id, count(*) FROM "+
		"chat_content WHERE seen = false AND user_id <> %d AND chat_id IN (SELECT id FROM chat_msg WHERE from_id = %d OR to_id = %d) GROUP BY chat_id) cc", userId, userId, userId)
	var count int
	err := o.Raw(queryBuilder).QueryRow(&count)
	if err != nil {
		return 0
	}
	return count
}

func (this *ChatController) GetUnreadChatContentNumber(excludeUserId int64, chatId int64) int {
	//Get unread number of chat
	o := orm.NewOrm()
	unreadCount, err := o.QueryTable(chatContentModel).Filter("chat_id", chatId).Filter("seen", false).Exclude("user_id", excludeUserId).Count()
	if err != nil {
		return 0
	}
	return int(unreadCount)
}
