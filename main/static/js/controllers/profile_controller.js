import BaseController from "./base_controller";

export default class extends BaseController {
  static values = {
    username: String,
    newUsername: String,
    dialogType: String,
  };

  async initialize() {
    this.username = this.data.get("username");
    const successFlg = this.data.get("successFlg");
    const successMsg = this.data.get("successfullyMsg");
    if (successFlg == "true") {
      this.showSuccessToast(successMsg);
    }
  }

  confirmDialogYes() {
    switch (this.dialogType) {
      case "username":
        this.confirmChangeUsername();
        break;
      case "add":
        this.handlerUpdatePasskey(false);
        break;
      case "reset":
        this.handlerUpdatePasskey(true);
    }
  }

  showAddCredentialDialog() {
    $("#dialogContent").text(
      "Resend your current passkey to your passkey manager. Would you like to continue?"
    );
    this.dialogType = "add";
    $("#usernameChangeConfirm")
      .on("shown.bs.modal", function () {})
      .modal("show");
  }

  showUpdatePasskeyDialog() {
    $("#dialogContent").text(
      "Invalidated your current passkey and makes a new one. Would you like to continue?"
    );
    this.dialogType = "reset";
    $("#usernameChangeConfirm")
      .on("shown.bs.modal", function () {})
      .modal("show");
  }

  handlerUpdatePasskey(isReset) {
    $("#usernameChangeConfirm")
      .on("shown.bs.modal", function () {})
      .modal("hide");
    if (!this.username || this.username == "") {
      $("#updateErr").removeClass("d-none");
      $("#updateErr").text("Username is empty. Please try again");
      return;
    }
    let sessionKey;
    let options;
    const _this = this;
    $.ajax({
      data: {},
      type: "POST", //OR GET
      url: "/passkey/updateStart", //The same form's action URL
      success: function (res) {
        if (!res.error) {
          const resultData = JSON.parse(res.data);
          if (resultData) {
            options = resultData.options;
            sessionKey = resultData.sessionkey;
          }
          if (!options) {
            $("#updateErr").removeClass("d-none");
            $("#updateErr").text("Registration error. Please try again!");
            return;
          }
          _this.handlerFinishUpdatePasskey(options, sessionKey, isReset);
        } else {
          $("#updateErr").removeClass("d-none");
          $("#updateErr").text(res.msg);
        }
      },
    });
  }

  closeConfirmDialog() {
    $("#loadingArea").addClass("d-none");
  }

  editBtnClick() {
    $("#usernameInput").val(this.username);
    $("#usernameLabel").addClass("d-none");
    $("#editBtn").addClass("d-none");
    $("#usernameInput").removeClass("d-none");
    $("#saveBtn").removeClass("d-none");
    $("#closeBtn").removeClass("d-none");
  }

  closeBtnClick() {
    $("#usernameLabel").removeClass("d-none");
    $("#editBtn").removeClass("d-none");
    $("#usernameInput").addClass("d-none");
    $("#saveBtn").addClass("d-none");
    $("#closeBtn").addClass("d-none");
    $("#newUsernameErr").addClass("d-none");
  }

  usernameChange() {
    const newUsername = $("#usernameInput").val().trim();
    if (!newUsername || newUsername == "") {
      $("#newUsernameErr").removeClass("d-none");
      $("#newUsernameErr").text("Username cannot be blank");
      $("#saveBtn").addClass("d-none");
      return;
    }

    if (newUsername == this.username) {
      $("#newUsernameErr").removeClass("d-none");
      $("#newUsernameErr").text("Username is the same as before");
      $("#saveBtn").addClass("d-none");
      return;
    }
    $("#newUsernameErr").addClass("d-none");
    $("#saveBtn").removeClass("d-none");
  }

  saveBtnClick() {
    const newUsername = $("#usernameInput").val().trim();
    if (!newUsername || newUsername == "") {
      $("#newUsernameErr").removeClass("d-none");
      $("#newUsernameErr").text("Username cannot be blank");
      return;
    }
    if (this.username == newUsername) {
      this.closeBtnClick();
      return;
    }
    //check username exist
    const _this = this;
    //check username exist on system
    $.ajax({
      data: {
        username: newUsername,
      },
      type: "GET", //OR GET
      url: "/check-user", //The same form's action URL
      success: function (res) {
        if (!res.error) {
          //check result
          const exist = res.data;
          if (exist) {
            $("#newUsernameErr").removeClass("d-none");
            $("#newUsernameErr").text("Username already exists");
            return;
          }
          _this.newUsername = newUsername;
          $("#loadingArea").removeClass("d-none");
          $("#loadingText").text("Updating username");
          $("#dialogContent").text(
            "Changing your user name will require changing your passkey to work with your new username. Would you like to continue?"
          );
          _this.dialogType = "username";
          $("#usernameChangeConfirm")
            .on("shown.bs.modal", function () {})
            .modal("show");
        } else {
          $("#newUsernameErr").removeClass("d-none");
          $("#newUsernameErr").text(res.msg);
        }
      },
    });
  }

  confirmChangeUsername() {
    $("#usernameChangeConfirm")
      .on("shown.bs.modal", function () {})
      .modal("hide");
    if (!this.newUsername || this.newUsername == "") {
      $("#newUsernameErr").removeClass("d-none");
      $("#newUsernameErr").text("Get new username failed");
      $("#loadingArea").addClass("d-none");
      return;
    }
    const _this = this;
    //check and create new username
    $.ajax({
      data: {
        username: _this.newUsername,
      },
      type: "POST", //OR GET
      url: "/passkey/registerStart", //The same form's action URL
      success: function (res) {
        let sessionKey;
        let options;
        if (!res.error) {
          const resultData = JSON.parse(res.data);
          if (resultData) {
            options = resultData.options;
            sessionKey = resultData.sessionkey;
          }
          if (!options) {
            if (sessionKey && sessionKey != "") {
              _this.cancelRegisterUser(sessionKey);
            }
            $("#newUsernameErr").removeClass("d-none");
            $("#newUsernameErr").text("Registration error. Please try again!");
            $("#loadingArea").addClass("d-none");
            return;
          }
          _this.handlerFinishChangeUsername(options, sessionKey);
        } else {
          _this.cancelRegisterUser(sessionKey);
          $("#newUsernameErr").removeClass("d-none");
          $("#newUsernameErr").text(res.msg);
          $("#loadingArea").addClass("d-none");
        }
      },
    });
  }

  async handlerFinishChangeUsername(options, sessionKey) {
    let attestationResponse;
    const _this = this;
    try {
      attestationResponse = await SimpleWebAuthnBrowser.startRegistration(
        options.publicKey
      );
    } catch (error) {
      _this.cancelRegisterUser(sessionKey);
      $("#loadingArea").addClass("d-none");
      console.log("Cancel user registration");
      return;
    }
    fetch("/passkey/changeUsernameFinish", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Session-Key": sessionKey,
        "Old-Username": _this.username,
      },
      body: JSON.stringify(attestationResponse),
    })
      .then(function (a) {
        return a.json(); // call the json method on the response to get JSON
      })
      .then(function (json) {
        if (!json.error) {
          window.location.reload();
        } else {
          $("#newUsernameErr").removeClass("d-none");
          $("#newUsernameErr").text(json.msg);
          $("#loadingArea").addClass("d-none");
          _this.cancelRegisterUser(sessionKey);
        }
      });
  }

  async handlerFinishUpdatePasskey(options, sessionKey, isReset) {
    const attestationResponse = await SimpleWebAuthnBrowser.startRegistration(
      options.publicKey
    );
    const _this = this;
    fetch("/passkey/updateFinish", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Session-Key": sessionKey,
        "Is-Reset-Key": isReset,
      },
      body: JSON.stringify(attestationResponse),
    })
      .then(function (a) {
        return a.json(); // call the json method on the response to get JSON
      })
      .then(function (json) {
        if (!json.error) {
          _this.showSuccessToast("Passkey updated successfully");
        } else {
          $("#updateErr").removeClass("d-none");
          $("#updateErr").text(json.msg);
        }
      });
  }
}
