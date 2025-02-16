import BaseController from "./base_controller";

export default class extends BaseController {
  static values = {
    randomUsername: String,
    isConfirm: Boolean,
    usePassword: Boolean,
    isRegister: Boolean,
  };

  async initialize() {}

  async hanlderFinish(opts, sessionKey, startConditionalUI) {
    const { startAuthentication } = SimpleWebAuthnBrowser;
    let asseResp;
    try {
      asseResp = await startAuthentication({ optionsJSON: opts.publicKey, useBrowserAutofill: startConditionalUI });
    } catch (error) {
      console.log("Conditional UI request was aborted");
      $("#loadingArea").addClass("d-none");
      $("#loginButton").removeClass("d-none");
      $("#sectionTitle").text("Login with passkey");
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
          $("#loginButton").removeClass("d-none");
          $("#sectionTitle").text("Login with passkey");
        }
      });
  }

  loginWithPassBtn() {
    this.usePassword = !this.usePassword
    this.isRegister ? this.setSignupComponent() : this.setLoginComponent()
  }

  switchMode(e) {
    const _this = this;
    this.isRegister = !this.isRegister
    $("#loginErr_msg").addClass("d-none");
    $("#passloginerr_msg").addClass("d-none");
    $("#registererr_msg").addClass("d-none");
    if (!this.isRegister) {
      this.setLoginComponent()
      return
    }
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
          _this.isRegister = true
          _this.setSignupComponent()
        }
        if (res.error) {
          $("#loginErr_msg").removeClass("d-none");
          $("#loginErr_msg").text(res.msg);
        }
      },
    });
  }

  setSignupComponent() {
    $("#usernameConfirmArea").removeClass("d-none")
    $("#switchModeText").text("Back to login")
    $("#backIcon").html("undo")
    $("#refreshUsernameBtn").removeClass("d-none")
    $("#randomText").removeClass("d-none")
    $("#loginButton").addClass("d-none");
    $("#confirmButton").removeClass("d-none");
    $("#usernameInput").val(this.randomUsername);
    if (this.usePassword) {
      $("#passwordLoginArea").removeClass("d-none")
      $("#loginTypeBtnText").text("Signup with passkey")
      $("#sectionTitle").text("Signup with password")
      $("#loginModeIcon").html("fingerprint")
    } else {
      $("#passwordLoginArea").addClass("d-none")
      $("#loginTypeBtnText").text("Signup with password")
      $("#sectionTitle").text("Signup with passkey")
      $("#loginModeIcon").html("keyboard_lock")
    }
  }

  setLoginComponent() {
    $("#switchModeText").text("Signup")
    $("#backIcon").html("app_registration")
    $("#randomText").addClass("d-none")
    $("#loginButton").removeClass("d-none");
    $("#confirmButton").addClass("d-none");
    $("#usernameInput").val("");
    if (this.usePassword) {
      $("#usernameConfirmArea").removeClass("d-none")
      $("#refreshUsernameBtn").addClass("d-none")
      $("#passwordLoginArea").removeClass("d-none")
      $("#loginTypeBtnText").text("Login with passkey")
      $("#sectionTitle").text("Login with password")
      $("#loginIcon").html("keyboard_lock")
      $("#loginModeIcon").html("fingerprint")
    } else {
      $("#usernameConfirmArea").addClass("d-none")
      $("#passwordLoginArea").addClass("d-none")
      $("#loginTypeBtnText").text("Login with password")
      $("#sectionTitle").text("Login with passkey")
      $("#loginIcon").html("fingerprint")
      $("#loginModeIcon").html("keyboard_lock")
    }
  }

  inputBlur() {
    if (!this.isRegister) {
      return
    }
    const curUsername = $("#usernameInput").val();
    if (!curUsername || curUsername == "") {
      $("#usernameInput").val(this.randomUsername);
      $("#randomUsernameMsg").removeClass("d-none");
    }
  }

  confirmRegister(e) {
    const _this = this;
    e.preventDefault();
    //check new username is valid
    const newUsername = $("#usernameInput").val().trim();
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
              $("#loginErr_msg").removeClass("d-none");
              $("#loginErr_msg").text("That name is not available");
              $("#confirmButton").prop("disabled", true);
              return;
            }
            _this.usePassword ? _this.registerByPassword() : _this.handlerRegister(newUsername);
            return
          } else {
            $("#confirmButton").prop("disabled", true);
            $("#loginErr_msg").removeClass("d-none");
            $("#loginErr_msg").text(res.msg);
            return;
          }
        },
      });
      return;
    }
    this.usePassword ? this.registerByPassword() : this.handlerRegister(newUsername);
  }

  registerByPassword() {
    const newUsername = $("#usernameInput").val().trim();
    const password = $("#passwordLoginInput").val().trim();
    if (password == "") {
      $("#passloginerr_msg").removeClass("d-none")
      $("#passloginerr_msg").text("Password cannot be blank")
      return
    }
    $.ajax({
      data: {
        username: newUsername,
        password: password,
      },
      type: "POST",
      url: "/password/register",
      success: function (res) {
        if (!res.error) {
          window.location.href = "/";
        } else {
          $("#loginErr_msg").removeClass("d-none");
          $("#loginErr_msg").text(res.msg);
        }
      },
    });
  }

  handlerRegister(newUsername) {
    let sessionKey;
    let options;
    const _this = this;
    $("#usernameConfirmArea").addClass("d-none");
    $("#loadingArea").removeClass("d-none");
    $("#loadingText").text("Creating new account...");
    $("#confirmButton").addClass("d-none");
    $("#sectionTitle").addClass("d-none");
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
            $("#confirmButton").removeClass("d-none");
            $("#loginErr_msg").removeClass("d-none");
            $("#loginErr_msg").text("Registration error. Please try again!");
            $("#sectionTitle").removeClass("d-none");
            return;
          }
          _this.handlerFinishRegistration(options, sessionKey);
        } else {
          _this.cancelRegisterUser(sessionKey);
          $("#loadingArea").addClass("d-none");
          $("#confirmButton").removeClass("d-none");
          $("#loginErr_msg").removeClass("d-none");
          $("#loginErr_msg").text(res.msg);
          $("#sectionTitle").removeClass("d-none");
        }
      },
    });
  }

  passwordChange() {
    if (!this.usePassword) {
      $("#passloginerr_msg").addClass("d-none")
      return
    }
    const pass = $("#passwordLoginInput").val().trim();
    if (pass == "") {
      $("#passloginerr_msg").removeClass("d-none")
      $("#passloginerr_msg").text("Password cannot be blank")
      return
    }
    $("#passloginerr_msg").addClass("d-none")
    // TODO: Will check the password reliability later.
  }

  resetToStartPage() {
    $("#loginButton").removeClass("d-none");
    $("#confirmButton").addClass("d-none");
    $("#usernameConfirmArea").addClass("d-none");
    $("#sectionTitle").text("Login with passkey");
    $("#registererr_msg").addClass("d-none");
    $("#loginErr_msg").addClass("d-none");
    $("#registerBtn").removeClass("d-none");
    $("#backBtn").addClass("d-none");
  }

  async handlerFinishRegistration(options, sessionKey) {
    const { startRegistration } = SimpleWebAuthnBrowser;
    let attestationResponse;
    const _this = this;
    try {
      attestationResponse = await startRegistration(
        { optionsJSON: options.publicKey }
      );
    } catch (error) {
      _this.cancelRegisterUser(sessionKey);
      console.log("Cancel user registration");
      $("#loadingArea").addClass("d-none");
      $("#sectionTitle").removeClass("d-none");
      _this.setSignupComponent();
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
          window.location.href = "/";
        } else {
          $("#loadingArea").addClass("d-none");
          $("#sectionTitle").removeClass("d-none");
          $("#loginErr_msg").removeClass("d-none");
          $("#loginErr_msg").text(json.msg);
          _this.cancelRegisterUser(sessionKey);
          _this.setSignupComponent();
        }
      });
  }

  usernameChange() {
    const newUsername = $("#usernameInput").val().trim();
    if (!newUsername || newUsername == "") {
      if (!this.isRegister) {
        $("#registererr_msg").removeClass("d-none");
        $("#registererr_msg").text("Username cannot be blank")
      }
      $("#confirmButton").prop("disabled", true);
      return;
    }
    $("#confirmButton").prop("disabled", false);
    $("#registererr_msg").addClass("d-none");
  }

  inputFocus() {
    if (!this.isRegister) {
      return
    }
    $("#registererr_msg").addClass("d-none");
    if (this.randomUsername != $("#usernameInput").val().trim()) {
      return;
    }
    //when focus to input, hide msg text and clear random username
    $("#randomUsernameMsg").addClass("d-none");
    $("#usernameInput").val("");
  }

  openLoginWithPasskey() {
    $("#loadingArea").removeClass("d-none");
    $("#sectionTitle").text("Select your Passkey");
    $("#loadingText").text("Loading...");
    $("#loginButton").addClass("d-none");
    $("#loginErr_msg").addClass("d-none");
    const _this = this;
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
          _this.hanlderFinish(opts, sessionKey, false);
        } else {
          _this.resetToStartPage();
        }
      },
    });
  }

  loginWithPassword() {
    const newUsername = $("#usernameInput").val().trim();
    const password = $("#passwordLoginInput").val().trim();
    if (newUsername == "") {
      $("#registererr_msg").removeClass("d-none")
      $("#registererr_msg").text("Username cannot be blank")
      return
    }
    if (password == "") {
      $("#passloginerr_msg").removeClass("d-none")
      $("#passloginerr_msg").text("Password cannot be blank")
      return
    }
    $.ajax({
      data: {
        username: newUsername,
        password: password,
      },
      type: "POST",
      url: "/password/login",
      success: function (res) {
        if (!res.error) {
          window.location.href = "/";
        } else {
          $("#loginErr_msg").removeClass("d-none");
          $("#loginErr_msg").text(res.msg);
        }
      },
    });
  }

  openLogin() {
    this.usePassword ? this.loginWithPassword() : this.openLoginWithPasskey()
  }
  
  refreshRandomUsername(e) {
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
          $("#usernameInput").val(_this.randomUsername);
          $("#loginErr_msg").addClass("d-none");
        }
        if (res.error) {
          $("#loginErr_msg").removeClass("d-none");
          $("#loginErr_msg").text(res.msg);
        }
      },
    });
  }
}
