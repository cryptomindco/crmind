import BaseController from "./base_controller";

export default class extends BaseController {
  updateSettings() {
    const assetStr = 'usd,btc,dcr,ltc'
    const assetArr = assetStr.split(',')
    const selected = []
    assetArr.forEach((asset) => {
      if($("#" + asset + "Select").is(':checked')) {
        selected.push(asset)
      }
    })
    const selectedAssetStr = selected.join(',')
    const _this = this
    $.ajax({
      data: {
        selectedAsset: selectedAssetStr
      },
      type: "POST", //OR GET
      url: '/updateSettings', //The same form's action URL
      success: function (data) {
        if (data["error"] == "") {
          $("#updateSettings_msg").addClass("d-none")
          _this.showSuccessToast("Update settings successfully");
        }
        if (data["error"] != "") {
          $("#updateSettings_msg").removeClass("d-none")
          $("#updateSettings_msg").text(data["error_msg"])
        }
      },
    });
  }

  syncTransactions() {
    const _this = this
    $.ajax({
      data: {},
      type: "POST", //OR GET
      url: '/syncTransactions', //The same form's action URL
      success: function (data) {
        if (data["error"] == "") {
          _this.showSuccessToast("Synchronization is being performed in the background. Please wait a few minutes for the synchronization process to complete");
        }

        if (data["error"] != "") {
          _this.showErrorToast(data["error_msg"]);
        }
      },
    });
  }
}
