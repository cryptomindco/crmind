import BaseController from "./base_controller";

export default class extends BaseController {
  static values = {
    chatList: Array,
    loginName: String,
    currentChatIndex: Number,
    createNewChatMsg: Object,
    createNewChatContent: Object,
    canSend: Boolean,
    newChatFlg: Boolean,
    unreadCount: Number,
    chatboxDisplay: Boolean,
    openedChatBox: Boolean,
    socket: Object,
    isMobileView: Boolean,
    contentShowing: Boolean,
  };

  async initialize() {
    if ($(window).width() > 768) {
      this.isMobileView = false
    } else {
      this.isMobileView = true
      $("#backToListBtn").removeClass('d-none')
    }
    const chatListJson = this.data.get("chatList");
    this.loginName = this.data.get("loginName");
    this.unreadCount = parseInt(this.data.get("unreadCount"));
    this.chatList = JSON.parse(chatListJson)
    this.currentChatIndex = 0
    this.canSend = false
    const _this = this
    // Create a socket
    const method = location.protocol.startsWith("https") ? 'wss' : 'ws'
    this.socket = new WebSocket(method + '://' + window.location.host + '/ws/connect');
    // Message received on the socket
    this.socket.onmessage = function (event) {
      var data = JSON.parse(event.data)
      //check type of message from socket
      switch(data.Type) {
        case EVENT_CHATMSG_TYPE:
        _this.handlerChatSocketSignal(JSON.parse(data.Content))  
        break;
        case EVENT_RATE_TYPE:
          const rateObject = JSON.parse(data.Content)
          RateJson = rateObject.usdRates
          AllRateJson = rateObject.allRates
          break;
        default:
          return
      }
    };

    $(window).resize(function(){
      const winWidth = $(window).width()
      if (winWidth > 768) {
        _this.isMobileView = false
        $("#chatMsgListCol").removeClass('d-none')
        $("#chatContentCol").removeClass('d-none')
        $("#backToListBtn").addClass('d-none')
        _this.initChatMsgList()
        return
      }
      _this.isMobileView = true
      $("#backToListBtn").removeClass('d-none')
      if (_this.contentShowing) {
        $("#chatMsgListCol").addClass('d-none')
        $("#chatContentCol").removeClass('d-none')
      } else {
        $("#chatMsgListCol").removeClass('d-none')
        $("#chatContentCol").addClass('d-none')
      }
      _this.initChatMsgList()
    })
  }

  handlerChatSocketSignal(content) {
      const _this = this
      //chat chatid in chat List and creator is not loginName
      var chatContent = content.lastContent
      var isMyValidUpdateChat = (content.fromName == _this.loginName || content.toName == _this.loginName) && chatContent.userName != _this.loginName
      //if is not my chat, return
      if (!isMyValidUpdateChat) {
        return
      }
      var targetName = content.fromName == _this.loginName ? content.toName : content.fromName
      var hasUpdate = false
      var updateIndex = -1
      _this.chatList.forEach((element, index) => {
        if(element.id == chatContent.chatId) {
          hasUpdate = true
          updateIndex = index
          return
        }
      })
      let isCurrent = false
      let hasUnread = false
      if(hasUpdate) {
        //update msg
        isCurrent = updateIndex == _this.currentChatIndex
        hasUnread = _this.chatList[updateIndex].unreadNum > 0
        _this.chatList[updateIndex].chatContentList.push(chatContent)
        _this.chatList[updateIndex].lastContent = chatContent
        //if is current display chat, update seen chat

        _this.chatList[updateIndex].unreadNum += 1
        if(updateIndex > 0) {
          var tempUpdateChatmsg = _this.chatList[updateIndex]
          if(updateIndex > _this.currentChatIndex) {
            _this.currentChatIndex += 1
          } else if (updateIndex == _this.currentChatIndex) {
            _this.currentChatIndex = 0
          }
          //remove update msg from list
          _this.chatList.splice(updateIndex, 1)
          //insert to first
          _this.chatList.unshift(tempUpdateChatmsg)
        }
        //swap first
      } else {
        //if is not update, insert new chat msg to list
        var contentList = []
        contentList.push(chatContent)
        content.chatContentList = content
        content.unreadNum += 1
        content.targetUser = targetName
        hasUnread = true
        isCurrent = false
        //insert to first of chat msg list
        _this.chatList.unshift(content)
        //increase current index 1
        if (_this.currentChatIndex != updateIndex) {
          _this.currentChatIndex += 1
        } else {
          _this.currentChatIndex = 0
        }
      }
      //update msg list
      _this.initChatMsgList()
      if(updateIndex == _this.currentChatIndex && _this.chatboxDisplay) {
        _this.updateUnreadCount(_this.chatList[_this.currentChatIndex])
      } else if((!_this.chatboxDisplay || (_this.chatboxDisplay && !isCurrent)) && !hasUnread) {
        _this.unreadCount++
          //if unread count is zero, remove badge
         if($("#floatingBtn").attr("data-after-type") == undefined) {
          $("#floatingBtn").attr("data-after-type", "red badge top right")
         }
        $("#floatingBtn").attr("data-after-text", _this.unreadCount)
      }
      //if is current chat, append to chat
      if(isCurrent) {
        $("#chatContentArea").append(_this.createChatContentHTML(_this.chatList[_this.currentChatIndex], chatContent))
        $('#chatContentArea').scrollTop($('#chatContentArea')[0].scrollHeight)
      }
  }

  swapChatMsgList(firstIndex, secondIndex, list) {
    var tempFirst = list[firstIndex]
    list[firstIndex] = list[secondIndex]
    list[secondIndex] = tempFirst
    return list
  }

  postMessage(msg) {
    this.socket.send(msg);
  }

  chatInputChange(e) {
    this.handlerSearchListDisplay()
  }

  handlerSearchListDisplay() {
    var filter, ul, li, a, i, txtValue;
    filter = $("#searchInput").val().toUpperCase();
    ul = document.getElementById("myUL");
    li = ul.getElementsByTagName("li");
    for (i = 0; i < li.length; i++) {
        a = li[i].getElementsByTagName("a")[0];
        txtValue = a.textContent || a.innerText;
        if (filter && txtValue.toUpperCase().indexOf(filter) > -1) {
            li[i].style.display = "";
        } else {
            li[i].style.display = "none";
        }
    }
  }

  newChat(e) {
    const username = e.params.username
    //check username on userlist
    let curIndex = -1
    const _this = this
    this.chatList.forEach((element, index) => {
      //if exist on list, open list
      if(username == element.fromName || username ==  element.toName) {
        curIndex = index
        return
      }
    })
    $("#searchInput").val("")
    this.handlerSearchListDisplay()
    if (this.isMobileView) {
      $("#chatMsgListCol").addClass('d-none')
      $("#chatContentCol").removeClass('d-none')
    }
    if(curIndex >= 0) {
      if(this.newChatFlg && curIndex > 0) {
        //if current is new chat, remove from list
        this.chatList.splice(this.currentChatIndex, 1)
        curIndex -= 1
        this.newChatFlg = false
        this.initChatMsgList()
      }
      if (!this.isMobileView){
        _this.changeChatItemActiveClass(_this.currentChatIndex, curIndex)
      }
      _this.currentChatIndex = curIndex
      _this.initChatMsgContent()
      return
    }
    //else, create new chat area
    this.createNewChatMsg(username)
    //update view on chat content and chat list
    this.initChatMsgList()
    this.initChatMsgContent()
    this.newChatFlg = true
  }

  createNewChatMsg(targetUser) {
    const newChatMsg = {
      fromName: this.loginName,
      toName: targetUser,
      pinMsg: "",
      createdt: (new Date()).getTime()/1000,
      hasContent: false,
      targetUser: targetUser,
      newMsg: true, 
    }
    //if before is new chat, replace from 0 index
    if(this.newChatFlg) {
      this.chatList[0] = newChatMsg
    } else {
      //add to first arrayList
      this.chatList.unshift(newChatMsg)
    }
    this.currentChatIndex = 0
  }

  initChatMsgContent() {
    if(!this.chatList || this.chatList.length == 0) {
      $("#currentTarget").empty()
      $("#pinMsg").empty()
      $("#chatContentArea").empty()
      $("#deleteChatBtn").addClass('d-none')
      return
    }
    $("#deleteChatBtn").removeClass('d-none')
    const chatMsg = this.chatList[this.currentChatIndex]
    $("#currentTarget").text(chatMsg.targetUser)
    //Pin msg
    $("#pinMsg").text(chatMsg.pinMsg)
    //Check chat Msg. If has unread count, update on db
    if(chatMsg.unreadNum > 0 && this.chatboxDisplay) {
      this.updateUnreadCount(chatMsg)
      if(this.unreadCount == 0) {
        $("#floatingBtn").removeAttr("data-after-type")
        $("#floatingBtn").removeAttr("data-after-text")
      }
    }
    //update chat content area
    let innerHTML = ""
    const _this = this
    if(chatMsg.chatContentList && chatMsg.chatContentList.length > 0) {
       chatMsg.chatContentList.forEach(element => {
         innerHTML += _this.createChatContentHTML(chatMsg, element)
      })
    }
    $("#chatContentArea").html(innerHTML)
    $('#chatContentArea').scrollTop($('#chatContentArea')[0].scrollHeight)
  }

  updateUnreadCount(chatMsg) {
    if(chatMsg.unreadNum <= 0) {
      return
    }
    const _this = this
    $.ajax({
      data: {
        chatId: chatMsg.id ? chatMsg.id : 0,
      },
      type: "POST", //OR GET
      url: '/updateUnread', //The same form's action URL
      success: function (data) {
        if (data["error"] == "") {
          //if update seen completed. Remove badge and update badge of chat item
          _this.chatList[_this.currentChatIndex].unreadNum = 0
          _this.initChatMsgList()
          //update unread count on floating button
          if(_this.unreadCount <= 0) {
            return
          }
          //update display count
          _this.unreadCount -= 1
          if(_this.unreadCount == 0) {
            //if unread count is zero, remove badge
            $("#floatingBtn").removeAttr("data-after-type")
            $("#floatingBtn").removeAttr("data-after-text")
            return
          }
          $("#floatingBtn").attr("data-after-type", "red badge top right")
          $("#floatingBtn").attr("data-after-text", _this.unreadCount)
        }
        if (data["error"] != "") {
          //if error, display successfully notification
          _this.showSuccessToast(data["error_msg"]);
        }
      },
    });
  }

  deleteChatBtnClick() {
    $("#deleteConfirmDialog").on("shown.bs.modal", function () {}).modal('show');
  }

  deleteChat() {
    const chatMsg = this.chatList[this.currentChatIndex]
    const _this = this
    $.ajax({
      data: {
        chatId: chatMsg.id,
      },
      type: "POST", //OR GET
      url: '/deleteChat', //The same form's action URL
      success: function (data) {
        if (data["error"] == "") {
          $("#deleteerr_msg").addClass("d-none")
          //remove from list
          _this.chatList.splice(_this.currentChatIndex, 1)
          _this.currentChatIndex = 0          
          _this.initChatMsgList()
          _this.initChatMsgContent()
          $("#deleteConfirmDialog").on("shown.bs.modal", function () {}).modal('hide');
        }
        if (data["error"] != "") {
          $("#deleteerr_msg").removeClass("d-none")
          $("#deleteerr_msg").text(data["error_msg"])
        }
      },
    });
  }

  createChatContentHTML(chatMsg, element) {
    const isTarget = element.userName == chatMsg.targetUser
    let innerHTML = ''
    if(element.isHello) {
      innerHTML += `<div class="no-gutters text-center c-b-grey">` +
      `<span class="fs-13">${this.GetChatTime(Number(element.createdt))}</span><br>` +
      `<span class="fs-14">${element.content}</span></div>`
    } else {
      innerHTML += `<div class="no-gutters d-flex justify-content-${isTarget ? 'start' : 'end'}">` +
      `<div class="chat-bubble chat-bubble--${isTarget ? 'left' : 'right'} mw-70 minw-40">` +
      `<div class="d-flex justify-content-between"><span class="fw-600 fs-14">${element.userName}</span><span class="fs-13">${this.GetChatTime(Number(element.createdt))}</span></div>${element.content}</div>` +
      `</div>`
    }
    return innerHTML
  }

  backToList() {
    this.contentShowing = false
    $("#chatMsgListCol").removeClass('d-none')
    $("#chatContentCol").addClass('d-none')
    if(this.newChatFlg) {
      //if current is new chat, remove from list
      this.chatList.splice(this.currentChatIndex, 1)
      this.newChatFlg = false
    }
    this.initChatMsgList()
    this.currentChatIndex = 0
  }

  chatMsgChange(e) {
    let afterIndex = e.params.index
    if (this.isMobileView) {
      $("#chatMsgListCol").addClass('d-none')
      $("#chatContentCol").removeClass('d-none')
      this.contentShowing = true
    }
    if(afterIndex == this.currentChatIndex) {
      return
    }
    if(this.newChatFlg) {
      //if current is new chat, remove from list
      this.chatList.splice(this.currentChatIndex, 1)
      afterIndex -= 1
      this.newChatFlg = false
    }
    this.initChatMsgList()
    //remove active class from old chat item
    if (!this.isMobileView) {
      this.changeChatItemActiveClass(this.currentChatIndex, afterIndex)
    } else {
      $("#chatItem" + this.currentChatIndex).removeClass("chat-active")
    }
    this.currentChatIndex = afterIndex
    this.initChatMsgContent()
  }

  changeChatItemActiveClass(oldIndex, newIndex) {
    $("#chatItem" + oldIndex).removeClass("chat-active")
    $("#chatItem" + newIndex).addClass("chat-active")
  }

  floatingBtnClick(e) {
    e.preventDefault();
    $("#floatingBtn").toggleClass('open');
    if($("#floatingBtn").children('.fa').hasClass('fa-comment'))
    {
        $("#floatingBtn").children('.fa').removeClass('fa-comment');
        $("#floatingBtn").children('.fa').addClass('fa-close');
    } 
    else if ($("#floatingBtn").children('.fa').hasClass('fa-close')) 
    {
        $("#floatingBtn").children('.fa').removeClass('fa-close');
        $("#floatingBtn").children('.fa').addClass('fa-comment');
    }
    const _this = this
    $('.floatingMenu').stop().slideToggle(200, function() {
          //check floating menu display or hidden
    if($("#floatingMenu").css('display') == 'none') {
      _this.chatboxDisplay = false
    } else {
      if(!_this.openedChatBox) {
        _this.openedChatBox = true
        _this.initChatMsgList()
        _this.initChatMsgContent()
        _this.handlerSearchListDisplay()
      }
        if(_this.chatList && _this.chatList.length > 0) {
          const chatMsg = _this.chatList[_this.currentChatIndex]
          if(chatMsg.unreadNum > 0) {
            _this.updateUnreadCount(chatMsg)
          }
        }
      _this.chatboxDisplay = true
      $('#chatContentArea').scrollTop($('#chatContentArea')[0].scrollHeight)
    }
    });
  }

  initChatMsgList() {
    if (!this.chatList || this.chatList.length == 0) {
      $("#chatMsgList").html("")
      return
    }
    const _this = this
    let innerHtml = ""
    this.chatList.forEach((element, index) => {
      if(element.hasContent || element.newMsg){
        const lastItem = element.lastContent
        innerHtml += `<div id="chatItem${index}" class="friend-drawer friend-drawer--onhover ${!_this.isMobileView && _this.currentChatIndex == index ? 'chat-active' : ''}" data-chat-index-param="${index}" data-action="click->chat#chatMsgChange">` +
        `<div ${element.unreadNum > 0 ? 'data-after-text="' + element.unreadNum +'" data-after-type="red badge top left"' : ''}><img class="profile-image" src="/static/images/avatar.png" alt=""></div>` +
        `<div class="text">` +
        `<div class="d-flex justify-content-between"><span class="fw-600">${element.targetUser}</span><span id="itemTime${index}" class="time text-muted small">${!lastItem ? '' : _this.GetChatTime(Number(lastItem.createdt))}</span></div>` +
        `<p class="text-muted elipse-text" id="itemLastContent${index}">${!lastItem ? '' : lastItem.content}</p>` +
        `</div>` +
        `</div>`
      }
    });
    $("#chatMsgList").html(innerHtml)
  }

  updateChatMsgContent(lastContent, index) {
    $("#itemTime" + index).text(!lastContent ? '' : this.GetChatTime(Number(lastContent.createdt)))
    $("#itemLastContent" + index).text(!lastContent ? '' : lastContent.content)
  }

  msgContentChange() {
   const content = $("#msgContent").val()
   if(!content || content == '') {
     this.canSend = false
      $("#sendButton").addClass('disabled-icon')
      $("#sendButton").removeClass('enable-icon')
      return
   }
   $("#sendButton").removeClass('disabled-icon')
   $("#sendButton").addClass('enable-icon')
   this.canSend = true
  }

  sendMsg() {
    if(!this.canSend) {
      return
    }
    //get chat content
    let chatMsg = this.chatList[this.currentChatIndex]
    if (!chatMsg) {
      return
    }
    const _this = this
    $.ajax({
      data: {
        chatId: chatMsg.id ? chatMsg.id : 0,
        fromUser: chatMsg.fromName,
        toName: chatMsg.toName,
        newMsg: $("#msgContent").val(),
      },
      type: "POST", //OR GET
      url: '/sendChatMessage', //The same form's action URL
      success: function (data) {
        if (data["error"] == "") {
          $("#chaterr_msg").addClass("d-none")
          const result = data["result"]
          _this.newChatFlg = false
          const jsonResult = JSON.parse(result)
          if(!jsonResult){
            return
          }
          const msgContent = jsonResult.newContent
          //get newChatMsgObject
          if(jsonResult.newMsg && jsonResult.newMsg.id > 0) {
            //replace with current new ChatMsg
            _this.chatList[_this.currentChatIndex] = jsonResult.newMsg
            _this.initChatMsgList()
            chatMsg = jsonResult.newMsg
          } else {
            chatMsg.chatContentList.push(msgContent)
            chatMsg.lastContent = msgContent
            _this.chatList[_this.currentChatIndex] = chatMsg
          }
          _this.updateChatMsgContent(msgContent, _this.currentChatIndex)
          //if index of chat difference with 0, swap to zero
          if(_this.currentChatIndex > 0) {
            //backup chatMsg
            const tempMsg = _this.chatList[_this.currentChatIndex]
            //remove from index
            _this.chatList.splice(_this.currentChatIndex, 1)
            //insert to the first of array
            _this.chatList.unshift(tempMsg)
            _this.currentChatIndex = 0
            //update msg list
            _this.initChatMsgList()
          }
          //Append to chat content
          $("#chatContentArea").append(_this.createChatContentHTML(chatMsg, msgContent))
          //reset input field
          $("#msgContent").val('')
          _this.msgContentChange()
          //scroll down to div
          $('#chatContentArea').scrollTop($('#chatContentArea')[0].scrollHeight)
          //Send to socket
          _this.postMessage(JSON.stringify(msgContent))
        }
        if (data["error"] != "") {
          $("#chaterr_msg").removeClass("d-none")
          $("#chaterr_msg").text(data["error_msg"])
        }
      },
    });
  }
}
