package services

import (
	"crauth/pkg/db"
	"crauth/pkg/logpack"
	"crauth/pkg/models"
	"crauth/pkg/passkey"
	"crauth/pkg/pb"
	"crauth/pkg/utils"
	"fmt"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

type Server struct {
	H   db.Handler       // handler
	Jwt utils.JWTWrapper // jwt wrapper
	pb.UnimplementedAuthServiceServer
}

func (s *Server) BeginRegistration(reqData *pb.RequestData) *pb.ResponseData {
	logpack.Info("begin registration ----------------------", utils.GetFuncName())
	usernameAny, paramExist := reqData.DataMap["username"]
	if !paramExist {
		return pb.ResponseError("Username param not found", utils.GetFuncName(), nil)
	}
	username := usernameAny.(string)

	var userObj models.User
	if utils.IsEmpty(username) {
		return pb.ResponseError("Username param failed", utils.GetFuncName(), nil)
	}
	if result := s.H.DB.Where(&models.User{Username: username}).First(&userObj); result.Error == nil {
		return pb.ResponseError("Username already exists", utils.GetFuncName(), nil)
	}

	//check if user exist in inmem
	exist := passkey.Datastore.CheckExistUser(username)
	if exist {
		return pb.ResponseError("User has been registered on passkey manager", utils.GetFuncName(), nil)
	}

	user := passkey.Datastore.GetUser(username) // Find or create the new user

	options, session, err := passkey.WebAuthn.BeginRegistration(user)
	if err != nil {
		passkey.Datastore.RemoveUser(username)
		return pb.ResponseError("can't begin registration", utils.GetFuncName(), nil)
	}
	t := uuid.New().String()
	passkey.Datastore.SaveSession(t, *session)
	response := make(map[string]any)
	response["options"] = options
	response["sessionkey"] = t
	return pb.ResponseSuccessfullyWithAnyData(0, "Begin registration successfully", utils.GetFuncName(), response)
}

func (s *Server) CancelRegister(reqData *pb.RequestData) *pb.ResponseData {
	sesKey, isExist := reqData.DataMap["sessionKey"]
	if !isExist {
		return pb.ResponseError("Session key param not found", utils.GetFuncName(), nil)
	}
	sessionKey := sesKey.(string)
	if utils.IsEmpty(sessionKey) {
		return pb.ResponseError("Session key is empty", utils.GetFuncName(), nil)
	}
	// Get the session data stored from the function above
	session := passkey.Datastore.GetSession(sessionKey) // FIXME: cover invalid session
	//remove user by session
	passkey.Datastore.RemoveUser(string(session.UserID))
	return pb.ResponseSuccessfully(0, "Remove session user successfully", utils.GetFuncName())
}

func (s *Server) FinishUpdatePasskey(reqData *pb.RequestData) *pb.ResponseData {
	sesKeyAny, sesExist := reqData.DataMap["sessionKey"]
	isResetAny, isResetExist := reqData.DataMap["isResetKey"]
	if !sesExist || !isResetExist {
		return pb.ResponseError("Param failed. Please try again", utils.GetFuncName(), nil)
	}
	sessionKey := sesKeyAny.(string)
	isResetKey := isResetAny.(bool)
	loginUser, check := s.Jwt.HanlderCheckLogin(reqData.AuthToken)
	if !check {
		return pb.ResponseError("Login authentication error. Please try again", utils.GetFuncName(), nil)
	}
	// Get the session data stored from the function above
	session := passkey.Datastore.GetSession(sessionKey) // FIXME: cover invalid session

	// In out example username == userID, but in real world it should be different
	user := passkey.Datastore.GetUser(string(session.UserID)) // Get the user
	username := user.WebAuthnName()
	if username != loginUser.Username {
		return pb.ResponseError("Username does not match", utils.GetFuncName(), nil)
	}

	credential, err := passkey.WebAuthn.FinishRegistration(user, session, &reqData.Request)
	if err != nil {
		return pb.ResponseError("can't finish registration", utils.GetFuncName(), err)
	}
	// If creation was successful, store the credential object
	//replace current key with new credential
	if isResetKey {
		user.ReplaceCredential(credential)
	} else {
		user.AddCredential(credential)
	}
	passkey.Datastore.SaveUser(user)
	// Delete the session data
	passkey.Datastore.DeleteSession(sessionKey)
	//update user
	//get user from db
	updateUser, upUserErr := s.H.GetUserByUsername(loginUser.Username)
	if upUserErr != nil {
		return pb.ResponseError("Get update user failed", utils.GetFuncName(), upUserErr)
	}
	tx := s.H.DB.Begin()
	updateUser.CredsArrJson = user.GetUserCredsJson()
	updateUser.Updatedt = time.Now().Unix()
	err2 := tx.Save(updateUser).Error
	if err2 != nil {
		tx.Rollback()
		return pb.ResponseError("Update user auth failed", utils.GetFuncName(), err2)
	}
	tx.Commit()
	//Login after registration
	tokenString, authClaim, err := s.Jwt.CreateAuthClaimSession(updateUser)
	if err != nil {
		return pb.ResponseError("Creating login session token failed", utils.GetFuncName(), err)
	}
	loginResponse := map[string]any{
		"token": tokenString,
		"user":  *authClaim,
	}
	return pb.ResponseSuccessfullyWithAnyData(0, "Update user passkey successfully", utils.GetFuncName(), loginResponse)
}

func (s *Server) FinishRegistration(reqData *pb.RequestData) *pb.ResponseData {
	sesKeyAny, sesExist := reqData.DataMap["sessionKey"]
	if !sesExist {
		return pb.ResponseError("Session key param not found", utils.GetFuncName(), nil)
	}
	sessionKey := sesKeyAny.(string)
	// Get the session data stored from the function above
	session := passkey.Datastore.GetSession(sessionKey) // FIXME: cover invalid session

	// In out example username == userID, but in real world it should be different
	user := passkey.Datastore.GetUser(string(session.UserID)) // Get the user
	username := user.WebAuthnName()
	//check username exist on DB again
	userExist, existErr := s.H.CheckUserExist(username)
	if existErr != nil {
		return pb.ResponseError("Check exist username on DB failed", utils.GetFuncName(), existErr)
	}

	if userExist {
		return pb.ResponseError("Username already exists. Unable to register", utils.GetFuncName(), nil)
	}

	credential, err := passkey.WebAuthn.FinishRegistration(user, session, &reqData.Request)
	if err != nil {
		return pb.ResponseError("can't finish registration", utils.GetFuncName(), err)
	}
	// If creation was successful, store the credential object
	user.AddCredential(credential)
	passkey.Datastore.SaveUser(user)
	// Delete the session data
	passkey.Datastore.DeleteSession(sessionKey)
	//register new user
	tx := s.H.DB.Begin()
	insertUser := models.User{
		Username:     username,
		Status:       int(utils.StatusActive),
		Role:         int(utils.RoleRegular),
		Createdt:     time.Now().Unix(),
		CredsArrJson: user.GetUserCredsJson(),
	}
	insertUser.Updatedt = insertUser.Createdt
	insertUser.LastLogindt = insertUser.Createdt

	err2 := tx.Create(&insertUser).Error
	if err2 != nil {
		tx.Rollback()
		return pb.ResponseError("User creation error. Try again!", utils.GetFuncName(), err2)
	}
	tx.Commit()
	//Login after registration
	tokenString, authClaim, err := s.Jwt.CreateAuthClaimSession(&insertUser)
	if err != nil {
		return pb.ResponseError("Creating login session token failed", utils.GetFuncName(), err)
	}
	loginResponse := map[string]any{
		"token": tokenString,
		"user":  *authClaim,
	}
	return pb.ResponseSuccessfullyWithAnyData(0, "Finish registration successfully", utils.GetFuncName(), loginResponse)
}

func (s *Server) AssertionOptions() *pb.ResponseData {
	options, sessionData, err := passkey.WebAuthn.BeginDiscoverableLogin()
	if err != nil {
		return pb.ResponseError(err.Error(), utils.GetFuncName(), nil)
	}
	t := uuid.New().String()
	passkey.Datastore.SaveSession(t, *sessionData)
	response := make(map[string]any)
	response["options"] = options
	response["sessionkey"] = t
	response["hasUser"] = passkey.Datastore.CheckHasUser()
	return pb.ResponseSuccessfullyWithAnyData(0, "Begin registration successfully", utils.GetFuncName(), response)
}

func (s *Server) AssertionResult(reqData *pb.RequestData) *pb.ResponseData {
	sesKeyAny, sesExist := reqData.DataMap["sessionKey"]
	if !sesExist {
		return pb.ResponseError("Session key param not found", utils.GetFuncName(), nil)
	}
	sessionKey := sesKeyAny.(string)
	// Get the session data stored from the function above
	session := passkey.Datastore.GetSession(sessionKey) // FIXME: cover invalid session
	//userKey
	var passkeyUser passkey.PasskeyUser
	var username string
	credential, err := passkey.WebAuthn.FinishDiscoverableLogin(func(rawId []byte, userhandle []byte) (user webauthn.User, err error) {
		// check userHandle
		username = string(userhandle)
		//check username exist on DB again
		userExist, existErr := s.H.CheckUserExist(username)
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
	}, session, &reqData.Request)
	if err != nil {
		return pb.ResponseError("Authentication failed. Login unsuccessful", utils.GetFuncName(), err)
	}
	if credential.Authenticator.CloneWarning {
		logpack.Warn("can't finish login", utils.GetFuncName())
	}
	passkeyUser.UpdateCredential(credential)
	passkey.Datastore.SaveUser(passkeyUser)
	passkey.Datastore.DeleteSession(sessionKey)
	//get user from username
	loginUser, err := s.H.GetUserByUsername(username)
	if err != nil {
		return pb.ResponseError(fmt.Sprintf("Get user from DB failed. Identifier: %s", passkeyUser.WebAuthnName()), utils.GetFuncName(), err)
	}
	loginUser.CredsArrJson = passkeyUser.GetUserCredsJson()
	//update loginUser
	tx := s.H.DB.Begin()
	loginUser.Updatedt = time.Now().Unix()
	loginUser.LastLogindt = loginUser.Updatedt
	updateErr := tx.Save(loginUser).Error
	if updateErr != nil {
		return pb.ResponseError("Update credential for login user failed", utils.GetFuncName(), updateErr)
	}
	tx.Commit()
	tokenString, authClaim, err := s.Jwt.CreateAuthClaimSession(loginUser)
	if err != nil {
		return pb.ResponseError("Creating login session token failed", utils.GetFuncName(), err)
	}
	loginResponse := map[string]any{
		"token": tokenString,
		"user":  *authClaim,
	}
	return pb.ResponseSuccessfullyWithAnyData(0, "Finish login by passkey successfully", utils.GetFuncName(), loginResponse)
}

func (s *Server) BeginConfirmPasskey(reqData *pb.RequestData) *pb.ResponseData {
	loginUser, check := s.Jwt.HanlderCheckLogin(reqData.AuthToken)
	if !check {
		return pb.ResponseError("Login authentication error. Please try again", utils.GetFuncName(), nil)
	}
	logpack.Info(fmt.Sprintf("begin confirm passkey for user: %s"), utils.GetFuncName())
	//check if user exist in inmem
	exist := passkey.Datastore.CheckExistUser(loginUser.Username)
	if !exist {
		return pb.ResponseError("Username has not been registered on passkey data", utils.GetFuncName(), nil)
	}
	user := passkey.Datastore.GetUser(loginUser.Username) // Find the user
	options, session, err := passkey.WebAuthn.BeginLogin(user)
	if err != nil {
		return pb.ResponseError(err.Error(), utils.GetFuncName(), nil)
	}

	// Make a session key and store the sessionData values
	t := uuid.New().String()
	passkey.Datastore.SaveSession(t, *session)

	response := make(map[string]any)
	response["options"] = options
	response["sessionkey"] = t
	return pb.ResponseSuccessfullyWithAnyData(0, "Begin registration successfully", utils.GetFuncName(), response)
}

func (s *Server) FinishConfirmPasskey(req *pb.RequestData) *pb.ResponseData {
	sesKeyAny, sesExist := req.DataMap["sessionKey"]
	if !sesExist {
		return pb.ResponseError("Session key param not found", utils.GetFuncName(), nil)
	}
	sessionKey := sesKeyAny.(string)
	loginUser, check := s.Jwt.HanlderCheckLogin(req.AuthToken)
	if !check {
		return pb.ResponseError("Login authentication error. Please try again", utils.GetFuncName(), nil)
	}
	// Get the session key from the header
	session := passkey.Datastore.GetSession(sessionKey) // FIXME: cover invalid session

	// In out example username == userID, but in real world it should be different
	user := passkey.Datastore.GetUser(string(session.UserID)) // Get the user
	//regist user to DB
	username := user.WebAuthnName()
	if username != loginUser.Username {
		return pb.ResponseError("The authentication username information does not match", utils.GetFuncName(), nil)
	}

	credential, err := passkey.WebAuthn.FinishLogin(user, session, &req.Request)
	if err != nil {
		return pb.ResponseError("can't finish confirm passkey", utils.GetFuncName(), err)
	}

	// Handle credential.Authenticator.CloneWarning
	if credential.Authenticator.CloneWarning {
		logpack.Warn("can't finish confirm passkey", utils.GetFuncName())
	}

	// If login was successful, update the credential object
	user.UpdateCredential(credential)
	passkey.Datastore.SaveUser(user)
	// Delete the session data
	passkey.Datastore.DeleteSession(sessionKey)
	return pb.ResponseSuccessfully(0, "Finish confirm passkey succefully", utils.GetFuncName())
}

func (s *Server) ChangeUsernameFinish(reqData *pb.RequestData) *pb.ResponseData {
	sesKeyAny, sesExist := reqData.DataMap["sessionKey"]
	oldUsernameAny, uExist := reqData.DataMap["oldUsername"]
	if !sesExist || !uExist {
		return pb.ResponseError("Param not found", utils.GetFuncName(), nil)
	}
	sessionKey := sesKeyAny.(string)
	oldUsername := oldUsernameAny.(string)
	// Get the session data stored from the function above
	session := passkey.Datastore.GetSession(sessionKey) // FIXME: cover invalid session
	// In out example username == userID, but in real world it should be different
	user := passkey.Datastore.GetUser(string(session.UserID)) // Get the user
	userKey := user.WebAuthnName()
	//check username exist on DB again
	userExist, existErr := s.H.CheckUserExist(userKey)
	if existErr != nil {
		return pb.ResponseError("Check exist user on DB failed", utils.GetFuncName(), existErr)
	}
	if userExist {
		return pb.ResponseError("Username already exists. Unable to register", utils.GetFuncName(), nil)
	}
	credential, err := passkey.WebAuthn.FinishRegistration(user, session, &reqData.Request)
	if err != nil {
		return pb.ResponseError("can't finish registration", utils.GetFuncName(), err)
	}
	// If creation was successful, store the credential object
	user.AddCredential(credential)
	passkey.Datastore.SaveUser(user)
	// Delete the session data
	passkey.Datastore.DeleteSession(sessionKey)
	passkey.Datastore.RemoveUser(oldUsername)
	//register new user
	tx := s.H.DB.Begin()
	//get old user
	oldUser, oldUErr := s.H.GetUserByUsername(oldUsername)
	if oldUErr != nil {
		return pb.ResponseError("Get old user from DB failed", utils.GetFuncName(), oldUErr)
	}

	oldUser.Username = userKey
	oldUser.CredsArrJson = user.GetUserCredsJson()
	oldUser.Updatedt = time.Now().Unix()
	updateErr := tx.Save(oldUser).Error
	if updateErr != nil {
		tx.Rollback()
		return pb.ResponseError("Update username failed. Try again!", utils.GetFuncName(), updateErr)
	}
	tx.Commit()
	//Login after registration
	tokenString, authClaim, err := s.Jwt.CreateAuthClaimSession(oldUser)
	if err != nil {
		return pb.ResponseError("Creating login session token failed", utils.GetFuncName(), err)
	}
	loginResponse := map[string]any{
		"token": tokenString,
		"user":  *authClaim,
	}
	return pb.ResponseSuccessfullyWithAnyData(0, "Finish registration successfully", utils.GetFuncName(), loginResponse)
}

func (s *Server) SyncUsernameDB(reqData *pb.RequestData) *pb.ResponseData {
	newUsernameAny, newUExist := reqData.DataMap["newUsername"]
	oldUsernameAny, oldUExist := reqData.DataMap["oldUsername"]
	if !newUExist || !oldUExist {
		return pb.ResponseError("Param not found", utils.GetFuncName(), nil)
	}
	newUsername := newUsernameAny.(string)
	oldUsername := oldUsernameAny.(string)
	//check logging in
	authClaims, isLoggingin := s.Jwt.HanlderCheckLogin(reqData.AuthToken)
	if !isLoggingin || authClaims.Username != newUsername {
		return pb.ResponseError("Target user is not logging in", utils.GetFuncName(), fmt.Errorf("Target user is not logging in"))
	}
	go func() {
		//update contact on user table
		oldContactString := fmt.Sprintf("\"userName\":\"%s\"", oldUsername)
		newContactString := fmt.Sprintf("\"userName\":\"%s\"", newUsername)
		err := s.H.DB.Exec("UPDATE public.user SET contacts = REPLACE(contacts, ?, ?)", oldContactString, newContactString).Error
		if err != nil {
			logpack.Error("Sync user data failed", utils.GetFuncName(), err)
			return
		}
		logpack.Info("Sync username on all table successfully", utils.GetFuncName())
	}()
	return pb.ResponseSuccessfully(0, "Sync user table successfully", utils.GetFuncName())
}
