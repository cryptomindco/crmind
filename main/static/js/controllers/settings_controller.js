import BaseController from "./base_controller";

export default class extends BaseController {
  updateSettings() {
    const assetStr = 'usd,btc,dcr,ltc'
    const serviceStr = 'auth,chat,assets'
    const assetArr = assetStr.split(',')
    const serviceArr = serviceStr.split(',')
    const selected = []
    const serviceSelected = []
    assetArr.forEach((asset) => {
      if($("#" + asset + "Select").is(':checked')) {
        selected.push(asset)
      }
    })
    serviceArr.forEach((service) => {
      if($("#" + service + "Select").is(':checked')) {
        serviceSelected.push(service)
      }
    })
    const selectedAssetStr = selected.join(',')
    const selectedServicesStr = serviceSelected.join(',')
    const _this = this
    $.ajax({
      data: {
        selectedAsset: selectedAssetStr,
        selectedServices: selectedServicesStr,
      },
      type: "POST", //OR GET
      url: '/updateSettings', //The same form's action URL
      success: function (res) {
        if (!res.error) {
          $("#updateSettings_msg").addClass("d-none")
          _this.showSuccessToast("Update settings successfully");
        } else {
          $("#updateSettings_msg").removeClass("d-none")
          $("#updateSettings_msg").text(res.msg)
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
      success: function (res) {
        if (!res.error) {
          _this.showSuccessToast("Synchronization is being performed in the background. Please wait a few minutes for the synchronization process to complete");
        } else {
          _this.showErrorToast(res.msg);
        }
      },
    });
  }
}
