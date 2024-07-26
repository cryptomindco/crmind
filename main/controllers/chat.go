package controllers

import (
	"crmind/services"
	"crmind/utils"
	"fmt"
	"net/url"
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
	chatIdStr := strings.TrimSpace(this.GetString("chatId"))

	formData := url.Values{
		"authorization": {this.GetLoginToken()},
		"chatId":        {chatIdStr},
	}
	var response utils.ResponseData
	if err := services.HttpPost(fmt.Sprintf("%s%s", this.ChatSite(), "/updateUnread"), formData, &response); err != nil {
		this.ResponseError("can't begin update unread chat", utils.GetFuncName(), err)
		return
	}
	if response.IsError {
		this.ResponseError(response.Msg, utils.GetFuncName(), fmt.Errorf(response.Msg))
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
	chatIdStr := strings.TrimSpace(this.GetString("chatId"))

	formData := url.Values{
		"authorization": {this.GetLoginToken()},
		"chatId":        {chatIdStr},
	}
	var response utils.ResponseData
	if err := services.HttpPost(fmt.Sprintf("%s%s", this.ChatSite(), "/deleteChat"), formData, &response); err != nil {
		this.ResponseError("can't delete chat", utils.GetFuncName(), err)
		return
	}
	if response.IsError {
		this.ResponseError(response.Msg, utils.GetFuncName(), fmt.Errorf(response.Msg))
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
	formData := url.Values{
		"authorization": {this.GetLoginToken()},
		"chatId":        {strings.TrimSpace(this.GetString("chatId"))},
		"fromName":      {strings.TrimSpace(this.GetString("fromName"))},
		"fromId":        {strings.TrimSpace(this.GetString("fromId"))},
		"toName":        {strings.TrimSpace(this.GetString("toName"))},
		"toId":          {strings.TrimSpace(this.GetString("toId"))},
		"newMsg":        {strings.TrimSpace(this.GetString("newMsg"))},
	}
	var response utils.ResponseData
	if err := services.HttpPost(fmt.Sprintf("%s%s", this.ChatSite(), "/sendChatMessage"), formData, &response); err != nil {
		this.ResponseError("can't send new chat message", utils.GetFuncName(), err)
		return
	}
	if response.IsError {
		this.ResponseError(response.Msg, utils.GetFuncName(), fmt.Errorf(response.Msg))
		return
	}
	this.ResponseSuccessfullyWithAnyData(loginUser.Id, "Send new chat successfully", utils.GetFuncName(), response.Data)
}
