import BaseController from "./base_controller";


let isClickOutContactList = false
let isWalletDetailPage = false

$(document).mouseup(function(e)
{
    var container = $("#contactList");
    // if the target of the click isn't the container nor a descendant of the container
    isClickOutContactList = !container.is(e.target) && container.has(e.target).length === 0
    if(isClickOutContactList && isWalletDetailPage) {
      handlerUsernameSearchList("")
    }
});

function handlerUsernameSearchList(receivername) {
  var filter, ul, li, a, i, txtValue;
  filter = receivername.toUpperCase();
  ul = document.getElementById("contactList");
  li = ul.getElementsByTagName("li");
  let hasKey = false
  for (i = 0; i < li.length; i++) {
      a = li[i].getElementsByTagName("a")[0];
      txtValue = a.textContent || a.innerText;
      if (filter && txtValue.toUpperCase().indexOf(filter) > -1) {
          li[i].style.display = "";
          hasKey = true
      } else {
          li[i].style.display = "none";
      }
  }
  return hasKey
}

export default class extends BaseController {
  static values = {
    balance: Number,
    assetType: String,
    assetId: Number,
    hasAddress: Boolean,
    currentAddress: String,
    defaultQueryParams: Object,
    queryParams: Object,
    currentTab: String,
    rate: Number,
    addressError: Boolean,
    amountError: Boolean,
    usernameError: Boolean,
    tradingAmountError: Boolean,
    paymentBalanceError: Boolean,
    btcFee: Number,
    confirmed: Boolean,
    sendBy: String,
    confirmed: Boolean,
    currentCancelCodeId: String,
    codeFilterVal: String,
    addressStatusFilter: String,
    addToContact: Boolean,
    paymentType: String,
    paymentRate: Number,
    tradingType: String,
    currentPaymentBalance: Number,
    assetRound: Number,
    confirmDialogType: String,
    currentAddressId: Number,
    addressAction: String,
  };

  static get targets() {
    return ["transferForm", "addressListTable"];
  }

  async initialize() {
    this.balance = parseFloat(this.data.get("balance"));
    this.assetType = this.data.get("assetType")
    this.assetRound = this.getRoundNumber(this.assetType)
    this.assetId = Number(this.data.get("assetId"))
    this.hasAddress = this.data.get("hasAddress") == "true"
    this.paymentType = this.data.get("paymentFirsttype")
    this.currentPaymentBalance = parseFloat(this.data.get("paymentFirstbalance"))
    this.tradingType = "buy"
    const _this = this
    //if has address, display address area
    if(this.hasAddress) {
      $("#qrAddressArea").removeClass('d-none')
      this.displayAddressArea($("#addressSelector").val())
      this.currentAddress = $("#addressSelector").val()
    }
    this.defaultQueryParams = {
      tab: "info",
      type: "",
      cstatus: "all",
      astatus: "active"
    };
    this.queryParams = this.getParamsFromUrl();
    if(!this.queryParams.tab || this.queryParams.tab == ""){
      this.queryParams.tab = this.defaultQueryParams.tab
    }

    if(!this.queryParams.cstatus || this.queryParams.cstatus == ""){
      this.queryParams.cstatus = this.defaultQueryParams.cstatus
      this.codeFilterVal = this.defaultQueryParams.cstatus
    } else {
      this.codeFilterVal = this.queryParams.cstatus
    }
    
    if(!this.queryParams.astatus || this.queryParams.astatus == ""){
      this.queryParams.astatus = this.defaultQueryParams.astatus
      this.addressStatusFilter = this.defaultQueryParams.astatus
    } else {
      this.addressStatusFilter = this.queryParams.astatus
    }
    $("#addressFilter").val(this.addressStatusFilter)
    this.currentTab = this.queryParams.tab
    this.activaTab(this.queryParams.tab)
    _this.transferFormTarget.addEventListener("submit", (e) => {
      e.preventDefault();
      if ($("#amountSend").val() == 0) {
        $("#amounterr_msg").removeClass("d-none")
        $("#amounterr_msg").text("The transfer amount must be greater than 0")
        $("#transferButton").prop("disabled", true)
        return false
      }
      //if is send to address and not confirmed fee. Return error
      if(this.assetType != "usd" && (this.sendBy == "address" || this.sendBy == "urlcode") && !this.confirmed) {
        $("#amounterr_msg").removeClass("d-none")
        $("#amounterr_msg").text("Please confirm the transaction before sending")
        $("#transferButton").prop("disabled", true)
        return false
      }
      const validSubmit = _this.checkValidTransferButton()
      if(!validSubmit) {
        return false
      }
      _this.submitTransferForm()
      return false;
    });
  this.sendBy = "username"
  this.amountError = false
  this.rate = 0.0
  $("input[type=radio][name=sendby]").change(function () {
    _this.sendBy = this.value
    $("#amountSend").val(0)
     _this.confirmed = false
     _this.btcFee = 0
     $("#btcFee").text(0)
     $("#btcCost").text(0)
     $("#feeExchange").text(0)
     $("#costExchange").text(0)
     $("#sendAmountExchange").text(0)
     _this.handlerAssetType()
    _this.handlerWalletDisplay()
    _this.checkValidTransferButton()
  });
  if (this.assetType !== 'usd') {
    this.updateExchangeRate()
  }
  switch(this.assetType) {
    case "btc":
      $("#amountSend").attr('step', '0.00000001');
      break
    case "ltc":
      $("#amountSend").attr('step', '0.0001');
      break
    case "dcr":
      $("#amountSend").attr('step', '0.001');
      break
    default:
      $("#amountSend").attr('step', '0.01');
  }
  this.handlerWalletDisplay()
  this.handlerAssetType()
  this.checkValidTransferButton()
  $("#urlSample").text(window.location.origin + "/withdrawl?code={code}")
  this.loadCodeList()
  this.addToContact = false
  handlerUsernameSearchList($("#receiver").val())
  isWalletDetailPage = true
  $("input[type=radio][name=tradingType]").change(function () {
    _this.tradingType = this.value
    if(_this.tradingType == "sell") {
      $("#paymentMethodTitle").text("Receive Asset")
      $("#assetSelectTitle").text("Asset")
      $("#haveToPayLabel").text("Will Receive")
    } else {
      $("#paymentMethodTitle").text("Payment Method")
      $("#assetSelectTitle").text("Asset to payment")
      $("#haveToPayLabel").text("Have to pay")
    }
    _this.amountTradingChange()
    _this.updateRateToDisplay()
  });
  if (this.assetType !== 'usd') {
    this.handlerPaymentTypeInfo()
    this.loadAddressList()
  }
  // //init host name
  // $("#host1").text(window.location.origin)
  // $("#host2").text(window.location.origin)

  // const postWithdrawToAccLink = "/withdrawl?token=" + this.token +"&asset=" + this.assetType + "&account=username&amount=0.001"
  // const postWithdrawToAddrLink = "/withdrawl?token=" + this.token +"&asset=" + this.assetType + "&address=toaddress&amount=0.001"
  // //Create withdraw shorten link
  // $("#postLink1").text(shortenString(postWithdrawToAccLink, 50))
  // $("#postLink2").text(shortenString(postWithdrawToAddrLink, 50))
  }

  amountTradingChange() {
    //check balance
    const amount = $("#amountTrading").val()
    if(!amount || amount == "") {
      $("#tradingAmountErr_msg").removeClass("d-none")
      $("#tradingAmountErr_msg").text("Trading amount cannot be left blank")
      this.tradingAmountError = true
      return
    }
    if(this.tradingType == "sell" && Number(amount) > this.balance) {
      $("#tradingAmountErr_msg").removeClass("d-none")
      $("#tradingAmountErr_msg").text("The balance is not enough to make this trading")
      this.tradingAmountError = true
      return
    }
    $("#tradingAmountErr_msg").addClass("d-none")
    this.tradingAmountError = false
    //update balance after trading
    $("#tradingAfterBalance").text(formatToLocalString(this.tradingType == "sell" ? this.balance - Number(amount) : this.balance + Number(amount), this.assetRound, this.assetRound))
    this.updateRateToDisplay()
    this.hanlderSendTradingRequestButton()
  }

  paymentTypeChange(e) {
    if(e.target.checked) {
      this.paymentType = e.target.value
      this.currentPaymentBalance = Number(e.params.balance)
      this.handlerPaymentTypeInfo()
      this.updateRateToDisplay()
    }
  }

  handlerPaymentTypeInfo() {
    const roundNumber = this.getRoundNumber(this.paymentType)
    $("#paymentAssetBalance").text(formatToLocalString(this.currentPaymentBalance, roundNumber, roundNumber))
    $("#balanceCurrency").text(this.paymentType.toUpperCase())
    $("#needCurrency").text(this.paymentType.toUpperCase())
    $("#paymentAfterCurrency").text(this.paymentType.toUpperCase())
    $("#exchangeSymbol").text(this.paymentType.toUpperCase() + "/" + this.assetType.toUpperCase())
  }

  cancelBtcClick(e) {
    this.currentCancelCodeId = e.params.codeid
    this.confirmDialogType = "code"
    this.showCancelDialog()
    $("#urlcodeCancelDialog").on("shown.bs.modal", function () {}).modal('show');
  }

  archiveClick(e) {
    this.currentAddressId = e.params.addrid
    this.confirmDialogType = "address"
    const archived = e.params.archived
    this.addressAction = e.params.archived ? "reuse" : "archive"
    $("#urlcodeCancelDialog").on("shown.bs.modal", function () {}).modal('show');
    this.showArchiveAddressDialog(archived)
  }

  showCancelDialog() {
    $("#warningMsg").addClass("d-none")
    $("#urlcodeCancelBackdropLabel").text("Confirm Cancel Withdraw Code")
    $("#confirmContent").text("Are you sure you want to cancel the withdrawal code?")
  }

  showArchiveAddressDialog(archived) {
    if(!archived) {
      $("#warningMsg").removeClass("d-none")
    } else {
      $("#warningMsg").addClass("d-none")
    }
    $("#urlcodeCancelBackdropLabel").text(archived ? "Address reuse confirmation" : "Address archive confirmation")
    $("#confirmContent").text("Are you sure you want to " + (archived ? "reuse" : "archive")+ " this address?")
  }

  confirmAction() {
    if(this.confirmDialogType == "code") {
      this.cancelTxCode()
      return
    }
    if(this.confirmDialogType == "address") {
      this.hanlderArchiveAddress()
    }
  }

  hanlderArchiveAddress() {
    if(!this.currentAddressId || this.currentAddressId <= 0 || !this.addressAction || this.addressAction == "") {
      return
    }
    const _this = this
    $.ajax({
      data: {
        assetId: _this.assetId,
        addressId: _this.currentAddressId,
        action: _this.addressAction
      },
      type: "POST", //OR GET
      url: '/confirmAddressAction', //The same form's action URL
      success: function (res) {
        if (!res.error) {
          window.location.reload()
        } else {
          $("#cancelerr_msg").removeClass("d-none")
          $("#cancelerr_msg").text(res.msg)
        }
      },
    });
  }

  cancelTxCode() {
    if(!this.currentCancelCodeId || this.currentCancelCodeId == "") {
      return
    }
    const _this = this
    $.ajax({
      data: {
        codeId: _this.currentCancelCodeId,
      },
      type: "POST", //OR GET
      url: '/cancelUrlCode', //The same form's action URL
      success: function (res) {
        if (!res.error) {
          window.location.reload()
        } else {
          $("#cancelerr_msg").removeClass("d-none")
          $("#cancelerr_msg").text(res.msg)
        }
      },
    });
  }

  handlerAssetType() {
    $("#amountSymbol").text(this.assetType.toUpperCase())
    const roundNumber = this.getRoundNumber(this.assetType)
    let amountDisp
    if (this.assetType == "usd") {
      amountDisp = "$" + formatToLocalString(this.balance, 2, 2)
      $("#sendAmountExchangeArea").addClass("d-none");
      $("#balanceAffterRateArea").addClass("d-none");
    } else {
      amountDisp = formatToLocalString(this.balance, roundNumber, roundNumber) + " " + this.assetType.toUpperCase()
      $("#sendAmountExchangeArea").removeClass("d-none");
      $("#balanceAffterRateArea").removeClass("d-none");
    }
    $("#balanceAfter").text(amountDisp);
  }

  receivingAddressChange(e) {
    if(this.assetType != "usd" && $("#toAddress").val() && $("#toAddress").val() != '' && Number($("#amountSend").val()) > 0) {
      $("#btcConfirmBtn").removeClass('d-none')
      this.confirmed = false
    } else {
      $("#btcConfirmBtn").addClass('d-none')
    }
    this.checkValidTransferButton()
  }

  handlerWalletDisplay() {
    if (this.assetType == "usd") {
      $("#sendbySelector").addClass("d-none")
      $("#receiveAddressInput").addClass("d-none")
      $("#receiverUserNameSelector").removeClass("d-none")
    } else {
      $("#sendbySelector").removeClass("d-none")
      //if withdraw by url code
      if(this.sendBy == "urlcode") {
        $("#receiverUserNameSelector").addClass("d-none")
        $("#receiveAddressInput").addClass("d-none")
        //change transfer button name
        $("#transferButton").html("Create Withdraw Code")
      //if send by address
      } else {
        $("#transferButton").html("Transfer")
        if (this.sendBy == "address") {
          $("#receiverUserNameSelector").addClass("d-none")
          $("#receiveAddressInput").removeClass("d-none")
        //if send by username
        } else {
          $("#receiverUserNameSelector").removeClass("d-none")
          $("#receiveAddressInput").addClass("d-none")
        }
      }
    }
  }

  amountInputChange(e) {
    const sendAmount = Number($("#amountSend").val());
    const _this = this
    if (sendAmount > this.balance) {
      $("#amounterr_msg").removeClass("d-none");
      $("#transferButton").prop("disabled", true);
      $("#amounterr_msg").text(
        "The transfer amount cannot be greater than the balance"
      );
      this.amountError = true
      return;
    }
    this.amountError = false
    $("#transferButton").prop("disabled", false);
    if (!$("#amounterr_msg").hasClass("d-none")) {
      $("#amounterr_msg").addClass("d-none");
    }
    if(this.assetType != "usd") {
      let sendAmoutExchange = this.rate * sendAmount;
      let afterAmountExchange = this.rate * (this.balance - sendAmount);
      $("#sendAmountExchange").text(formatToLocalString(sendAmoutExchange, 2, 2));
      $("#balanceAfterExchange").text(formatToLocalString(afterAmountExchange, 2, 2));
    }
    let balanceAfterDisp = ''
    const diff = this.getRoundNumber(this.assetType)
    const afterAmount = formatToLocalString(this.balance - sendAmount, diff, diff);
      balanceAfterDisp = this.assetType == "usd" ? "$" + afterAmount : afterAmount + " " + this.assetType.toUpperCase()
    $("#balanceAfter").text(balanceAfterDisp);
    $("#btcCost").text(formatToLocalString(sendAmount, diff, diff));
    if(this.assetType != "usd" && (this.sendBy == "address" || this.sendBy == "urlcode") && sendAmount > 0) {
      $("#btcConfirmBtn").removeClass('d-none')
      this.confirmed = false
    } else {
      $("#btcConfirmBtn").addClass('d-none')
    }
    this.checkValidTransferButton()
  }

  usernameChange2(e) {
    $("#addToContactArea").addClass("d-none")
    $("#usernameErrMsg").addClass("d-none")
    const username = $("#receiver").val()
    const hasItem = handlerUsernameSearchList(username)
    if(hasItem) {
      return
    }
    if(!username || username == "") {
      $("#usernameErrMsg").removeClass("d-none")
      $("#usernameErrMsg").text("Recipient cannot be left blank")
      $("#transferButton").prop("disabled", true);
      this.usernameError = true
      return
    }
  }

  usernameChange(e) {
    $("#addToContactArea").addClass("d-none")
    $("#usernameErrMsg").addClass("d-none")
    const username = $("#receiver").val()
    handlerUsernameSearchList(username)
    if(!username || username == "") {
      $("#usernameErrMsg").removeClass("d-none")
      $("#usernameErrMsg").text("Recipient cannot be left blank")
      $("#transferButton").prop("disabled", true);
      this.usernameError = true
      return
    }
    $("#usernameErrMsg").addClass("d-none")
    const _this = this
      //check user exist
      $.ajax({
        data: {
          username: username,
        },
        type: "GET", //OR GET
        url: '/check-contact-user', //The same form's action URL
        success: function (res) {
          if (!res.error && res.data.exist) {
            const contactExist = res.data.contactExist
            _this.usernameError = false
            $("#addToContactArea").removeClass("d-none")
            //check result
            $("#addToContactCheckbox").prop('checked', contactExist)
          } else {
            $("#usernameErrMsg").removeClass("d-none")
            $("#usernameErrMsg").text(res.msg)
            $("#transferButton").prop("disabled", true);
            _this.usernameError = true
          }
          _this.checkValidTransferButton()
        },
      });
  }

  usernameBlur() {
    const username = $("#receiver").val()
    handlerUsernameSearchList("")
    if(!username || username == "") {
      $("#usernameErrMsg").removeClass("d-none")
      $("#usernameErrMsg").text("Recipient cannot be left blank")
      $("#transferButton").prop("disabled", true);
      this.usernameError = true
      return
    }
    $("#usernameErrMsg").addClass("d-none")
    const _this = this
      //check user exist
      $.ajax({
        data: {
          username: username,
        },
        type: "GET", //OR GET
        url: '/check-contact-user', //The same form's action URL
        success: function (res) {
          if (!res.error && res.data.exist) {
            const contactExist = res.data.contactExist
            _this.usernameError = false
            $("#addToContactArea").removeClass("d-none")
            //check result
            $("#addToContactCheckbox").prop('checked', contactExist)
          } else {
            $("#usernameErrMsg").removeClass("d-none")
            $("#usernameErrMsg").text(res.msg)
            $("#transferButton").prop("disabled", true);
            _this.usernameError = true
          }
          _this.checkValidTransferButton()
        },
      });
  }

  addToContactValueChange(e) {
    this.addToContact = e.target.checked
  }

  checkValidTransferButton() {
    $("#urlcodeFeeNotify").addClass("d-none")
    $("#existAddressSystem").addClass('d-none')
    $("#amounterr_msg").addClass('d-none')
    if(this.assetType != "usd" && (this.sendBy == "address") || this.sendBy == "urlcode") {
      $("#feeArea").removeClass('d-none')
    } else {
      $("#feeArea").addClass('d-none')
    }
    const sendAmount = Number($("#amountSend").val());
    if(this.assetType != "usd" && this.sendBy == "address") {
        const address = $("#toAddress").val()
        this.addressError = !address || address == ""
        if(this.addressError) {
          $("#addresserror_msg").removeClass("d-none")
        } else {
          $("#addresserror_msg").addClass("d-none")
        }
    }

    //check transfer button
    let disableButton = false
    if(this.assetType != "usd") {
      if(this.sendBy == "address" && this.confirmed) {
          //if cost more than balance, return error
          if(Number(this.btcFee) + sendAmount > this.balance) {
            $("#confirmAmountErr_msg").removeClass('d-none')
            this.amountError = true
          } else {
            $("#confirmAmountErr_msg").addClass('d-none')
          }
      } else {
        $("#confirmAmountErr_msg").addClass('d-none')
      }
      if(this.sendBy == "username") {
        disableButton = this.amountError || this.usernameError
      } else if (this.sendBy == "urlcode") {
        disableButton = this.amountError
      } else {
        disableButton = this.amountError || this.addressError
      }
    } else {
      disableButton = this.amountError || this.usernameError
    }
    $("#transferButton").prop("disabled", disableButton)
    return !disableButton
  }

  confirmCryptoAmount() {
    const amount = Number($("#amountSend").val())
    if(amount == 0 || this.walletType == 'usd') {
      return
    }
    const _this = this
    $.ajax({
      data: {
        amount: amount,
        asset: _this.assetType,
        sendBy: _this.sendBy,
        toaddress: $("#toAddress").val()
      },
      type: "POST", //OR GET
      url: '/confirmAmount', //The same form's action URL
      success: function (res) {
        const diff = _this.getRoundNumber(_this.assetType)
        if (!res.error) {
          $("#amounterr_msg").addClass("d-none")
          const result = JSON.parse(res.data)
          _this.btcFee = Number(result.fee)
          _this.confirmed = true
          _this.checkValidTransferButton()
          let balanceAfterDisp = ''
          const afterMountFloat = _this.balance - (_this.btcFee + amount)
          if(afterMountFloat < 0) {
            //if after amount less than 0, show error
            $("#amounterr_msg").removeClass("d-none")
            $("#amounterr_msg").text("Total Cost can't not greater than balance")
            $("#transferButton").prop("disabled", true);
          }
          const afterAmount = formatToLocalString( _this.balance - (_this.btcFee + amount), diff, diff);
          balanceAfterDisp =  _this.assetType == "usd" ? "$" + afterAmount : afterAmount + " " + _this.assetType.toUpperCase()
          $("#balanceAfter").text(balanceAfterDisp);
          //display urlcode fee notify
          if(_this.sendBy == "urlcode") {
            $("#urlcodeFeeNotify").removeClass("d-none")
            $("#urlcodeFeeNotify").text("*You'll not be charged fee if the recipient chooses to withdraw using system username")
          }
        }
        if (res.error && res.code == "exist") {
          _this.btcFee = 0.0
          _this.confirmed = true
          $("#existAddressSystem").removeClass('d-none')
          $("#existAddressSystem").html(res.msg)
        }

        if (!res.error || (res.error && res.code == "exist")) {
          $("#btcFee").text(formatToLocalString(Number(_this.btcFee), diff, diff))
          $("#btcCost").text(formatToLocalString(Number(_this.btcFee) + amount, diff, diff))
          $("#feeExchange").text(formatToLocalString(Number(_this.btcFee) * _this.rate, 3, 3))
          $("#costExchange").text(formatToLocalString((Number(_this.btcFee) + amount) * _this.rate, 3, 3))
          $("#btcConfirmBtn").addClass('d-none')
          return
        }

        if (res.error) {
          $("#amounterr_msg").removeClass("d-none")
          $("#amounterr_msg").text(res.msg)
          $("#transferButton").prop("disabled", true);
        }
      },
    });
  }

  selectContact(e) {
    const selectedUsername = e.params.username
    $("#receiver").val(selectedUsername)
    this.usernameBlur()
  }

  tradingSendRequest() {
    if( this.tradingAmountError || this.paymentBalanceError) {
      return
    }
    const amount = $("#amountTrading").val()
    $.ajax({
      data:  {
        asset: this.assetType,
        tradingType: this.tradingType,
        amount: amount,
        paymentType: this.paymentType,
        rate: this.paymentRate },
      type: "POST", //OR GET
      url: "/send-trading-request", //The same form's action URL
      success: function (res) {
        if (!res.error) {
          window.location.href = "/";
          return;
        } else {
          $("#tradingSendingErr_msg").removeClass("d-none")
          $("#tradingSendingErr_msg").text(res.msg)
        }
      },
    });
  }

  submitTransferForm() {
    const url = '/transfer-amount'
    const username = $("#receiver").val()
    const address = $("#toAddress").val()
    const _this = this
    let data = {
        receiver: username,
        address: address,
        sendBy: this.sendBy,
        currency: this.assetType,
        amount: $("#amountSend").val(),
        rate: this.rate,
        note: $("#noteText").val(),
        addToContact: this.addToContact,
      }
    $.ajax({
      data: data,
      type: "POST", //OR GET
      url: url, //The same form's action URL
      success: function (res) {
        if (!res.error) {
          if(_this.sendBy == "urlcode") {
            //return detail page with URL Code tab
            window.location.href = "/assets/detail?type=" + _this.assetType + "&tab=urlcode"
            return
          }
          window.location.href = "/";
          return;
        } else {
          if(_this.assetType == "usd") {
            $("#resulterr_msg").removeClass("d-none");
            $("#resulterr_msg").text(res.msg);
          } else {
            $("#confirmErr").removeClass("d-none");
            $("#confirmErr").text(res.msg);
          }
        } 
      },
    });
  }

  activaTab(tab){
    $("#" + tab + "-tab")[0].click()
  }

  tabChange(e) {
   const tabId = e.target.id
   if (!tabId || tabId == "") {
    return
   }
   const tabName = tabId.replace("-tab", "")
   if(this.currentTab == tabName) {
    return
   }
   this.currentTab = tabName
   this.queryParams.tab = tabName
   this.updateQueryUrl(this.queryParams, this.defaultQueryParams)
  }

  addressSelectChange() {
    this.displayAddressArea($("#addressSelector").val())
    this.currentAddress = $("#addressSelector").val()
  }

  loadCodeList() {
 //get filter code
 const _this = this
 $.ajax({
   url: "/GetCodeListData",
   data: {
     asset: _this.assetType,
     codeStatus: _this.codeFilterVal,
   },
   type: "GET",
   success: function (data) {
     if (!data) {
       return
     }
     const dataList = data.list
     //create history list
     $("#codeListTable").html(_this.createCodeListTable(dataList))
   },
 });
  }

  filterCodeStatus(e) {
    this.codeFilterVal = e.target.value
    this.queryParams.cstatus = e.target.value
    this.updateQueryUrl(this.queryParams, this.defaultQueryParams)
    this.loadCodeList()
  }

  filterAddresses(e) {
    this.addressStatusFilter = e.target.value
    this.queryParams.cstatus = e.target.value
    this.updateQueryUrl(this.queryParams, this.defaultQueryParams)
    this.loadAddressList()
  }

  loadAddressList() {
    //get filter code
    const _this = this
    $.ajax({
      url: "/GetAddressListData",
      data: {
        assetId: _this.assetId,
        status: _this.addressStatusFilter,
      },
      type: "GET",
      success: function (data) {
        if (!data) {
          return
        }
        const dataList = data.list
        const userToken = data.userToken
        //create address list
        _this.createAddressListTable(dataList, userToken)
      },
    });
     }
 
  labelClick(e) {
    const id = e.params.id
    const oldMainLabel = e.params.mainlabel
    //hide label text
    $("#" + id + "LabelText").addClass("d-none")
    $("#" + id + "EditLabelIcon").addClass("d-none")
    //display label input
    $("#" + id + "LabelUserToken").removeClass("d-none")
    $("#" + id + "LabelInput").removeClass("d-none")
    $("#" + id + "LabelInput").val(oldMainLabel)
    $("#" + id + "LabelInput").focus()
  }

  cancelLabelEdit(id, oldValue) {
    if(!oldValue || oldValue == "") {
      $("#" + id + "LabelUserToken").addClass("d-none")
      $("#" + id + "EditLabelIcon").removeClass("d-none")
    } else {
      $("#" + id + "LabelUserToken").removeClass("d-none")
      $("#" + id + "EditLabelIcon").addClass("d-none")
    }
    $("#" + id + "LabelText").removeClass("d-none")
    $("#" + id + "LabelInput").addClass("d-none")
  }

  labelInputBlur(e) {
    const id = e.params.id
    const mainLabel = $("#" + id + "LabelInput").val()
    const oldMainLabel = e.params.mainlabel
    if(!mainLabel || mainLabel == "") {
      this.cancelLabelEdit(id, oldMainLabel)
      return
    }
    if(mainLabel == oldMainLabel) {
      this.cancelLabelEdit(id, oldMainLabel)
      return
    }
    $("#" + id + "LabelErr_msg").addClass("d-none")
    const _this = this
    //update main label
    $.ajax({
      data: {
        assetId: _this.assetId,
        assetType: _this.assetType,
        addressId: id,
        newMainLabel: mainLabel
      },
      type: "POST", //OR GET
      url: '/updateNewLabel', //The same form's action URL
      success: function (res) {
        if (!res.error) {
          $("#" + id + "LabelInput").addClass("d-none")
          $("#" + id + "LabelText").removeClass("d-none")
          $("#" + id + "LabelText").text(mainLabel)
          return;
        } else {
          $("#" + id + "LabelErr_msg").removeClass("d-none");
          $("#" + id + "LabelErr_msg").text(res.msg);
        }
      },
    });

  }

  createAddressListTable(dataList, userToken) {
    let inner = ''
    if (!dataList || dataList.length < 1) {
      inner = '<tr class="urlcode-table-row"><td colspan="100%" class="text-center"><p>Data does not exist</p></td></tr>'
    }
    const _this = this
    dataList.forEach(element => {
      inner += `<tr role="alert" class="urlcode-table-row">` +
      `<td><span class="fs-16 ms-2 cursor-pointer c-primary code-link" data-action="click->wallet#copyAddressParam" data-wallet-id-param="${element.id}" data-wallet-text-param="${element.address}">${element.address}</span>` +
      `<span class="material-symbols-outlined fs-17i ms-1 cursor-pointer" id="${element.id}AddressCopy" data-wallet-id-param="${element.id}" data-wallet-text-param="${element.address}" data-action="click->wallet#copyAddressParam">content_copy</span>` +
      `<span class="c-green fs-15 d-none" id="${element.id}AddressCopiedSpan">Copied</span></td>`
      if(_this.assetType != "dcr"){
      inner += `<td class="d-flex ai-center justify-content-center">`
      if(!element.archived) {
      inner += `<span id="${element.id}LabelUserToken" class="fs-14 ms-2 ${!element.labelMainPart || element.labelMainPart == "" ? 'd-none' : ''}">${userToken}_</span>` +
      `<span data-action="click->wallet#labelClick" data-wallet-mainlabel-param="${element.labelMainPart}" data-wallet-id-param="${element.id}" id="${element.id}EditLabelIcon" class="${!element.labelMainPart || element.labelMainPart == "" ? '' : 'd-none'} cursor-pointer material-symbols-outlined">edit_square</span>` +
      `<a data-action="click->wallet#labelClick" data-wallet-mainlabel-param="${element.labelMainPart}" data-wallet-id-param="${element.id}" id="${element.id}LabelText" class="cursor-pointer t-decor-none c-primary">${element.labelMainPart}</a>` +
      `<input id="${element.id}LabelInput" value="${element.labelMainPart}" data-wallet-mainlabel-param="${element.labelMainPart}" data-wallet-id-param="${element.id}" data-action="blur->wallet#labelInputBlur" autocomplete="off" type="text" class="form-control w-75 d-none">` +
      `<br><span class="error d-none" id="${element.id}LabelErr_msg"></span>`
    }
      inner += `</td>`
    }
     inner += `<td class="text-center"><span class="px-2">${formatToLocalString(element.totalReceived, _this.roundNumber, _this.roundNumber)} ${_this.assetType.toUpperCase()}</span></td>` +
     `<td class="text-center"><span class="px-2">${element.transactions}</span></td>` +
     `<td class="text-center"><span class="fs-13 px-2 py-1 sent-receive-label text-white" style="background-color: ${_this.addressStatusColor(element.archived)};">${element.archived ? 'Archived' : 'Active'}</span></td>` +
     `<td class="text-center">`
     inner += `<a class="btn ${element.archived ? 'btn-warning' : 'btn-primary'} small-btn" data-wallet-addrid-param="${element.id}" data-wallet-archived-param="${element.archived}" data-action="click->wallet#archiveClick">${element.archived ? 'Reuse' : 'Archive'}</a>`
     inner += `</td></tr>`
    });
    this.addressListTableTarget.innerHTML = inner
  }

  createCodeListTable(dataList) {
    if (!dataList || dataList.length < 1) {
      return '<tr class="urlcode-table-row"><td colspan="100%" class="text-center"><p>Data does not exist</p></td></tr>'
    }
    let inner = ''
    const _this = this
    dataList.forEach(element => {
      inner += `<tr role="alert" class="urlcode-table-row">` +
      `<td><span class="fs-16 ms-2 cursor-pointer c-primary code-link" data-action="click->wallet#copyCode" data-wallet-id-param="${element.id}" data-wallet-code-param="${element.code}">${element.code}</span>` +
      `<span class="material-symbols-outlined fs-17i ms-1 cursor-pointer" id="${element.id}CodeCopy" data-wallet-id-param="${element.id}" data-wallet-code-param="${element.code}" data-action="click->wallet#copyCode">content_copy</span>` +
      `<span class="c-green fs-15 d-none" id="${element.id}CopiedSpan">Copied</span></td><td><span class="fs-16 ms-2 fw-600">${element.amount} ${element.asset.toUpperCase()}</span>` +
     `</td><td><span class="fs-13 px-2 py-1 sent-receive-label text-white" style="background-color: ${_this.codeStatusColor(element.status)};">${element.statusDisplay}</span></td><td>` +
      `<span class="fs-16 ms-2">${element.createdtDisplay}</span></td><td>`
      if (element.isCreatedt) {
        inner += `<a class="btn btn-outline-light small-btn" data-wallet-codeid-param="${element.id}" data-action="click->wallet#cancelBtcClick">Cancel</a>`
      }
      if (element.isConfirmed) {
        inner += `<a class="btn btn-primary small-btn" data-action="click->wallet#collapseHandler" data-wallet-id-param="` + element.id + `">Detail</a>`
      }
      inner += `</td></tr>`
      if (element.isConfirmed) {
        inner += `<tr class="collapse code-collapse-row urlcode-table-row" id="` +element.id + `CollapseExample"><td colspan="100%" class="p-3 pt-0"><div class="table-border p-2">`
        if(element.txHistory && element.txHistory.id > 0) {
          //if txid empty, is username withdraw
          if(!element.txid || element.txid == "") {
            inner += `<p>Confirmed User: ${element.txHistory.receiver}</p>`
          } else {
            inner += `<p>Withdraw to: ${element.txHistory.toAddress}</p>`
            inner += `<p>Txid: ${element.txid}</p>`
          }
          inner += `<p>Rate: $${formatToLocalString(element.txHistory.rate, 2, 2)}</p>`
          inner += `<p>Confirmed Date: ${element.confirmdtDisplay}</p>`
          inner += `<a href="/transaction/detail?id=`+ element.historyId +`">Tx History Detail</a>`
        }
        inner += `</div></td></tr>`
      }
    });
    return inner
  }

  collapseHandler(e) {
    const id = e.params.id
    $("#" + id + "CollapseExample").toggle(200, function(){
    //hide all collapse of other row
    $('.code-collapse-row').each(
      function(index, element) {
        if((id + "CollapseExample") != element.id) {
          if($("#" + element.id).css('display') != 'none') {
            $("#" + element.id).css('display', 'none')
          }
        }
      }
    )
    })
  }

  codeStatusColor(status) {
    switch(status) {
      //unconfirmed
      case 0:
        return "#dbc272"
      //confimed
      case 1:
        return "#008000"
      //cancelled
      case 2:
        return "#dc3545"
      default:
        return ""
    }
  }

  addressStatusColor(archived) {
    return archived ? "#8e8f8e" : "#008000"
  }

  copyURL() {
    this.copyText(window.location.origin + "/withdrawl?code={code}")
    $("#urlCopy").text('done')
    $("#urlCopy").addClass('c-green')
    $("#urlCopiedSpan").removeClass('d-none')
    setInterval(function(){
      $("#urlCopy").text('content_copy')
      $("#urlCopy").removeClass('c-green')
      $("#urlCopiedSpan").addClass('d-none')
    },1500);
  }

  copyCode(e) {
    const code = e.params.code
    const id = e.params.id
    this.copyText(code)
    $("#" + id + "CodeCopy").text('done')
    $("#" + id + "CodeCopy").addClass('c-green')
    $("#" + id + "CopiedSpan").removeClass('d-none')
    setInterval(function(){
      $("#" + id + "CodeCopy").text('content_copy')
      $("#" + id + "CodeCopy").removeClass('c-green')
      $("#" + id + "CopiedSpan").addClass('d-none')
    },1500);
  }

  copyAddressParam(e) {
    const text = e.params.text
    const id = e.params.id
    this.copyText(text)
    $("#" + id + "AddressCopy").text('done')
    $("#" + id + "AddressCopy").addClass('c-green')
    $("#" + id + "AddressCopiedSpan").removeClass('d-none')
    setInterval(function(){
      $("#" + id + "AddressCopy").text('content_copy')
      $("#" + id + "AddressCopy").removeClass('c-green')
      $("#" + id + "AddressCopiedSpan").addClass('d-none')
    },1500);
  }

  copyAccountLink(e) {
    const token = e.params.token
    this.copyText(window.location.origin +"/withdrawl?token=" + token + "&asset=" + this.assetType + "&account=username&amount=0.001")
    $("#copyBtn1").text('done')
    $("#copyBtn1").addClass('c-green')
    $("#copiedSpan1").removeClass('d-none')
    setInterval(function(){
      $("#copyBtn1").text('content_copy')
      $("#copyBtn1").removeClass('c-green')
      $("#copiedSpan1").addClass('d-none')
    },1500);
  }

  copyAddressLink(e) {
    const token = e.params.token
    this.copyText(window.location.origin +"/withdrawl?token=" + token + "&asset=" + this.assetType + "&address=toaddress&amount=0.001")
    $("#copyBtn2").text('done')
    $("#copyBtn2").addClass('c-green')
    $("#copiedSpan2").removeClass('d-none')
    setInterval(function(){
      $("#copyBtn2").text('content_copy')
      $("#copyBtn2").removeClass('c-green')
      $("#copiedSpan2").addClass('d-none')
    },1500);
  }

  copyAddress(e) {
    this.copyText(this.currentAddress)
    $("#copyButton").text('done')
    $("#copyButton").addClass('c-green')
    $("#copiedSpan").removeClass('d-none')
    setInterval(function(){
      $("#copyButton").text('content_copy')
      $("#copyButton").removeClass('c-green')
      $("#copiedSpan").addClass('d-none')
    },1500);
  }

  createAddress() {
    const _this = this
    $.ajax({
      data: {
        assetType: _this.assetType
      },
      type: "POST", //OR GET
      url: '/createNewAddress', //The same form's action URL
      success: function (res) {
        if (!res.error) {
          const newAddress = res.data
          _this.currentAddress = newAddress
          //append to select options
          $("#addressSelector").append($('<option>', {
            value: newAddress,
            text: newAddress
          }))
          $("#createAddr_msg").addClass("d-none");
          _this.displayAddressArea(newAddress)
          $("#addressSelector").val(newAddress)
          return;
        } else {
          $("#createAddr_msg").removeClass("d-none");
          $("#createAddr_msg").text(res.msg);
        }
      },
    });
  }

  displayAddressArea(address) {
    $("#qrAddressArea").removeClass('d-none')
    var url = 'https://api.qrserver.com/v1/create-qr-code/?data=' + address + '&amp;size=150x150';
    $('#qrImageDisp').attr('src', url);
    $("#addressdisp").text(address)
  }

  updateExchangeRate() {
    const _this = this;
    _this.updateRateToDisplay()
    setInterval(async function () {
      _this.updateRateToDisplay()
    }, 7000);
  }

  hanlderSendTradingRequestButton() {
    const disabled = this.tradingAmountError || this.paymentBalanceError
    $("#tradingRequestButton").prop("disabled", disabled)
  }

  handlerUpdateRateToDisplay(rateMapJson, allRateMapJson) {
    const rateString = rateMapJson[this.assetType]
    if (!rateString) {
      return
    }
    const rate = parseFloat(rateString)
    this.rate = rate
    $("#exchangeRate").text(formatToLocalString(this.balance * rate, 2, 2))
    let sendValue = $("#amountSend").val()
    if(!sendValue || sendValue <= 0) {
      sendValue = 0
    }
    if (!this.btcFee) {
      this.btcFee = 0.0
    }
    $("#sendAmountExchange").text(formatToLocalString( sendValue * rate, 2, 2))
    $("#balanceAfterExchange").text(formatToLocalString((this.balance - sendValue) * rate, 2, 2))
    $("#feeExchange").text(formatToLocalString(Number(this.btcFee)* rate, 3, 3))
    $("#costExchange").text(formatToLocalString((Number(this.btcFee) + Number(sendValue))* rate, 3, 3))

    //check and set for payment type exchange rate
    //check payment type
    if(!this.paymentType || this.paymentType == "") {
      $("#paymentNeedAmount").text("0")
      $("#paymentAfterBalance").text("0")
      $("#paymentExchangeRate").text("0")
      return
    }

    //check symbol exchange rate
    const symbolRateStr = allRateMapJson[this.assetType + this.paymentType]
    if(!symbolRateStr || symbolRateStr == "") {
      $("#paymentNeedAmount").text("0")
      $("#paymentAfterBalance").text("0")
      $("#paymentExchangeRate").text("0")
      return 
    }
    this.paymentRate = parseFloat(symbolRateStr)
    const amount = $("#amountTrading").val()
    if(!amount || amount == "") {
      amount = 0
    }
    const roundNumber = this.getRoundNumber(this.paymentType)
    const needAmount = Number(amount) * this.paymentRate
    //check if need amount more than balance
    if(needAmount > this.currentPaymentBalance && this.tradingType == "buy") {
      $("#paymentBalanceErr_msg").removeClass("d-none")
      $("#paymentBalanceErr_msg").text("The " + this.paymentType.toUpperCase() +" balance is not enough to perform this trading")
      this.paymentBalanceError = true
    } else {
      this.paymentBalanceError = false
      $("#paymentBalanceErr_msg").addClass("d-none")
    }
    this.hanlderSendTradingRequestButton()
    const afterBalance = this.tradingType == "sell" ? needAmount + this.currentPaymentBalance : this.currentPaymentBalance - needAmount
    $("#paymentNeedAmount").text(formatToLocalString(needAmount, roundNumber, roundNumber))
    $("#paymentAfterBalance").text(formatToLocalString(afterBalance, roundNumber, roundNumber))
    $("#paymentExchangeRate").text(formatToLocalString(this.paymentRate, roundNumber, roundNumber))
  }

  updateRateToDisplay() {
    let rateMapJson = RateJson
    let allRateMapJson = AllRateJson
    if (rateMapJson == null || allRateMapJson == null) {
      this.fetchRate()
      return
    }
    this.handlerUpdateRateToDisplay(rateMapJson, allRateMapJson)
  }

  fetchRate() {
    const _this = this
    $.ajax({
      type: "GET", //OR GET
      url: '/fetch-rate', //The same form's action URL
      success: function (res) {
        if (res.error) {
          return
        }
        const rateStr = res.data
        const rateObject = JSON.parse(rateStr)
        const rateMapJson = rateObject.usdRates
        const allRateMapJson = rateObject.allRates
        if (!rateMapJson) {
          return
        }
        _this.handlerUpdateRateToDisplay(rateMapJson, allRateMapJson)
      },
    });
  }
}
