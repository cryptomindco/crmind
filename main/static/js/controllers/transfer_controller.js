
import BaseController from "./base_controller";

export default class extends BaseController {
  static values = {
    sendBy: String,
    currentIdStr: String,
    walletType: String,
    addressError: Boolean,
    amountError: Boolean,
    btcFee: Number,
    fundHex: String,
    confirmed: Boolean,
    rateMap: Object,
    balanceMap: Object,
    typeList: Object,
  };

  static get targets() {
    return ["transferForm"];
  }
  async initialize() {
    $("#transferTopBtn").css("backgroundColor", "#0e8b5b")
    $("#transferTopBtn").css("borderColor", "#0e8b5b")
    const _this = this;
    let typeJson = _this.data.get("types");
    this.typeList = JSON.parse(typeJson)
    this.btcFee = 0.0

    //init balance map
    this.balanceMap = new Map()
    if (this.typeList !== null) {
      for(let i = 0; i < this.typeList.length ; i++) {
        this.balanceMap.set(this.typeList[i], parseFloat($("#" + this.typeList[i] + "Value").text()))
      }
    }

    $("#currency").on("change", function (e) {
      const currency = e.target.value
      if(!currency || currency == "") {
        return
      }
      if(currency != "usd") {
        $("#feeArea").removeClass('d-none')
      } else {
        $("#feeArea").addClass('d-none')
      }
      switch(currency) {
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
      _this.handlerWalletDisplay()
      _this.handlerCurrencyChange()
      _this.checkValidTransferButton()
    });

    _this.transferFormTarget.addEventListener("submit", (e) => {
      e.preventDefault();
      if ($("#amountSend").val() == 0) {
        $("#amounterr_msg").removeClass("d-none")
        $("#amounterr_msg").text("The transfer amount must be greater than 0")
        $("#transferButton").prop("disabled", true)
        return false
      }
      const validSubmit = _this.checkValidTransferButton()
      if(!validSubmit) {
        return false
      }
      if(this.walletType != "usd") {
        $("#confirmPasswordDialog").on("shown.bs.modal", function () {}).modal('show');
        return false
      }
      _this.submitTransferForm()
      return false;
    });

    this.walletType = "usd"
    this.sendBy = "username"
    this.amountError = false

    $("input[type=radio][name=sendby]").change(function () {
      _this.sendBy = this.value
      _this.handlerWalletDisplay()
      _this.checkValidTransferButton()
    });

    this.handlerWalletDisplay()
    this.handlerCurrencyChange()
    this.updateExchangeRate()
  }

  confirmBtcAmount() {
    const amount = Number($("#amountSend").val())
    if(amount == 0 || this.walletType == 'usd') {
      return
    }
    const userId = $("#username").val()
    const _this = this
    $.ajax({
      data: {
        amount: amount,
        asset: _this.walletType
      },
      type: "POST", //OR GET
      url: '/confirmAmount', //The same form's action URL
      success: function (data) {
        if (data["error"] == "") { 
          $("#amounterr_msg").addClass("d-none")
          const result = data["result"]
          const jsonResult = JSON.parse(result)
          if(!jsonResult){
            return
          }
          _this.btcFee = jsonResult.fee
          _this.fundHex = jsonResult.hex
          $("#btcFee").text(formatToLocalString(Number(_this.btcFee), 8, 8))
          $("#btcCost").text(formatToLocalString(Number(_this.btcFee + amount), 8, 8))
          _this.checkValidTransferButton()
          _this.btcFee = 0.0
          $("#btcConfirmBtn").addClass('d-none')
        }
        if (data["error"] != "") {
          $("#amounterr_msg").removeClass("d-none")
          $("#amounterr_msg").text(data["error_msg"])
        }
      },
    });
  }

  submitTransferForm() {
    const url = '/transfer-amount'
    const userId = $("#username").val()
    const address = $("#toAddress").val()
    const password = $("#password").val()
    const _this = this
    const currency = $("#currency").val()
    let currentRate = 0.0
    if(currency != "usd" && this.rateMap != null) {
      currentRate = this.rateMap.get(currency)
    }
    let data = {
        id: userId,
        address: address,
        fundHex: this.fundHex,
        sendBy: this.sendBy,
        password: password,
        currency: $("#currency").val(),
        amount: $("#amountSend").val(),
        rate: currentRate,
        note: $("#noteText").val(),
      }
    $.ajax({
      data: data,
      type: "POST", //OR GET
      url: url, //The same form's action URL
      success: function (data) {
        if (data["error"] == "") {
          window.location.href = "/";
          return;
        }
        if (data["error"] != "") {
          if(_this.walletType == "usd") {
            $("#resulterr_msg").removeClass("d-none");
            $("#resulterr_msg").text(data["error_msg"]);
          } else {
            $("#confirmErr").removeClass("d-none");
            $("#confirmErr").text(data["error_msg"]);
          }
        }
      },
    });
  }

  receivingAddressChange(e) {
    this.checkValidTransferButton()
  }

  confirmPassword() {
    this.submitTransferForm()
  }

  usernameChange(e) {
    if(this.walletType != 'usd') {
      this.checkValidTransferButton()
    }
  }

  toAssetDetail(e) {
    const type = e.params.type
    window.location.href = '/assets/detail?type=' + type
  }

  amountInputChange(e) {
    const sendAmount = Number($("#amountSend").val());
    const _this = this
    const currency = $("#currency").val()
    const amount = this.balanceMap.get(currency)
    if (sendAmount > amount) {
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

    if($("#currency").val() != "usd") {
      const rate = this.rateMap.get(currency)
      let sendAmoutExchange = rate * sendAmount;
      let afterAmountExchange = rate * (amount - sendAmount);
      $("#sendAmountExchange").text(formatToLocalString(sendAmoutExchange, 2, 2));
      $("#balanceAfterExchange").text(formatToLocalString(afterAmountExchange, 2, 2));
    }
    let balanceAfterDisp = ''
    const diff = this.getRoundNumber(currency)
    const afterAmount = formatToLocalString(amount - sendAmount, diff, diff);
      balanceAfterDisp = currency == "usd" ? "$" + afterAmount : afterAmount + " " + currency.toUpperCase()
    $("#balanceAfter").text(balanceAfterDisp);
    this.checkValidTransferButton()
  }

  checkValidTransferButton() {
    const sendAmount = Number($("#amountSend").val());
    if(this.walletType != "usd" && this.sendBy == "address") {
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
    if(this.walletType != "usd") {
          //get balance of currency
          const balance = this.balanceMap.get(this.walletType)
          //if cost more than balance, return error
          if(this.btcFee + sendAmount > balance) {
            $("#confirmAmountErr_msg").removeClass('d-none')
            this.amountError = true
          } else {
            $("#confirmAmountErr_msg").addClass('d-none')
          }
      if(this.sendBy == "username") {
        disableButton = this.amountError
      } else {
        disableButton = this.amountError || this.addressError
      }
    } else {
      disableButton = this.amountError
    }
    $("#transferButton").prop("disabled", disableButton)
    if(this.walletType != "usd" && sendAmount > 0 && !disableButton) {
      $("#btcConfirmBtn").removeClass('d-none')
      this.confirmed = false
    } else {
      $("#btcConfirmBtn").addClass('d-none')
    }
    return !disableButton
  }

  handlerCurrencyChange() {
    let amountDisp;
    const currency = $("#currency").val()
    $("#amountSend").val(0);
    $("#amountSymbol").text(currency.toUpperCase())
    const roundNumber = this.getRoundNumber(currency)
    //get amount value
    const amountFloat = this.balanceMap.get(currency)
    if (currency == "usd") {
      amountDisp = "$" + formatToLocalString(Number(amountFloat), 2, 2)
      $("#balanceRateArea").addClass("d-none");
      $("#sendAmountExchangeArea").addClass("d-none");
      $("#balanceAffterRateArea").addClass("d-none");
    } else {
      amountDisp = formatToLocalString(Number(amountFloat), roundNumber, roundNumber) + " " + currency.toUpperCase()
      $("#balanceRateArea").removeClass("d-none");
      $("#sendAmountExchangeArea").removeClass("d-none");
      $("#balanceAffterRateArea").removeClass("d-none");
    }
    $("#amountDisplay").text(amountDisp);
    $("#balanceAfter").text(amountDisp);
  }

  handlerWalletDisplay() {
    this.walletType = $("#currency").val()
    if ($("#currency").val() == "usd") {
      $("#sendbySelector").addClass("d-none")
      $("#receiveAddressInput").addClass("d-none")
      $("#receiverUserNameSelector").removeClass("d-none")
    } else {
      $("#sendbySelector").removeClass("d-none")
      //if send by address
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

  updateExchangeRate() {
    const _this = this;
    _this.updateRateToDisplay()
    setInterval(async function () {
      _this.updateRateToDisplay()
    }, 7000);
  }

  updateRateToDisplay() {
    const _this = this
    $.ajax({
      type: "GET", //OR GET
      url: '/fetch-rate', //The same form's action URL
      success: function (data) {
        const rateStr = data["rateMap"]
        const rateObject = JSON.parse(rateStr)
        const rateMapJson = rateObject.usdRates
        if (!rateMapJson) {
          return
        }
        let currentRate = 0.0
        let currentBalance = 0.0
        _this.rateMap = new Map()
        //update rate of all assets
        for(let i = 0; i < _this.typeList.length ; i++) {
          const type = _this.typeList[i]
          const rateStr = rateMapJson[type]
          if (!rateStr) {
            continue
          }
          const rate = parseFloat(rateStr)
          _this.rateMap.set(type, rate)
          const balanceStr = $("#" + type + "Value").text()
          const balance = parseFloat(balanceStr.trim())
          $("#" + type + "Rate").text(formatToLocalString(balance * rate, 2, 2))
          if(_this.walletType == type) {
            currentRate = rate
            currentBalance = balance
          }
        }
        $("#balanceExchange").text(formatToLocalString(currentBalance * currentRate, 2, 2))
        let sendValue = $("#amountSend").val()
        if(!sendValue || sendValue <= 0) {
          sendValue = 0
        }
        $("#sendAmountExchange").text(formatToLocalString( sendValue * currentRate, 2, 2))
        $("#balanceAfterExchange").text(formatToLocalString( (currentBalance - sendValue) * currentRate, 2, 2))
      },
    });
  }
}
