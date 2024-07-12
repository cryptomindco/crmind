package controllers

import (
	"auth/logpack"
	"auth/models"
	"auth/passkey"
	"auth/utils"
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

type AuthController struct {
	BaseLoginController
}

func (this *AuthController) BeginUpdatePasskey() {
	loginUser, check := this.SimpleAuthCheck()
	if !check {
		this.ResponseError("Login authentication error. Please try again", utils.GetFuncName(), nil)
		return
	}
	logpack.Info(fmt.Sprintf("begin update passkey of: %s", loginUser.Username), utils.GetFuncName())
	//check if user exist in inmem
	user := passkey.Datastore.GetUser(loginUser.Username) // Find or create the new user

	options, session, err := passkey.WebAuthn.BeginRegistration(user)
	if err != nil {
		this.ResponseError("can't begin registration", utils.GetFuncName(), err)
		return
	}

	// Make a session key and store the sessionData values
	t := uuid.New().String()
	passkey.Datastore.SaveSession(t, *session)
	response := make(map[string]any)
	response["options"] = options
	response["sessionkey"] = t
	this.ResponseSuccessfullyWithAnyData(nil, "Begin registration successfully", utils.GetFuncName(), response)
}

// Logout handler
func (this *AuthController) Quit() {
	this.DestroySession()
	this.Redirect("/login", 302)
	this.StopRun()
}

// Logout handler
func (this *AuthController) IsLoggingOn() {
	user, isLogin := this.SimpleAuthCheck()
	if !isLogin {
		this.ResponseError("User is not logged in", utils.GetFuncName(), nil)
		return
	}
	this.ResponseSuccessfullyWithAnyData(user, fmt.Sprintf("LoginUser Id: %d", user.Id), utils.GetFuncName(), user.Id)
}

func (this *AuthController) GenRandomUsername() {
	username, err := utils.GetNewRandomUsername()
	if err != nil {
		this.ResponseError("Create new username failed", utils.GetFuncName(), err)
		return
	}
	result := make(map[string]string)
	result["username"] = username
	this.ResponseSuccessfullyWithAnyData(nil, fmt.Sprintf("Get random username: %s", username), utils.GetFuncName(), utils.ObjectToJsonString(result))
}

func (this *AuthController) BeginRegistration() {
	logpack.Info("begin registration ----------------------", utils.GetFuncName())
	username := this.GetString("username")
	if utils.IsEmpty(username) {
		this.ResponseError("Username param failed", utils.GetFuncName(), nil)
		return
	}
	o := orm.NewOrm()
	//Check the user exists with username
	userExist, existErr := utils.CheckUserExist(username, o)
	if existErr != nil {
		this.ResponseError("Check exist user on DB failed", utils.GetFuncName(), existErr)
		return
	}
	//if user exist on DB
	if userExist {
		this.ResponseError("User already exists on DB", utils.GetFuncName(), nil)
		return
	}
	//check if user exist in inmem
	exist := passkey.Datastore.CheckExistUser(username)
	if exist {
		this.ResponseError("User has been registered on passkey manager", utils.GetFuncName(), nil)
		return
	}

	user := passkey.Datastore.GetUser(username) // Find or create the new user

	options, session, err := passkey.WebAuthn.BeginRegistration(user)
	if err != nil {
		this.ResponseError("can't begin registration", utils.GetFuncName(), err)
		passkey.Datastore.RemoveUser(username)
		return
	}

	// Make a session key and store the sessionData values
	t := uuid.New().String()
	passkey.Datastore.SaveSession(t, *session)
	response := make(map[string]any)
	response["options"] = options
	response["sessionkey"] = t
	this.ResponseSuccessfullyWithAnyData(nil, "Begin registration successfully", utils.GetFuncName(), response)
}

func (this *AuthController) CancelRegister() {
	sessionKey := this.GetString("sessionKey")
	if utils.IsEmpty(sessionKey) {
		this.ResponseError("Session key is empty", utils.GetFuncName(), nil)
		return
	}
	// Get the session data stored from the function above
	session := passkey.Datastore.GetSession(sessionKey) // FIXME: cover invalid session
	//remove user by session
	passkey.Datastore.RemoveUser(string(session.UserID))
	this.ResponseSuccessfully(nil, "Remove session user successfully", utils.GetFuncName())
}

func (this *AuthController) FinishUpdatePasskey() {
	loginUser, check := this.SimpleAuthCheck()
	if !check {
		this.ResponseError("Login authentication error. Please try again", utils.GetFuncName(), nil)
		return
	}
	// Get the session key from the header
	t := this.Ctx.Request.Header.Get("Session-Key")
	isResetStr := this.Ctx.Request.Header.Get("Is-Reset-Key")
	// Get the session data stored from the function above
	session := passkey.Datastore.GetSession(t) // FIXME: cover invalid session

	// In out example username == userID, but in real world it should be different
	user := passkey.Datastore.GetUser(string(session.UserID)) // Get the user
	username := user.WebAuthnName()
	if username != loginUser.Username {
		this.ResponseError("Username does not match", utils.GetFuncName(), nil)
		return
	}

	credential, err := passkey.WebAuthn.FinishRegistration(user, session, this.Ctx.Request)
	if err != nil {
		this.ResponseError("can't finish registration", utils.GetFuncName(), err)
		return
	}
	// If creation was successful, store the credential object
	//replace current key with new credential
	if isResetStr == "true" {
		user.ReplaceCredential(credential)
	} else {
		user.AddCredential(credential)
	}
	passkey.Datastore.SaveUser(user)
	// Delete the session data
	passkey.Datastore.DeleteSession(t)
	//update user
	o := orm.NewOrm()
	//get user from db
	updateUser, upUserErr := utils.GetUserByUsername(loginUser.Username, o)
	if upUserErr != nil {
		this.ResponseError("Get update user failed", utils.GetFuncName(), upUserErr)
		return
	}
	tx, beginErr := o.Begin()
	if beginErr != nil {
		this.ResponseError("An error has occurred. Please try again!", utils.GetFuncName(), beginErr)
		return
	}
	updateUser.CredsArrJson = user.GetUserCredsJson()
	updateUser.Updatedt = time.Now().Unix()
	_, err2 := tx.Update(updateUser)
	if err2 != nil {
		tx.Rollback()
		this.ResponseError("Update user auth failed", utils.GetFuncName(), err2)
		return
	}
	tx.Commit()
	this.SetSession("LoginUser", passkey.InitSessionUser(*updateUser))
	this.SetSession("successMessage", "Update user passkey successfully")
	this.ResponseSuccessfully(nil, "Update user passkey successfully", utils.GetFuncName())
}

func (this *AuthController) FinishRegistration() {
	// Get the session key from the header
	t := this.Ctx.Request.Header.Get("Session-Key")
	// Get the session data stored from the function above
	session := passkey.Datastore.GetSession(t) // FIXME: cover invalid session

	// In out example username == userID, but in real world it should be different
	user := passkey.Datastore.GetUser(string(session.UserID)) // Get the user
	username := user.WebAuthnName()
	//check username exist on DB again
	o := orm.NewOrm()
	userExist, existErr := utils.CheckUserExist(username, o)
	if existErr != nil {
		this.ResponseError("Check exist username on DB failed", utils.GetFuncName(), existErr)
		return
	}

	if userExist {
		this.ResponseError("Username already exists. Unable to register", utils.GetFuncName(), nil)
		return
	}

	credential, err := passkey.WebAuthn.FinishRegistration(user, session, this.Ctx.Request)
	if err != nil {
		this.ResponseError("can't finish registration", utils.GetFuncName(), err)
		return
	}
	// If creation was successful, store the credential object
	user.AddCredential(credential)
	passkey.Datastore.SaveUser(user)
	// Delete the session data
	passkey.Datastore.DeleteSession(t)
	//register new user
	tx, beginErr := o.Begin()
	if beginErr != nil {
		this.ResponseError("An error has occurred. Please try again!", utils.GetFuncName(), beginErr)
		return
	}
	//Create new token for user
	token, _ := utils.CreateNewUserToken()
	insertUser := models.User{
		Username:     username,
		Status:       int(utils.StatusActive),
		Role:         int(utils.RoleRegular),
		Createdt:     time.Now().Unix(),
		Token:        token,
		CredsArrJson: user.GetUserCredsJson(),
	}
	insertUser.Updatedt = insertUser.Createdt
	insertUser.LastLogindt = insertUser.Createdt

	_, err2 := tx.Insert(&insertUser)
	if err2 != nil {
		tx.Rollback()
		this.ResponseError("User creation error. Try again!", utils.GetFuncName(), err2)
		return
	}
	tx.Commit()
	//set user session
	this.SetSession("LoginUser", passkey.InitSessionUser(insertUser))
	this.ResponseSuccessfully(nil, "Finish registration successfully", utils.GetFuncName())
}

func (this *AuthController) AssertionOptions() {
	options, sessionData, err := passkey.WebAuthn.BeginDiscoverableLogin()
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), nil)
		return
	}
	t := uuid.New().String()
	passkey.Datastore.SaveSession(t, *sessionData)
	response := make(map[string]any)
	response["options"] = options
	response["sessionkey"] = t
	response["hasUser"] = passkey.Datastore.CheckHasUser()
	this.ResponseSuccessfullyWithAnyData(nil, "Begin registration successfully", utils.GetFuncName(), response)
}

func (this *AuthController) AssertionResult() {
	// Get the session key from the header
	t := this.Ctx.Request.Header.Get("Session-Key")
	// Get the session data stored from the function above
	session := passkey.Datastore.GetSession(t) // FIXME: cover invalid session
	//userKey
	var passkeyUser passkey.PasskeyUser
	o := orm.NewOrm()
	var username string
	credential, err := passkey.WebAuthn.FinishDiscoverableLogin(func(rawId []byte, userhandle []byte) (user webauthn.User, err error) {
		// check userHandle
		username = string(userhandle)
		//check username exist on DB again
		userExist, existErr := utils.CheckUserExist(username, o)
		if existErr != nil {
			err = fmt.Errorf("%s", "Check exist user on DB failed")
			return
		}

		if !userExist {
			err = fmt.Errorf("%s", "Username does not exists. Unable to login")
			return
		}

		passkeyUser = passkey.Datastore.GetUser(username)
		if passkeyUser == nil {
			passkey.Datastore.RemoveUser(string(userhandle))
			err = fmt.Errorf("%s", "Passkey user does not exists")
			return
		}
		tmpUser := &passkey.User{
			ID:       passkeyUser.WebAuthnID(),
			Username: passkeyUser.WebAuthnName(),
		}
		tmpUser.SetCredential(passkeyUser.WebAuthnCredentials())
		user = tmpUser
		return
	}, session, this.Ctx.Request)
	if err != nil {
		this.ResponseError("Authentication failed. Login unsuccessful", utils.GetFuncName(), err)
		return
	}
	if credential.Authenticator.CloneWarning {
		logpack.Warn("can't finish login", utils.GetFuncName())
	}
	passkeyUser.UpdateCredential(credential)
	passkey.Datastore.SaveUser(passkeyUser)
	passkey.Datastore.DeleteSession(t)
	//get user from username
	loginUser, err := utils.GetUserByUsername(username, o)
	if err != nil {
		this.ResponseError(fmt.Sprintf("Get user from DB failed. Identifier: %s", passkeyUser.WebAuthnName()), utils.GetFuncName(), err)
		return
	}
	loginUser.CredsArrJson = passkeyUser.GetUserCredsJson()
	//update loginUser
	tx, beginErr := o.Begin()
	if beginErr != nil {
		this.ResponseError("An error has occurred. Please try again!", utils.GetFuncName(), beginErr)
		return
	}
	loginUser.Updatedt = time.Now().Unix()
	loginUser.LastLogindt = loginUser.Updatedt
	_, updateErr := tx.Update(loginUser)
	if updateErr != nil {
		this.ResponseError("Update credential for login user failed", utils.GetFuncName(), updateErr)
		return
	}
	tx.Commit()
	this.SetSession("LoginUser", passkey.InitSessionUser(*loginUser))
	this.ResponseSuccessfully(nil, "Finish login by passkey successfully", utils.GetFuncName())
}

func (this *AuthController) BeginConfirmPasskey() {
	loginUser, check := this.SimpleAuthCheck()
	if !check {
		this.ResponseError("Login authentication error. Please try again", utils.GetFuncName(), nil)
		return
	}
	logpack.Info(fmt.Sprintf("begin confirm passkey for user: %s"), utils.GetFuncName())
	//check if user exist in inmem
	exist := passkey.Datastore.CheckExistUser(loginUser.Username)
	if !exist {
		this.ResponseError("Username has not been registered on passkey data", utils.GetFuncName(), nil)
		return
	}
	user := passkey.Datastore.GetUser(loginUser.Username) // Find the user
	options, session, err := passkey.WebAuthn.BeginLogin(user)
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), nil)
		return
	}

	// Make a session key and store the sessionData values
	t := uuid.New().String()
	passkey.Datastore.SaveSession(t, *session)

	response := make(map[string]any)
	response["options"] = options
	response["sessionkey"] = t
	this.ResponseSuccessfullyWithAnyData(nil, "Begin registration successfully", utils.GetFuncName(), response)
}

func (this *AuthController) FinishConfirmPasskey() {
	loginUser, check := this.SimpleAuthCheck()
	if !check {
		this.ResponseError("Login authentication error. Please try again", utils.GetFuncName(), nil)
		return
	}
	// Get the session key from the header
	t := this.Ctx.Request.Header.Get("Session-Key")
	session := passkey.Datastore.GetSession(t) // FIXME: cover invalid session

	// In out example username == userID, but in real world it should be different
	user := passkey.Datastore.GetUser(string(session.UserID)) // Get the user
	//regist user to DB
	username := user.WebAuthnName()
	if username != loginUser.Username {
		this.ResponseError("The authentication username information does not match", utils.GetFuncName(), nil)
		return
	}

	credential, err := passkey.WebAuthn.FinishLogin(user, session, this.Ctx.Request)
	if err != nil {
		this.ResponseError("can't finish confirm passkey", utils.GetFuncName(), err)
		return
	}

	// Handle credential.Authenticator.CloneWarning
	if credential.Authenticator.CloneWarning {
		logpack.Warn("can't finish confirm passkey", utils.GetFuncName())
	}

	// If login was successful, update the credential object
	user.UpdateCredential(credential)
	passkey.Datastore.SaveUser(user)
	// Delete the session data
	passkey.Datastore.DeleteSession(t)
	this.ResponseSuccessfully(nil, "Finish confirm passkey succefully", utils.GetFuncName())
}
