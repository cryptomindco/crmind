import { Controller } from "@hotwired/stimulus";

export default class BaseController extends Controller {
  getQueryParamString(defaultParamValues, paramValues) {
    let params = [];
    if (paramValues == undefined || paramValues == null) {
      return "";
    }
    for (const k in paramValues) {
      if (
        !paramValues[k] ||
        paramValues[k].toString() === defaultParamValues[k].toString()
      )
        continue;
      params.push(k + "=" + paramValues[k]);
    }

    return params.join("&");
  }

  cancelRegisterUser(sessionKey) {
    $.ajax({
      data: {
        sessionKey: sessionKey,
      },
      type: "POST", //OR GET
      url: "/passkey/cancelRegister", //The same form's action URL
      success: function (data) {
        return;
      },
    });
  }

  GetChatTime(createdt) {
   const createDate = new Date(createdt*1000)
   const now = new Date()
    //if is tody, only display time. else, display month, day, year and time
    if (createDate.getUTCFullYear() == now.getUTCFullYear() && createDate.getUTCMonth() == now.getUTCMonth() && createDate.getUTCDay() == now.getUTCDay()) {
      return formatDate(createDate, 'hh:mm')
    }
    //if the same year, return with month, day, hour, minute
    if(createDate.getUTCFullYear() == now.getUTCFullYear()) {
      return formatDate(createDate, 'MM/dd, hh:mm')
    }
    return formatDate(createDate, 'yyyy/MM/dd, hh:mm')
  }

  getRoundNumber(currency) {
    switch(currency) {
      case "dcr":
        return 7
      case "ltc":
        return 8
      case "btc":
        return 8
      default:
        return 2
    }
  }

  getStepByType(type) {
    switch(type) {
      case "btc":
        return '0.00000001'
      case "ltc":
        return '0.00000001'
      case "dcr":
        return '0.0000001'
      default:
        return '0.01'
    }
  }

  getAssetColor(currency) {
    switch(currency) {
      case "dcr":
        return "#D4F3E1"
      case "ltc":
        return "#ffe9f9"
      case "btc":
        return "#ebf5ff"
      default:
        return "#fff2f2"
    }
  }

  getAssetIcon(currency) {
    switch(currency) {
      case "dcr":
        return "dcr-icon.svg"
      case "ltc":
        return "ltc-icon.svg"
      case "btc":
        return "btc-icon.svg"
      default:
        return "usd-icon.svg"
    }
  }

  getParamsFromUrl() {
    var search = location.search.substring(1);
    var queryString = decodeURI(search)
      .replace(/"/g, '\\"')
      .replace(/&/g, '","')
      .replace(/=/g, '":"');
    if (queryString === "") {
      return {};
    }
    var params = JSON.parse('{"' + queryString + '"}');
    return params;
  }

  updateQueryUrl(queryParams, defaultSettings) {
    const [queryArr, settings, defaults] = [[], queryParams, defaultSettings]
    for (const k in settings) {
      if (!settings[k] || settings[k].toString() === defaults[k].toString()) continue
      queryArr.push(k + '=' + settings[k])
    }
    if (queryArr.length > 0) {
      const param = queryArr.join("&")
      window.history.replaceState(null, null, "?" + param);
    } else {
      window.history.pushState({}, null, location.href.split('?')[0])
    }
  }

  async updateExchangeRateImmediatelly(elem, url, amountBtc) {
    const response = await fetch(url);
    const text = await response.text();
    if (text.trim() == "") {
      elem.textContent = "NaN";
    } else {
      elem.textContent = (Number(text) * amountBtc).toLocaleString("en");
    }
  }

  updateExchangeRate(elem, url, amountBtc) {
    setInterval(async function () {
      const response = await fetch(url);
      const text = await response.text();
      if (text.trim() == "") {
        elem.textContent = "NaN";
      } else {
        elem.textContent = (Number(text) * amountBtc).toLocaleString("en");
      }
    }, 5000);
  }

  showSuccessToast(content) {
    $("#successBody").text(content);
    $("#successToast").toast({
      autohide: true,
    });
    // Show toast
    $("#successToast").toast("show");
  }

  showErrorToast(content) {
    $("#errorBody").text(content);
    $("#errorToast").toast({
      autohide: true,
    });
    // Show toast
    $("#errorToast").toast("show");
  }

  copyText(text) {
    navigator.clipboard.writeText(text);
  }

  toUpperFirstCase(input) {
    if(!input || input == "") {
      return input
    }
    return input[0].toUpperCase() + input.slice(1)
  }
}
