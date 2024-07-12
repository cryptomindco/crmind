import BaseController from "./base_controller";

export default class extends BaseController {
  static values = {
    defaultQueryParams: Object,
    queryParams: Object,
    rate: Number,
    typeList: Object,
  };

  static get targets() {
    return ["historyTable"];
  }

  async initialize() {
    const isSuperUser = this.data.get("superuser")
    if (isSuperUser != "true") {
      const typeJson = this.data.get("types");
      this.typeList = JSON.parse(typeJson)
      this.updateAssetsExchangeRate()
    }
    this.loadLastTx()
  }

  toAssetDetail(e) {
    const type = e.params.type
    window.location.href = '/assets/detail?type=' + type
  }

  loadLastTx() {
    const _this = this
    $.ajax({
      url: "/transfer/GetLastTxs",
      data: {
        limit: 5,
      },
      type: "GET",
      success: function (data) {
        if (!data) {
          return
        }
        //create history list
        _this.historyTableTarget.innerHTML = _this.createHistoryTable(data)
      },
    });
  }

  createHistoryTable(data) {
    if (!data || data.length < 1) {
      return '<p class="mt-2">Data does not exist</p>'
    }
    let inner = ''
    const _this = this
    data.forEach(element => {
      inner += `<div class="row p-2 transaction-history-row border-bottom-lblue" onclick="toTransactionDetail(${element.id})">` +
        `<p class="mb-0 ps-0 d-flex ai-center">` +
        `<img src="/static/images/icons/${_this.getAssetIcon(element.currency)}" alt="mdo" width="25" height="25" class="rounded-circle">` +
        `<span class="ms-2">` +
        `<span class="fs-15 sent-receive-label ${element.isTrading ? element.tradingType : (element.isSender ? 'send' : 'receive')}-label">${element.isTrading ? (_this.toUpperFirstCase(element.tradingType)) : (element.isSender ? 'Sent ' : 'Received ')}</span>`

      let diff = _this.getRoundNumber(element.currency)
      inner += `<span class="fw-600 fs-17">${element.currency === 'usd' ? '$' : ''}${formatToLocalString(element.amount, diff, diff)} ${element.currency !== 'usd' ? element.currency.toUpperCase() : ''}</span>`

      if(!element.isTrading) {
        if (element.currency !== 'usd') {
          inner += `(~$<span class="amount-number">${formatToLocalString(element.rateValue, 2, 2)}</span>)`
        }
        if (element.senderId > 0) {
          inner += `<span class="fs-15">${element.isSender ? ' to ' : ' from '}</span>`
          if (element.currency !== 'usd' && element.isSender && element.receiverId < 1) {
            inner += `<span class="fw-600 fs-17">${element.toAddress}</span>`
          } else {
            inner += `<span class="fw-600 fs-17">${element.isSender ? element.receiver : element.sender}</span>`
          }
        }  
      } else {
        inner += '.'
        const paymentRoundNumber = this.getRoundNumber(element.paymentType)
        inner += `<span class="fs-15">${element.tradingType == "buy" ? ' Paid by ' : ' Received by '}</span>`
        inner += `<span class="fw-600 fs-17">${element.paymentType.toUpperCase()}</span><span class="fs-15"> with </span>`
        inner += `<span class="fw-600 fs-17">${formatToLocalString(element.amount*element.rate, paymentRoundNumber, paymentRoundNumber)} ${element.paymentType.toUpperCase()}</span>`
      }
      inner += `<span class="fs-15"> on ${element.createdtDisp}</span>`
      inner += `<span class="fs-14"> (*Note: <em>${element.description}</em>)</span> `
      if (element.currency !== 'usd' && element.txid && element.txid != '' && !element.isOffChain) {
        inner += `<br /><span class="fs-13"><strong>Txid: </strong>${element.txid}</span><b class="fs-13" style="color:${element.confirmed ? 'green' : '#a8184f'};">` +
          ` (${element.confirmed ? 'Confirmed' : 'Unconfirmed'}: ${element.confirmed ? '' : element.confirmations + '/'}${element.confirmationNeed}${element.confirmed ? '+' : ''} <span class="material-symbols-outlined fs-13i">${element.confirmed ? 'lock' : 'lock_open_right'}</span>)</b>`
      }
      inner += `</span></p></div>`
    });
    return inner
  }

  fetchRate() {
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
        _this.handlerUpdateRateToDisplay(rateMapJson)
      },
    });
  }

  handlerUpdateRateToDisplay(rateMapJson) {
    //update rate of all assets
    for (let i = 0; i < this.typeList.length; i++) {
      const type = this.typeList[i]
      const rateStr = rateMapJson[type]
      if (!rateStr) {
        continue
      }
      const rate = parseFloat(rateStr)
      $("#" + type + "Rate").text(formatToLocalString(rate, 2, 2))
    }
}

  updateAssetsExchangeRate() {
    const _this = this;
    _this.updateRateToDisplay()
    setInterval(async function () {
      _this.updateRateToDisplay()
    }, 7000);
  }

  updateRateToDisplay() {
    let rateMapJson = RateJson
    if (rateMapJson == null) {
      this.fetchRate()
      return
    }
    this.handlerUpdateRateToDisplay(rateMapJson)
  }
}
