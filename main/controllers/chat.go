package controllers

import (
	"crmind/pb/chatpb"
	"crmind/services"
	"crmind/utils"
	"strings"
)

type ChatController struct {
	BaseController
}

func (this *ChatController) UpdateUnreadForChat() {
	loginUser, err := this.GetLoginUser()
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	chatId, chatIdErr := this.GetInt64("chatId")
	if chatIdErr != nil {
		this.ResponseLoginError(loginUser.Id, chatIdErr.Error(), utils.GetFuncName(), chatIdErr)
		return
	}
	_, err = services.UpdateUnreadForChatHandler(this.Ctx.Request.Context(), &chatpb.UpdateUnreadForChatRequest{
		ChatId:   chatId,
		UserName: loginUser.Username,
		Common: &chatpb.CommonRequest{
			LoginName: loginUser.Username,
		},
	})

	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	this.ResponseSuccessfully(loginUser.Id, "Update unread successfully", utils.GetFuncName())
}

func (this *ChatController) DeleteChat() {
	loginUser, err := this.GetLoginUser()
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	chatId, chatIdErr := this.GetInt64("chatId")
	if chatIdErr != nil {
		this.ResponseLoginError(loginUser.Id, chatIdErr.Error(), utils.GetFuncName(), chatIdErr)
		return
	}

	_, err = services.DeleteChatHandler(this.Ctx.Request.Context(), &chatpb.DeleteChatRequest{
		Common: &chatpb.CommonRequest{
			LoginName: loginUser.Username,
		},
		ChatId: chatId,
	})
	if err != nil {
		this.ResponseLoginError(loginUser.Id, "can't delete chat", utils.GetFuncName(), err)
		return
	}
	this.ResponseSuccessfully(loginUser.Id, "Delete chat successfully", utils.GetFuncName())
}

func (this *ChatController) SendChatMessage() {
	loginUser, err := this.GetLoginUser()
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	chatId, chatIdErr := this.GetInt64("chatId")
	if chatIdErr != nil {
		this.ResponseLoginError(loginUser.Id, chatIdErr.Error(), utils.GetFuncName(), chatIdErr)
		return
	}
	res, err := services.SendChatMessageHandler(this.Ctx.Request.Context(), &chatpb.SendChatMessageRequest{
		Common: &chatpb.CommonRequest{
			LoginName: loginUser.Username,
		},
		ChatId:   chatId,
		FromName: strings.TrimSpace(this.GetString("fromName")),
		ToName:   strings.TrimSpace(this.GetString("toName")),
		NewMsg:   strings.TrimSpace(this.GetString("newMsg")),
	})
	if err != nil {
		this.ResponseError("can't send new chat message", utils.GetFuncName(), err)
		return
	}
	this.ResponseSuccessfullyWithAnyData(loginUser.Id, "Send new chat successfully", utils.GetFuncName(), res.Data)
}
