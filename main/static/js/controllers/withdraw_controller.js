import BaseController from "./base_controller";

export default class extends BaseController {
  static values = {
    code: String,
    withdrawBy: String,
    createAccountFlg: Boolean,
    session: Number,
    randomUsername: String,
    isConfirm: Boolean,
  };

  async initialize() {
    this.code = this.data.get("code");
    this.withdrawBy = "username";
    this.handlerDisplayComponent();
    this.createAccountFlg = false;
    const _this = this;
    if (this.withdrawBy == "username") {
      $("#userAddrLabel").addClass("d-none");
      $("#targetInput").addClass("d-none");
      $("#confirmBtn").addClass("d-none");
      $("#loginBtn").removeClass("d-none");
      $("#registerBtn").removeClass("d-none");
    } else {
      $("#userAddrLabel").removeClass("d-none");
      $("#targetInput").removeClass("d-none");
      $("#targetInput").val("");
      $("#confirmBtn").removeClass("d-none");
      $("#loginBtn").addClass("d-none");
      $("#registerBtn").addClass("d-none");
    }
    $("input[type=radio][name=withdrawby]").change(function () {
      _this.withdrawBy = this.value;
      _this.handlerDisplayComponent();
    });
  }

  handlerDisplayComponent() {
    $("#noteArea").addClass("d-none");
    $("#createAndWithdrawBtn").addClass("d-none");
    $("#cancelBtn").addClass("d-none");
    if (this.withdrawBy == "username") {
      $("#userAddrLabel").addClass("d-none");
      $("#targetInput").addClass("d-none");
      $("#confirmBtn").addClass("d-none");
      $("#loginBtn").removeClass("d-none");
      $("#registerBtn").removeClass("d-none");
    } else {
      $("#userAddrLabel").removeClass("d-none");
      $("#userAddrLabel").text("Receive Address");
      $("#targetInput").removeClass("d-none");
      $("#targetInput").val("");
      $("#confirmBtn").removeClass("d-none");
      $("#loginBtn").addClass("d-none");
      $("#registerBtn").addClass("d-none");
    }
  }

  targetInputChange() {
    const target = $("#targetInput").val();
    if (!target || target == "") {
      $("#targeterr_msg").removeClass("d-none");
      $("#targeterr_msg").text(
        (this.withdrawBy == "username" ? "Username" : "Address") +
          " cannot be blank"
      );
      if (this.withdrawBy == "username") {
        $("#createAndWithdrawBtn").prop("disabled", true);
      } else {
        $("#confirmBtn").prop("disabled", true);
      }
      return;
    }
    $("#targeterr_msg").addClass("d-none");
    if (this.withdrawBy == "username") {
      $("#createAndWithdrawBtn").prop("disabled", false);
    } else {
      $("#confirmBtn").prop("disabled", false);
    }
  }

  confirmWithdraw() {
    const target = $("#targetInput").val();
    //check username
    if (!target || target == "") {
      $("#targeterr_msg").removeClass("d-none");
      $("#targeterr_msg").text("Address cannot be blank");
      return;
    }
    const _this = this;
    $.ajax({
      data: {
        code: _this.code,
        target: target,
      },
      type: "POST", //OR GET
      url: "/confirmWithdraw", //The same form's action URL
      success: function (res) {
        if (!res.error) {
          $("#confirmArea").addClass("d-none");
          $("#successMsgArea").removeClass("d-none");
          $("#resultMsg").text(
            "Withdraw successfully! Withdraw is confirmed to be sent to the address successfully"
          );
          return;
        } else {
          $("#confirmerr_msg").removeClass("d-none");
          $("#confirmerr_msg").text(res.msg);
        }
      },
    });
  }

  withdrawRegister(e) {
    const _this = this;
    e.preventDefault();
    $.ajax({
      data: {},
      type: "GET", //OR GET
      url: "/gen-random-username", //The same form's action URL
      success: function (res) {
        if (!res.error) {
          const result = JSON.parse(res.data);
          _this.randomUsername = result.username;
          _this.isConfirm = true;
          $("#userAddrLabel").removeClass("d-none");
          $("#userAddrLabel").text("Username");
          $("#noteArea").removeClass("d-none");
          $("#createAndWithdrawBtn").removeClass("d-none");
          $("#loginBtn").addClass("d-none");
          $("#registerBtn").addClass("d-none");
          $("#cancelBtn").removeClass("d-none");
          $("#targetInput").removeClass("d-none");
          $("#targetInput").val(_this.randomUsername);
        } else {
          $("#confirmerr_msg").removeClass("d-none");
          $("#confirmerr_msg").text(res.msg);
        }
      },
    });
  }

  inputFocus() {
    if (this.withdrawBy != "username") {
      return;
    }
    $("#targeterr_msg").addClass("d-none");
    if (this.randomUsername != $("#targetInput").val().trim()) {
      return;
    }
    //when focus to input, hide msg text and clear random username
    $("#noteArea").addClass("d-none");
    $("#targetInput").val("");
  }

  inputBlur() {
    if (this.withdrawBy != "username") {
      return;
    }
    const curUsername = $("#targetInput").val();
    if (!curUsername || curUsername == "") {
      $("#targetInput").val(this.randomUsername);
      $("#noteArea").removeClass("d-none");
      $("#targeterr_msg").addClass("d-none");
      $("#createAndWithdrawBtn").prop("disabled", false);
    }
  }

  createAndWithdraw(e) {
    const _this = this;
    e.preventDefault();
    //check new username is valid
    const newUsername = $("#targetInput").val().trim();
    if (newUsername != this.randomUsername) {
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
              $("#targeterr_msg").removeClass("d-none");
              $("#targeterr_msg").text("That name is not available");
              $("#createAndWithdrawBtn").prop("disabled", true);
              return;
            }
            _this.handlerRegister(newUsername);
          } else {
            $("#createAndWithdrawBtn").prop("disabled", true);
            $("#targeterr_msg").removeClass("d-none");
            $("#targeterr_msg").text(res.msg);
            return;
          }
        },
      });
      return;
    }
    this.handlerRegister(newUsername);
  }

  handlerRegister(newUsername) {
    let sessionKey;
    let options;
    const _this = this;
    $("#userAddrLabel").addClass("d-none");
    $("#targetInput").addClass("d-none");
    $("#noteArea").addClass("d-none");
    $("#loadingArea").removeClass("d-none");
    $("#loadingText").text("Creating new account...");
    $("#createAndWithdrawBtn").addClass("d-none");
    $("#cancelBtn").addClass("d-none");
    //check and create new username
    $.ajax({
      data: {
        username: newUsername,
      },
      type: "POST", //OR GET
      url: "/passkey/registerStart", //The same form's action URL
      success: function (res) {
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
            $("#loadingArea").addClass("d-none");
            _this.handlerDisplayComponent();
            $("#confirmerr_msg").removeClass("d-none");
            $("#confirmerr_msg").text("Registration error. Please try again!");
            return;
          }
          _this.handlerFinishRegistration(options, sessionKey);
        } else {
          _this.cancelRegisterUser(sessionKey);
          $("#loadingArea").addClass("d-none");
          _this.handlerDisplayComponent();
          $("#confirmerr_msg").removeClass("d-none");
          $("#confirmerr_msg").text(res.msg);
        }
      },
    });
  }

  withdrawRegister2() {
    const _this = this;
    let sessionKey;
    let options;
    $("#loadingArea").removeClass("d-none");
    $("#loadingText").text("Withdrawing with new user...");
    //check and create new username
    $.ajax({
      data: {},
      type: "POST", //OR GET
      url: "/passkey/registerStart", //The same form's action URL
      success: function (res) {
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
            $("#loadingArea").addClass("d-none");
            $("#confirmerr_msg").removeClass("d-none");
            $("#confirmerr_msg").text("Registration error. Please try again!");
            return;
          }
          _this.handlerFinishRegistration(options, sessionKey);
        } else {
          _this.cancelRegisterUser(sessionKey);
          $("#loadingArea").addClass("d-none");
          $("#confirmerr_msg").removeClass("d-none");
          $("#confirmerr_msg").text(res.msg);
        }
      },
    });
  }

  async handlerFinishRegistration(options, sessionKey) {
    let attestationResponse;
    const _this = this;
    try {
      attestationResponse = await SimpleWebAuthnBrowser.startRegistration(
        options.publicKey
      );
    } catch (error) {
      _this.cancelRegisterUser(sessionKey);
      console.log("Cancel user registration");
      $("#loadingArea").addClass("d-none");
      _this.handlerDisplayComponent();
      return;
    }
    fetch("/passkey/withdrawWithNewAccountFinish", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Session-Key": sessionKey,
        "Url-Code": _this.code,
      },
      body: JSON.stringify(attestationResponse),
    })
      .then(function (a) {
        return a.json(); // call the json method on the response to get JSON
      })
      .then(function (json) {
        if (!json.error) {
          window.location.href = "/";
        } else {
          $("#loadingArea").addClass("d-none");
          $("#confirmerr_msg").removeClass("d-none");
          $("#confirmerr_msg").text(json.error_msg);
          _this.handlerDisplayComponent();
          _this.cancelRegisterUser(sessionKey);
        }
      });
  }

  withdrawLogin() {
    const _this = this;
    $("#loadingArea").removeClass("d-none");
    $("#loadingText").text("Loading passkeys...");
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
          _this.session++;
          const opts = resultData.options;
          const sessionKey = resultData.sessionkey;
          _this.hanlderFinish(opts, sessionKey, false, _this.session);
        }
      },
    });
  }

  async hanlderFinish(opts, sessionKey, startConditionalUI, objSession) {
    const { startAuthentication } = SimpleWebAuthnBrowser;
    let asseResp;
    try {
      asseResp = await startAuthentication(opts.publicKey, startConditionalUI);
    } catch (error) {
      if (objSession == this.session) {
        $("#loadingArea").addClass("d-none");
      }
      console.log("Conditional UI request was aborted");
      return;
    }
    const _this = this;
    fetch("/assertion/withdrawConfirmLoginResult", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Session-Key": sessionKey,
        "Url-Code": _this.code,
      },
      body: JSON.stringify(asseResp),
    })
      .then(function (a) {
        return a.json(); // call the json method on the response to get JSON
      })
      .then(function (json) {
        if (!json.error) {
          window.location.href = "/";
        } else {
          $("#loadingArea").addClass("d-none");
          $("#confirmerr_msg").removeClass("d-none");
          $("#confirmerr_msg").text(json.error_msg);
        }
      });
  }
}
