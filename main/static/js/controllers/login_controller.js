import BaseController from "./base_controller";

export default class extends BaseController {
  static values = {
    randomUsername: String,
    isConfirm: Boolean,
  }

  async initialize() {
  }

  async hanlderFinish(opts, sessionKey, startConditionalUI) {
    const { startAuthentication } = SimpleWebAuthnBrowser;
    let asseResp;
    try {
      asseResp = await startAuthentication(opts.publicKey, startConditionalUI);
    } catch (error) {
      console.log("Conditional UI request was aborted");
      $("#loadingArea").addClass("d-none");
      $("#loginButton").removeClass("d-none")
      $("#sectionTitle").text("Login with passkey")
      $("#createAccount").removeClass("d-none")
      $("#orText").removeClass("d-none")
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
          window.location.href = "/";
        } else {
          $("#loginErr_msg").removeClass("d-none");
          $("#loginErr_msg").text(json.msg);
          $("#loadingArea").addClass("d-none");
          $("#loginButton").removeClass("d-none")
          $("#sectionTitle").text("Login with passkey")
          $("#createAccount").removeClass("d-none")
          $("#orText").removeClass("d-none")
        }
      });
  }

  registerUser(e) {
    const _this = this;
    e.preventDefault();
    $.ajax({
      data: {},
      type: "GET", //OR GET
      url: "/gen-random-username", //The same form's action URL
      success: function (res) {
        if (!res.error) {
          const result = JSON.parse(res.data)
          _this.randomUsername = result.username
          _this.isConfirm = true;
          $("#loginButton").addClass("d-none")
          $("#confirmButton").removeClass("d-none")
          $("#createAccount").addClass("d-none")
          $("#orText").addClass("d-none")
          $("#usernameConfirmArea").removeClass("d-none")
          $("#usernameInput").val(_this.randomUsername)
          $("#sectionTitle").text('Create Account')
          $("#registerBtn").addClass('d-none')
          $("#footerArea").removeClass('d-none')
          $("#backBtn").removeClass('d-none')
          $("#loginErr_msg").addClass("d-none")
        }
        if (res.error) {
          $("#loginErr_msg").removeClass("d-none")
          $("#loginErr_msg").text(res.msg)
        }
      },
    });
  }

  inputBlur() {
    const curUsername = $("#usernameInput").val()
    if (!curUsername || curUsername == '') {
      $("#usernameInput").val(this.randomUsername)
      $("#randomUsernameMsg").removeClass('d-none')
    }
  }

  confirmRegister(e) {
    const _this = this;
    e.preventDefault();
    //check new username is valid
    const newUsername = $("#usernameInput").val().trim()
    if (newUsername != this.randomUsername) {
      $.ajax({
        data: {
          username: newUsername,
        },
        type: "GET", //OR GET
        url: '/check-user', //The same form's action URL
        success: function (res) {
          if (!res.error) {
            //check result
            if(res.data) {
              const result = res.data
              if (result.exist) {
                $("#registererr_msg").removeClass("d-none")
                $("#registererr_msg").text('That name is not available')
                $("#confirmButton").prop("disabled", true)
                return
              }
            }
            _this.handlerRegister(newUsername)
          } else {
            $("#confirmButton").prop("disabled", true)
            $("#registererr_msg").removeClass("d-none")
            $("#registererr_msg").text(res.msg)
            return
          }
        },
      });
      return
    }
    this.handlerRegister(newUsername)
  }

  handlerRegister(newUsername) {
    let sessionKey
    let options;
    const _this = this
    $("#usernameConfirmArea").addClass("d-none");
    $("#loadingArea").removeClass("d-none");
    $("#loadingText").text("Creating new account...");
    $("#footerArea").addClass("d-none")
    $("#confirmButton").addClass("d-none")
    $("#sectionTitle").addClass("d-none")
    //check and create new username
    $.ajax({
      data: {
        username: newUsername,
      },
      type: "POST", //OR GET
      url: "/passkey/registerStart", //The same form's action URL
      success: function (res) {
        if (!res.error) {
          const resultData = res.data
          if (resultData) {
            options = resultData.options;
            sessionKey = resultData.sessionkey;
          }
          if (!options) {
            if (sessionKey && sessionKey != "") {
              _this.cancelRegisterUser(sessionKey);
            }
            $("#loadingArea").addClass("d-none");
            $("#footerArea").removeClass("d-none")
            $("#confirmButton").removeClass("d-none")
            $("#registererr_msg").removeClass("d-none");
            $("#registererr_msg").text("Registration error. Please try again!");
            $("#sectionTitle").removeClass("d-none")
            return;
          }
          _this.handlerFinishRegistration(options, sessionKey);
        } else {
          _this.cancelRegisterUser(sessionKey);
          $("#loadingArea").addClass("d-none");
          $("#footerArea").removeClass("d-none")
          $("#confirmButton").removeClass("d-none")
          $("#registererr_msg").removeClass("d-none");
          $("#registererr_msg").text(res.msg);
          $("#sectionTitle").removeClass("d-none")
        }
      },
    });
  }

  resetToStartPage() {
    $("#loginButton").removeClass("d-none")
    $("#footerArea").addClass("d-none")
    $("#confirmButton").addClass("d-none")
    $("#createAccount").removeClass("d-none")
    $("#orText").removeClass("d-none")
    $("#usernameConfirmArea").addClass("d-none")
    $("#sectionTitle").text('Login with passkey')
    $("#registererr_msg").addClass("d-none")
    $("#loginErr_msg").addClass("d-none")
    $("#registerBtn").removeClass('d-none')
    $("#backBtn").addClass('d-none')
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
      $("#confirmButton").removeClass("d-none")
      $("#sectionTitle").removeClass("d-none")
      _this.resetToStartPage()
      return;
    }
    fetch("/passkey/registerFinish", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Session-Key": sessionKey,
      },
      body: JSON.stringify(attestationResponse),
    })
      .then(function (a) {
        return a.json(); // call the json method on the response to get JSON
      })
      .then(function (json) {
        if (!json.error) {
         //redirect to homepage
         window.location.href = "/"
        } else {
          $("#loadingArea").addClass("d-none");
          $("#sectionTitle").removeClass("d-none")
          $("#confirmButton").removeClass("d-none")
          $("#registererr_msg").removeClass("d-none");
          $("#registererr_msg").text(json.msg);
          _this.cancelRegisterUser(sessionKey);
          _this.resetToStartPage()
        }
      });
  }

  usernameChange() {
    $("#registererr_msg").addClass("d-none");
    const newUsername = $("#usernameInput").val().trim()
    if (!newUsername || newUsername == '') {
      $("#confirmButton").prop("disabled", true)
      return
    }
    $("#confirmButton").prop("disabled", false)
  }

  inputFocus() {
    $("#registererr_msg").addClass("d-none");
    if (this.randomUsername != $("#usernameInput").val().trim()) {
      return
    }
    //when focus to input, hide msg text and clear random username
    $("#randomUsernameMsg").addClass('d-none')
    $("#usernameInput").val('')
  }

  openLogin() {
    $("#loadingArea").removeClass("d-none");
    $("#createAccount").addClass("d-none")
    $("#orText").addClass("d-none")
    $("#sectionTitle").text("Select your Passkey")
    $("#loadingText").text("Loading...");
    $("#loginButton").addClass("d-none")
    $("#loginErr_msg").addClass("d-none")
    const _this = this;
    $.ajax({
      data: {},
      type: "POST", //OR GET
      url: "/assertion/options", //The same form's action URL
      success: function (res) {
        if (!res.error) {
          const resultData = res.data;
          if (!resultData || !resultData.options) {
            return;
          }
          _this.session++;
          const opts = resultData.options;
          const sessionKey = resultData.sessionkey;
          _this.hanlderFinish(opts, sessionKey, false);
        } else {
          _this.resetToStartPage()
        }
      },
    });
  }
}
