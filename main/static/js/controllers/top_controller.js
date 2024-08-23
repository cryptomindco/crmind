import BaseController from "./base_controller";

export default class extends BaseController {
  async initialize() {}

  async hanlderFinish(opts, sessionKey, startConditionalUI) {
    const _this = this
    const { startAuthentication } = SimpleWebAuthnBrowser;
    let asseResp;
    try {
      asseResp = await startAuthentication(opts.publicKey, startConditionalUI);
    } catch (error) {
      $("#loadingModal").on("shown.bs.modal", function () {}).modal('hide');
      console.log("Conditional UI request was aborted");
      return;
    }
    fetch("/assertion/result", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Session-Key": sessionKey,
      },
      body: JSON.stringify(asseResp),
    })
      .then(function (a) {
        return a.json(); // call the json method on the response to get JSON
      })
      .then(function (json) {
        if (!json.error) {
          window.location.reload()
        } else {
          _this.showErrorToast('Switch account failed: ' + json.msg)
          $("#loadingModal").on("shown.bs.modal", function () {}).modal('hide');
        }
      });
  }

  switchAccount() {
    const _this = this
    $("#loadingModal").on("shown.bs.modal", function () {}).modal('show');
    $.ajax({
      data: {},
      type: "POST", //OR GET
      url: "/assertion/options", //The same form's action URL
      success: function (res) {
        if (!res.error) {
          const resultData = JSON.parse(res.data);
          if (!resultData || !resultData.options) {
            return;
          }
          const opts = resultData.options;
          const sessionKey = resultData.sessionkey;
          _this.hanlderFinish(opts, sessionKey, false);
        }
      },
    });
  }
}
