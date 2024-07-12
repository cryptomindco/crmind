import BaseController from "./base_controller";

export default class extends BaseController {
  static values = {
    selectedType: String,
    currentAddress: String,
  }

  async initialize() {
    this.handlerCoinSelect()
  }

  createAddress() {
    if(!this.selectedType || this.selectedType === '') {
      $("#createAddr_msg").removeClass("d-none")
      $("#createAddr_msg").text('There was an error with the selected Asset Type. Try refreshing the page')
      return
    }
    const _this = this
    $.ajax({
      data: {
        assetType: _this.selectedType
      },
      type: "POST", //OR GET
      url: '/createNewAddress', //The same form's action URL
      success: function (data) {
        if (data["error"] == "") {
          const newAddress = data["address"]
          _this.currentAddress = newAddress
          $("#createNewAddressArea").addClass('d-none')
          $("#qrAddressArea").removeClass('d-none')
          //set src of qr image
          $("#qrImageDisp").attr("src", "https://chart.googleapis.com/chart?chs=250x250&cht=qr&chl=" + newAddress)
          $("#addressdisp").text(newAddress)
          $("#" + _this.selectedType + "Option").attr("value", _this.selectedType + "---" + newAddress)
          $("#assetSelector").val(_this.selectedType + "---" + newAddress)
          return;
        }
        if (data["error"] != "") {
          $("#createAddr_msg").removeClass("d-none");
          $("#createAddr_msg").text(data["error_msg"]);
        }
      },
    });
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

  coinSelectChange() {
    this.handlerCoinSelect()
  }

  handlerCoinSelect() {
    $("#createAddr_msg").addClass("d-none")
    const selectedValue = $("#assetSelector").val()
    if (!selectedValue || selectedValue === '') {
      this.selectedType = ''
      return
    }
    const selectTypeArr = selectedValue.trim().split("---")
    if(selectTypeArr.length == 0){
      this.selectedType = ''
      return
    }
    //update current type
    this.selectedType = selectTypeArr[0]
    //if array less than 2, have no address. Display Create address button
    if(selectTypeArr.length < 2) {
      $("#createNewAddressArea").removeClass('d-none')
      $("#qrAddressArea").addClass('d-none')
      return
    }
    this.currentAddress = selectTypeArr[1]
    $("#createNewAddressArea").addClass('d-none')
    $("#qrAddressArea").removeClass('d-none')
    //set src of qr image
    $("#qrImageDisp").attr("src", "https://chart.googleapis.com/chart?chs=250x250&cht=qr&chl=" + this.currentAddress)
    $("#addressdisp").text(this.currentAddress)
  }
}
