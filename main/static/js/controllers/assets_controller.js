import BaseController from "./base_controller";

export default class extends BaseController {
  static values = {
    defaultQueryParams: Object,
    queryParams: Object,
    typeList: Object,
    balanceMap: Object,
  };

  static get targets() {
    return ["historyTable", "paginationTopBar", "paginationBottomBar"];
  }

  async initialize() {
    const successFlg = this.data.get("successFlg");
    const successMsg = this.data.get("successfullyMsg");
    const assetsActive = this.data.get("assetActive")
    if (successFlg == "true") {
      this.showSuccessToast(successMsg);
    }
    if (assetsActive != "true") {
      return
    }
    const _this = this;
    let typeJson = _this.data.get("types");
    this.typeList = JSON.parse(typeJson);

    //init balance map
    this.balanceMap = new Map();
    if (this.typeList !== null) {
      for (let i = 0; i < this.typeList.length; i++) {
        this.balanceMap.set(
          this.typeList[i],
          parseFloat($("#" + this.typeList[i] + "Value").text())
        );
      }
    }

    this.updateAssetsExchangeRate();
    this.defaultQueryParams = {
      type: "all",
      atype: "all",
      direction: "all",
      perpage: "15",
      pageNum: 1,
    };
    this.queryParams = this.getParamsFromUrl();
    if (this.queryParams.type && this.queryParams.type !== "") {
      $("#transFilter").val(this.queryParams.type);
    }

    if (this.queryParams.direction && this.queryParams.direction !== "") {
      $("#directionFilter").val(this.queryParams.direction);
    }

    if (this.queryParams.perpage && this.queryParams.perpage !== "") {
      $("#numperpage").val(this.queryParams.perpage);
      $("#numperpageBottom").val(this.queryParams.perpage);
    }

    if (!this.queryParams.pageNum || this.queryParams.pageNum == "") {
      this.queryParams.pageNum = this.defaultQueryParams.pageNum;
    }

    if (!this.queryParams.perpage || this.queryParams.perpage == "") {
      this.queryParams.perpage = this.defaultQueryParams.perpage;
      $("#numperpage").val(this.queryParams.perpage);
      $("#numperpageBottom").val(this.queryParams.perpage);
    }

    this.loadHistoryList();
    this.loadAssetsList();
  }

  updateAssetsExchangeRate() {
    const _this = this;
    _this.updateRateToDisplay();
    setInterval(async function () {
      _this.updateRateToDisplay();
    }, 7000);
  }

  fetchRate() {
    const _this = this;
    $.ajax({
      type: "GET", //OR GET
      url: "/fetch-rate", //The same form's action URL
      success: function (res) {
        if (res.error) {
          return
        }
        const rateStr = res.data
        const rateObject = JSON.parse(rateStr);
        const rateMapJson = rateObject.usdRates;
        if (!rateMapJson) {
          return;
        }
        _this.handlerUpdateRateToDisplay(rateMapJson);
      },
    });
  }

  handlerUpdateRateToDisplay(rateMapJson) {
    //update rate of all assets
    for (let i = 0; i < this.typeList.length; i++) {
      const type = this.typeList[i];
      const rateStr = rateMapJson[type];
      if (!rateStr) {
        continue;
      }
      const rate = parseFloat(rateStr);
      const balanceStr = $("#" + type + "Value").text();
      const balance = parseFloat(balanceStr.trim());
      $("#" + type + "Rate").text(formatToLocalString(balance * rate, 2, 2));
    }
  }

  updateRateToDisplay() {
    let rateMapJson = RateJson;
    if (rateMapJson == null) {
      this.fetchRate();
      return;
    }
    this.handlerUpdateRateToDisplay(rateMapJson);
  }

  filterWallet(e) {
    this.queryParams.type = e.target.value;
    this.updateQueryUrl(this.queryParams, this.defaultQueryParams);
    this.loadHistoryList();
  }

  numPerPageChange(e) {
    this.queryParams.perpage = e.target.value;
    this.queryParams.pageNum = 1;
    $("#numperpageBottom").val(this.queryParams.perpage);
    this.updateQueryUrl(this.queryParams, this.defaultQueryParams);
    this.loadHistoryList();
  }

  numPerBottomPageChange(e) {
    this.queryParams.perpage = e.target.value;
    this.queryParams.pageNum = 1;
    $("#numperpage").val(this.queryParams.perpage);
    this.updateQueryUrl(this.queryParams, this.defaultQueryParams);
    this.loadHistoryList();
  }

  loadListByPageNumber(e) {
    this.queryParams.pageNum = e.target.value;
    this.updateQueryUrl(this.queryParams, this.defaultQueryParams);
    this.loadHistoryList();
  }

  filterDirection(e) {
    this.queryParams.direction = e.target.value;
    this.updateQueryUrl(this.queryParams, this.defaultQueryParams);
    this.loadHistoryList();
  }

  loadAssetsList() {
    if (this.queryParams.atype === "wallet") {
      $("#walletListArea").removeClass("d-none");
      $("#companyListArea").addClass("d-none");
      return;
    } else if (this.queryParams.atype === "company") {
      $("#walletListArea").addClass("d-none");
      $("#companyListArea").removeClass("d-none");
      return;
    }
    $("#walletListArea").removeClass("d-none");
    $("#companyListArea").removeClass("d-none");
  }

  loadHistoryList() {
    const _this = this;
    $.ajax({
      url: "/transfer/GetHistoryList",
      data: {
        type: this.queryParams.type,
        direction: this.queryParams.direction,
        perpage: this.queryParams.perpage,
        pageNum: this.queryParams.pageNum,
      },
      type: "GET",
      success: function (res) {
        if (!res) {
          return;
        }
        if (res.error) {
          return;
        }
        const data = JSON.parse(res.data);
        const pageCount = data.pageCount;
        //handler pagination
        _this.paginationTopBarTarget.innerHTML =
          _this.createPaginationBar(pageCount);
        _this.paginationBottomBarTarget.innerHTML =
          _this.createPaginationBar(pageCount);
        const dataList = data.list;
        //create history list
        _this.historyTableTarget.innerHTML = _this.createHistoryTable(dataList);
        if (!dataList || dataList.length < 1) {
          _this.historyTableTarget.classList.remove("history-table");
        } else {
          _this.historyTableTarget.classList.add("history-table");
        }
      },
    });
  }

  toAssetDetail(e) {
    const type = e.params.type;
    window.location.href = "/assets/detail?type=" + type;
  }

  pageNumberClick(e) {
    if (this.queryParams.pageNum == e.params.pagenumber) {
      return;
    }
    this.queryParams.pageNum = e.params.pagenumber;
    this.updateQueryUrl(this.queryParams, this.defaultQueryParams);
    this.loadHistoryList();
  }

  prevPageClick(e) {
    if (this.queryParams.pageNum == 1) {
      return;
    }
    this.queryParams.pageNum -= 1;
    this.updateQueryUrl(this.queryParams, this.defaultQueryParams);
    this.loadHistoryList();
  }

  nextPageClick(e) {
    if (this.queryParams.pageNum == e.params.pagecount) {
      return;
    }
    this.queryParams.pageNum += 1;
    this.updateQueryUrl(this.queryParams, this.defaultQueryParams);
    this.loadHistoryList();
  }

  createPaginationBar(pageCount) {
    let inner = `<div class="pagination-btn ${
      this.queryParams.pageNum == 1 ? "pagination-btn-disabled" : ""
    }"><a class="pagination-prev-next-btn" data-action="click->assets#prevPageClick"><span class="material-symbols-outlined">chevron_left</span></a></div>`;
    for (let i = 1; i <= pageCount; i++) {
      inner += `<div data-action="click->assets#pageNumberClick" data-assets-pagenumber-param="${i}" class="pagination-btn${
        this.queryParams.pageNum == i ? " active-btn" : ""
      }"><a>${i}</a></div>`;
    }
    inner += `<div class="pagination-btn ${
      this.queryParams.pageNum == pageCount ? "pagination-btn-disabled" : ""
    }"><a class="pagination-prev-next-btn" data-assets-pagecount-param="${pageCount}" data-action="click->assets#nextPageClick"><span class="material-symbols-outlined">chevron_right</span></a></div>`;
    return inner;
  }

  createHistoryTable(data) {
    if (!data || data.length < 1) {
      return "<p>Data does not exist</p>";
    }
    let inner = "";
    const _this = this;

    data.forEach((element) => {
      inner +=
        `<div class="row p-2 transaction-history-row border-bottom-lblue" onclick="toTransactionDetail(${element.id})">` +
        `<p class="mb-0 ps-0 d-flex ai-center">` +
        `<img src="/static/images/icons/${_this.getAssetIcon(
          element.currency
        )}" alt="mdo" width="25" height="25" class="rounded-circle">` +
        `<span class="ms-2 w-95">` +
        `<span class="fs-15 sent-receive-label ${
          element.isTrading
            ? element.tradingType
            : element.isSender
            ? "send"
            : "receive"
        }-label">${
          element.isTrading
            ? _this.toUpperFirstCase(element.tradingType)
            : element.isSender
            ? "Sent "
            : "Received "
        }</span>`;

      let diff = _this.getRoundNumber(element.currency);
      inner += `<span class="fw-600 fs-17">${
        element.currency === "usd" ? "$" : ""
      }${formatToLocalString(element.amount, diff, diff)} ${
        element.currency !== "usd" ? element.currency.toUpperCase() : ""
      }</span>`;

      if (!element.isTrading) {
        if (element.currency !== "usd") {
          inner += ` (~$<span class="amount-number">${formatToLocalString(
            element.rateValue,
            2,
            2
          )}</span>)`;
        }
        if (element.senderId > 0) {
          inner += `<span class="fs-15">${
            element.isSender ? " to " : " from "
          }</span>`;
          if (
            element.currency !== "usd" &&
            element.isSender &&
            element.receiverId < 1
          ) {
            inner += `<span class="fw-600 fs-17">${element.toAddress}</span>`;
          } else {
            inner += `<span class="fw-600 fs-17">${
              element.isSender ? element.receiver : element.sender
            }</span>`;
          }
        }
      } else {
        inner += ".";
        const paymentRoundNumber = this.getRoundNumber(element.paymentType);
        inner += `<span class="fs-15">${
          element.tradingType == "buy" ? " Paid by " : " Received by "
        }</span>`;
        inner += `<span class="fw-600 fs-17">${element.paymentType.toUpperCase()}</span><span class="fs-15"> with </span>`;
        inner += `<span class="fw-600 fs-17">${formatToLocalString(
          element.amount * element.rate,
          paymentRoundNumber,
          paymentRoundNumber
        )} ${element.paymentType.toUpperCase()}</span>`;
      }
      inner += `<span class="fs-15"> on ${element.createdtDisp}</span>`;
      inner += `<span class="fs-14"> (*Note: <em>${element.description}</em>)</span> `;
      if (
        element.currency !== "usd" &&
        element.txid &&
        element.txid != "" &&
        !element.isOffChain
      ) {
        inner +=
          `<br /><span class="fs-13"><strong>Txid: </strong>${
            element.txid
          }</span><b class="fs-13" style="color:${
            element.confirmed ? "green" : "#a8184f"
          };">` +
          ` (${element.confirmed ? "Confirmed" : "Unconfirmed"}: ${
            element.confirmed ? "" : element.confirmations + "/"
          }${element.confirmationNeed}${
            element.confirmed ? "+" : ""
          } <span class="material-symbols-outlined fs-13i">${
            element.confirmed ? "lock" : "lock_open_right"
          }</span>)</b>`;
      }
      inner += `</span></p></div>`;
    });
    return inner;
  }
}
