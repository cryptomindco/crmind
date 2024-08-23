import BaseController from "./base_controller";

export default class extends BaseController {
  static values = {
    userId: Number,
    username: String,
    currentAction: String,
    activeBtnId: String,
    activeType: String,
    currentBalance: Number,
    currentErrorId: String,
  };

  static get targets() {
    return ["passwordForm", "password", "cpassword"];
  }

  async initialize() {
    const _this = this;
    this.userId = parseInt(this.data.get("userId"));
    this.username = this.data.get("userName")
    const successFlg = this.data.get("successFlg");
    const successMsg = this.data.get("successfullyMsg");
    if (successFlg == "true") {
      this.showSuccessToast(successMsg);
    }
  }

  changeBalance(e) {
    const type = e.params.type
    if(type == this.activeType && this.currentAction == "update") {
      return
    }
    //remove error
    if(this.currentErrorId) {
      $("#" + this.currentErrorId).addClass("d-none")
    }
    this.activeType = type
    const balance = e.params.balance
    this.currentBalance = balance
    this.currentAction = "update"
    $("#" + type + "Input").val(balance)
    //display input area
    $("#" + type + "InputArea").removeClass('d-none')
    //change active button
    $("#" + type + "Change").removeClass("btn-primary")
    $("#" + type + "Change").addClass("btn-success")
    //remove before active
    if(this.activeBtnId) {
      $("#" + this.activeBtnId).removeClass("btn-success")
      $("#" + this.activeBtnId).addClass("btn-primary")
    }
    
    this.activeBtnId = type + "Change"
    $('.admin-asset-action-area').each(
      function(index, element) {
        if((type + "InputArea") != element.id) {
          $(this).addClass('d-none')
        }
      }
    )
  }

  depositBalance(e) {
    const type = e.params.type
    if(type == this.activeType && this.currentAction == "deposit") {
      return
    }
    //remove error
    if(this.currentErrorId) {
      $("#" + this.currentErrorId).addClass("d-none")
    }
    $("#" + type + "Input").val(0)
    this.activeType = type
    this.currentAction = "deposit"
    const balance = e.params.balance
    this.currentBalance = balance
    $("#" + type + "InputArea").removeClass('d-none')

    //change active button
    $("#" + type + "Deposit").removeClass("btn-primary")
    $("#" + type + "Deposit").addClass("btn-success")
    //remove before active
    if(this.activeBtnId) {
      $("#" + this.activeBtnId).removeClass("btn-success")
      $("#" + this.activeBtnId).addClass("btn-primary")
    }
    
    this.activeBtnId = type + "Deposit"
    $('.admin-asset-action-area').each(
      function(index, element) {
        if((type + "InputArea") != element.id) {
          $(this).addClass('d-none')
        }
      }
    )
  }

  withdrawalBalance(e) {
    const type = e.params.type
    if(type == this.activeType && this.currentAction == "withdrawal") {
      return
    }
    //remove error
    if(this.currentErrorId) {
      $("#" + this.currentErrorId).addClass("d-none")
    }
    $("#" + type + "Input").val(0)
    this.activeType = type
    this.currentAction = "withdrawal"
    const balance = e.params.balance
    this.currentBalance = balance
    $("#" + type + "InputArea").removeClass('d-none')
    //change active button
    $("#" + type + "Withdrawal").removeClass("btn-primary")
    $("#" + type + "Withdrawal").addClass("btn-success")
    //remove before active
    if(this.activeBtnId) {
      $("#" + this.activeBtnId).removeClass("btn-success")
      $("#" + this.activeBtnId).addClass("btn-primary")
    }  
    this.activeBtnId = type + "Withdrawal"
    $('.admin-asset-action-area').each(
      function(index, element) {
        if((type + "InputArea") != element.id) {
          $(this).addClass('d-none')
        }
      }
    )
  }

  inputChange() {
    if(!this.currentAction || !this.activeType){
      return
    }
    //get amount
    const inputValue = $("#" + this.activeType + "Input").val()

    //if update and balance equal with input value
    if(this.currentAction == "update" && inputValue == this.currentBalance) {
      this.handlerErrorForInput("The balance does not change. Please enter another value")
      return
    }

    //if withdrawal, check value less than balance
    if(this.currentAction == "withdrawal" && inputValue > this.currentBalance) {
        this.handlerErrorForInput("The withdrawal amount cannot exceed the balance")
        return
    }
    //if input is zero, notifications
    if (this.currentAction != "update" && (!inputValue || inputValue == 0 || inputValue == "")) {
      this.handlerErrorForInput("The number entered cannot be 0")
      return
    }
    $("#" + this.activeType + "Inputerr_msg").addClass('d-none')
    $("#" + this.activeType + "UpdateBtn").prop("disabled", false)
  }

  updateAction() {
    if(!this.currentAction || !this.activeType){
      return
    }
    //get amount
    const inputValue = $("#" + this.activeType + "Input").val()

    //if update and balance equal with input value
    if(this.currentAction == "update" && inputValue == this.currentBalance) {
      this.handlerErrorForInput("The balance does not change. Please enter another value")
      return
    }

    //if withdrawal, check value less than balance
     if(this.currentAction == "withdrawal" && inputValue > this.currentBalance) {
      this.handlerErrorForInput("The withdrawal amount cannot exceed the balance")
      return
    }
    //if input is zero, notifications
    if (this.currentAction != "update" && (!inputValue || inputValue == 0 || inputValue == "")) {
      this.handlerErrorForInput("The number entered cannot be 0")
      return
    }

    $("#confirmDialog").on("shown.bs.modal", function () {}).modal('show');
  }

  handlerErrorForInput(errMsg) {
    $("#" + this.activeType + "Inputerr_msg").removeClass('d-none')
    $("#" + this.activeType + "Inputerr_msg").text(errMsg)
    $("#" + this.activeType + "UpdateBtn").prop("disabled", true)
    this.currentErrorId = this.activeType + "Inputerr_msg"
  }

  submitUpdateBalance() {
    const _this = this
    //post to api
    $.ajax({
      data: {
        input: $("#" + _this.activeType + "Input").val(),
        action: _this.currentAction,
        type: _this.activeType,
        username: _this.username
      },
      type: "POST", //OR GET
      url: '/adminUpdateBalance', //The same form's action URL
      success: function (data) {
        if (data["error"] == "") {
          window.location.href = "/admin/user?id=" + _this.userId;
          return;
        }
        if (data["error"] != "") {
          _this.handlerErrorForInput(data["error_msg"])
        }
      },
    });
  }

  balanceChange() {
    const balance = $("#usdBalance").val()
    if(!balance || balance === '') {
      $("#balanceUpdateBtn").prop("disabled", true);
      $("#balanceInputErr").removeClass('d-none')
      $("#balanceInputErr").text('Balance cannot be left blank')
      return
    }
    $("#balanceUpdateBtn").prop("disabled", false);
    $("#balanceInputErr").addClass('d-none')
  }

  updateBalance() {
    const balance = $("#usdBalance").val()
    if(!balance || balance === '') {
      $("#balanceInputErr").removeClass('d-none')
      $("#balanceInputErr").text('Balance cannot be left blank')
      $("#balanceUpdateBtn").prop("disabled", true);
      return
    }
    const _this = this
    $("#balanceInputErr").addClass('d-none')
    $.ajax({
      data: {
        userId: this.userId,
        usdBalance: balance
      },
      type: "POST",
      url: "/admin/updateUsdBalance",
      success: function (data) {
        if (data["error"] == "") {
          _this.showSuccessToast('Update USD Balance successfully');
          return;
        }
        if (data["error"] != "") {
          $("#balanceInputErr").removeClass('d-none')
          $("#balanceInputErr").text(data["error_msg"])
        }
      },
    });
  }

  onActiveChange(e) {
    const _this = this
    const activeFlg = $("#userActiveSwitch").is(":checked")
    $("#activeText").text(activeFlg ? 'Active' : 'Deactive')
    $.ajax({
      data: {
        userId: this.userId,
        active: activeFlg ? 1 : 0
      },
      type: "POST",
      url: "/admin/ChangeUserStatus",
      success: function (res) {
        if (!res.error) {
          const activeStr = activeFlg ? "activated" : "deactivated"
          _this.showSuccessToast('User has been ' + activeStr);
          $("#usdBalance").prop("disabled", !activeFlg);
          $("#balanceUpdateBtn").prop("disabled", !activeFlg);
          return;
        } else {
          _this.showErrorToast("An error has occurred. Please reload the page");
        }
      },
    });
  }
}
