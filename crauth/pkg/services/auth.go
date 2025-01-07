package services

import (
	"context"
	"crauth/pkg/db"
	"crauth/pkg/logpack"
	"crauth/pkg/models"
	"crauth/pkg/passkey"
	"crauth/pkg/pb"
	"crauth/pkg/utils"
	"fmt"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

type Server struct {
	H   db.Handler       // handler
	Jwt utils.JWTWrapper // jwt wrapper
	pb.UnimplementedAuthServiceServer
}

func (s *Server) BeginRegistration(ctx context.Context, reqData *pb.WithUsernameRequest) (*pb.ResponseData, error) {
	logpack.Info("begin registration ----------------------", utils.GetFuncName())
	username := reqData.Username
	if utils.IsEmpty(username) {
		return ResponseError("Username param not found", utils.GetFuncName(), nil)
	}
	var userObj models.User
	if utils.IsEmpty(username) {
		return ResponseError("Username param failed", utils.GetFuncName(), nil)
	}
	if result := s.H.DB.Where(&models.User{Username: username}).First(&userObj); result.Error == nil {
		return ResponseError("Username already exists", utils.GetFuncName(), nil)
	}

	//check if user exist in inmem
	exist := passkey.Datastore.CheckExistUser(username)
	if exist {
		return ResponseError("User has been registered on passkey manager", utils.GetFuncName(), nil)
	}

	user := passkey.Datastore.GetUser(username) // Find or create the new user
	registerOptions := func(credCreationOpts *protocol.PublicKeyCredentialCreationOptions) {
		credCreationOpts.CredentialExcludeList = user.CredentialExcludeList()
		credCreationOpts.AuthenticatorSelection.ResidentKey = protocol.ResidentKeyRequirementRequired
		credCreationOpts.AuthenticatorSelection.RequireResidentKey = protocol.ResidentKeyRequired()
		credCreationOpts.AuthenticatorSelection.UserVerification = protocol.VerificationRequired
	}
	options, session, err := passkey.WebAuthn.BeginRegistration(user, registerOptions)
	if err != nil {
		passkey.Datastore.RemoveUser(username)
		return ResponseError("can't begin registration", utils.GetFuncName(), nil)
	}
	t := uuid.New().String()
	passkey.Datastore.SaveSession(t, *session)
	response := make(map[string]any)
	response["options"] = options
	response["sessionkey"] = t
	return ResponseSuccessfullyWithAnyData("", "Begin registration successfully", utils.GetFuncName(), response)
}

func (s *Server) CancelRegister(ctx context.Context, reqData *pb.CancelRegisterRequest) (*pb.ResponseData, error) {
	sessionKey := reqData.SessionKey
	if utils.IsEmpty(sessionKey) {
		return ResponseError("Session key is empty", utils.GetFuncName(), nil)
	}
	// Get the session data stored from the function above
	session := passkey.Datastore.GetSession(sessionKey) // FIXME: cover invalid session
	//remove user by session
	passkey.Datastore.RemoveUser(string(session.UserID))
	return ResponseSuccessfully("", "Remove session user successfully", utils.GetFuncName())
}

func (s *Server) BeginUpdatePasskey(ctx context.Context, reqData *pb.CommonRequest) (*pb.ResponseData, error) {
	//Get token
	loginUser, check := s.Jwt.HanlderCheckLogin(reqData.AuthToken)
	if !check {
		return ResponseError("Login authentication error. Please try again", utils.GetFuncName(), nil)
	}
	logpack.Info(fmt.Sprintf("begin update passkey of: %s", loginUser.Username), utils.GetFuncName())
	//check if user exist in inmem
	user := passkey.Datastore.GetUser(loginUser.Username) // Find or create the new user

	options, session, err := passkey.WebAuthn.BeginRegistration(user)
	if err != nil {
		return ResponseLoginError(loginUser.Username, "can't begin update passkey", utils.GetFuncName(), nil)
	}
	// Make a session key and store the sessionData values
	t := uuid.New().String()
	passkey.Datastore.SaveSession(t, *session)
	response := make(map[string]any)
	response["options"] = options
	response["sessionkey"] = t
	return ResponseSuccessfullyWithAnyData(loginUser.Username, "Begin registration successfully", utils.GetFuncName(), response)
}

func (s *Server) FinishUpdatePasskey(ctx context.Context, reqData *pb.FinishUpdatePasskeyRequest) (*pb.ResponseData, error) {
	sessionKey := reqData.SessionKey
	isResetKey := reqData.IsReset
	if utils.IsEmpty(sessionKey) {
		return ResponseError("Param failed. Please try again", utils.GetFuncName(), nil)
	}
	loginUser, check := s.Jwt.HanlderCheckLogin(reqData.Common.AuthToken)
	if !check {
		return ResponseError("Login authentication error. Please try again", utils.GetFuncName(), nil)
	}
	// Get the session data stored from the function above
	session := passkey.Datastore.GetSession(sessionKey) // FIXME: cover invalid session

	// In out example username == userID, but in real world it should be different
	user := passkey.Datastore.GetUser(string(session.UserID)) // Get the user
	username := user.WebAuthnName()
	if username != loginUser.Username {
		return ResponseLoginError(loginUser.Username, "Username does not match", utils.GetFuncName(), nil)
	}
	//init request data
	if reqData.Request == nil || utils.IsEmpty(reqData.Request.BodyJson) {
		return ResponseLoginError(loginUser.Username, "Request body failed", utils.GetFuncName(), nil)
	}

	request, parseRequestErr := utils.ConvertBodyJsonToRequest(reqData.Request.BodyJson)
	if parseRequestErr != nil {
		return ResponseLoginError(loginUser.Username, parseRequestErr.Error(), utils.GetFuncName(), parseRequestErr)
	}
	credential, err := passkey.WebAuthn.FinishRegistration(user, session, request)
	if err != nil {
		return ResponseLoginError(loginUser.Username, "can't finish registration", utils.GetFuncName(), err)
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
		return ResponseLoginError(loginUser.Username, "Get update user failed", utils.GetFuncName(), upUserErr)
	}
	tx := s.H.DB.Begin()
	updateUser.CredsArrJson = user.GetUserCredsJson()
	updateUser.Updatedt = time.Now().Unix()
	err2 := tx.Save(updateUser).Error
	if err2 != nil {
		tx.Rollback()
		return ResponseLoginError(loginUser.Username, "Update user auth failed", utils.GetFuncName(), err2)
	}
	tx.Commit()
	//Login after registration
	tokenString, authClaim, err := s.Jwt.CreateAuthClaimSession(updateUser)
	if err != nil {
		return ResponseLoginError(loginUser.Username, "Creating login session token failed", utils.GetFuncName(), err)
	}
	loginResponse := map[string]any{
		"token": tokenString,
		"user":  *authClaim,
	}
	return ResponseSuccessfullyWithAnyData(loginUser.Username, "Update user passkey successfully", utils.GetFuncName(), loginResponse)
}

func (s *Server) FinishRegistration(ctx context.Context, reqData *pb.SessionKeyAndHttpRequest) (*pb.ResponseData, error) {
	sessionKey := reqData.SessionKey
	if utils.IsEmpty(sessionKey) {
		return ResponseError("Session key param not found", utils.GetFuncName(), nil)
	}
	// Get the session data stored from the function above
	session := passkey.Datastore.GetSession(sessionKey) // FIXME: cover invalid session

	// In out example username == userID, but in real world it should be different
	user := passkey.Datastore.GetUser(string(session.UserID)) // Get the user
	username := user.WebAuthnName()
	//check username exist on DB again
	userExist, existErr := s.H.CheckUserExist(username)
	if existErr != nil {
		return ResponseError("Check exist username on DB failed", utils.GetFuncName(), existErr)
	}

	if userExist {
		return ResponseError("Username already exists. Unable to register", utils.GetFuncName(), nil)
	}
	request, parseRequestErr := utils.ConvertBodyJsonToRequest(reqData.Request.BodyJson)
	if parseRequestErr != nil {
		return ResponseError(parseRequestErr.Error(), utils.GetFuncName(), parseRequestErr)
	}
	credential, err := passkey.WebAuthn.FinishRegistration(user, session, request)
	if err != nil {
		return ResponseError("can't finish registration", utils.GetFuncName(), err)
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
		return ResponseError("User creation error. Try again!", utils.GetFuncName(), err2)
	}
	tx.Commit()
	//Login after registration
	tokenString, authClaim, err := s.Jwt.CreateAuthClaimSession(&insertUser)
	if err != nil {
		return ResponseError("Creating login session token failed", utils.GetFuncName(), err)
	}
	loginResponse := map[string]any{
		"token": tokenString,
		"user":  *authClaim,
	}
	return ResponseSuccessfullyWithAnyData("", "Finish registration successfully", utils.GetFuncName(), loginResponse)
}

func (s *Server) AssertionOptions(ctx context.Context, reqData *pb.CommonRequest) (*pb.ResponseData, error) {
	options, sessionData, err := passkey.WebAuthn.BeginDiscoverableLogin()
	if err != nil {
		return ResponseError(err.Error(), utils.GetFuncName(), nil)
	}
	t := uuid.New().String()
	passkey.Datastore.SaveSession(t, *sessionData)
	response := make(map[string]any)
	response["options"] = options
	response["sessionkey"] = t
	response["hasUser"] = passkey.Datastore.CheckHasUser()
	return ResponseSuccessfullyWithAnyData("", "Begin registration successfully", utils.GetFuncName(), response)
}

func (s *Server) AssertionResult(ctx context.Context, reqData *pb.SessionKeyAndHttpRequest) (*pb.ResponseData, error) {
	sessionKey := reqData.SessionKey
	if utils.IsEmpty(sessionKey) {
		return ResponseError("Session key param not found", utils.GetFuncName(), nil)
	}
	// Get the session data stored from the function above
	session := passkey.Datastore.GetSession(sessionKey) // FIXME: cover invalid session
	//userKey
	var passkeyUser passkey.PasskeyUser
	var username string
	request, parseRequestErr := utils.ConvertBodyJsonToRequest(reqData.Request.BodyJson)
	if parseRequestErr != nil {
		return ResponseError(parseRequestErr.Error(), utils.GetFuncName(), parseRequestErr)
	}
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
	}, session, request)
	if err != nil {
		return ResponseError("Authentication failed. Login unsuccessful", utils.GetFuncName(), err)
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
		return ResponseError(fmt.Sprintf("Get user from DB failed. Identifier: %s", passkeyUser.WebAuthnName()), utils.GetFuncName(), err)
	}
	loginUser.CredsArrJson = passkeyUser.GetUserCredsJson()
	//update loginUser
	tx := s.H.DB.Begin()
	loginUser.Updatedt = time.Now().Unix()
	loginUser.LastLogindt = loginUser.Updatedt
	updateErr := tx.Save(loginUser).Error
	if updateErr != nil {
		tx.Rollback()
		return ResponseError("Update credential for login user failed", utils.GetFuncName(), updateErr)
	}
	tx.Commit()
	tokenString, authClaim, err := s.Jwt.CreateAuthClaimSession(loginUser)
	if err != nil {
		return ResponseError("Creating login session token failed", utils.GetFuncName(), err)
	}
	loginResponse := map[string]any{
		"token": tokenString,
		"user":  *authClaim,
	}
	return ResponseSuccessfullyWithAnyData("", "Finish login by passkey successfully", utils.GetFuncName(), loginResponse)
}

func (s *Server) BeginConfirmPasskey(ctx context.Context, reqData *pb.CommonRequest) (*pb.ResponseData, error) {
	loginUser, check := s.Jwt.HanlderCheckLogin(reqData.AuthToken)
	if !check {
		return ResponseError("Login authentication error. Please try again", utils.GetFuncName(), nil)
	}
	logpack.Info(fmt.Sprintf("begin confirm passkey for user: %s"), utils.GetFuncName())
	//check if user exist in inmem
	exist := passkey.Datastore.CheckExistUser(loginUser.Username)
	if !exist {
		return ResponseError("Username has not been registered on passkey data", utils.GetFuncName(), nil)
	}
	user := passkey.Datastore.GetUser(loginUser.Username) // Find the user
	options, session, err := passkey.WebAuthn.BeginLogin(user)
	if err != nil {
		return ResponseError(err.Error(), utils.GetFuncName(), nil)
	}

	// Make a session key and store the sessionData values
	t := uuid.New().String()
	passkey.Datastore.SaveSession(t, *session)

	response := make(map[string]any)
	response["options"] = options
	response["sessionkey"] = t
	return ResponseSuccessfullyWithAnyData("", "Begin registration successfully", utils.GetFuncName(), response)
}

func (s *Server) FinishConfirmPasskey(ctx context.Context, reqData *pb.SessionKeyAndHttpRequest) (*pb.ResponseData, error) {
	sessionKey := reqData.SessionKey
	if utils.IsEmpty(sessionKey) {
		return ResponseError("Session key param not found", utils.GetFuncName(), nil)
	}
	loginUser, check := s.Jwt.HanlderCheckLogin(reqData.Common.AuthToken)
	if !check {
		return ResponseError("Login authentication error. Please try again", utils.GetFuncName(), nil)
	}
	// Get the session key from the header
	session := passkey.Datastore.GetSession(sessionKey) // FIXME: cover invalid session

	// In out example username == userID, but in real world it should be different
	user := passkey.Datastore.GetUser(string(session.UserID)) // Get the user
	//regist user to DB
	username := user.WebAuthnName()
	if username != loginUser.Username {
		return ResponseError("The authentication username information does not match", utils.GetFuncName(), nil)
	}
	//init request body
	request, parseRequestErr := utils.ConvertBodyJsonToRequest(reqData.Request.BodyJson)
	if parseRequestErr != nil {
		return ResponseError(parseRequestErr.Error(), utils.GetFuncName(), parseRequestErr)
	}
	credential, err := passkey.WebAuthn.FinishLogin(user, session, request)
	if err != nil {
		return ResponseError("can't finish confirm passkey", utils.GetFuncName(), err)
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
	return ResponseSuccessfully("", "Finish confirm passkey succefully", utils.GetFuncName())
}

func (s *Server) ChangeUsernameFinish(ctx context.Context, reqData *pb.ChangeUsernameFinishRequest) (*pb.ResponseData, error) {
	sessionKey := reqData.SessionKey
	oldUsername := reqData.OldUsername
	if utils.IsEmpty(sessionKey) || utils.IsEmpty(oldUsername) {
		return ResponseError("Param not found", utils.GetFuncName(), nil)
	}
	// Get the session data stored from the function above
	session := passkey.Datastore.GetSession(sessionKey) // FIXME: cover invalid session
	// In out example username == userID, but in real world it should be different
	user := passkey.Datastore.GetUser(string(session.UserID)) // Get the user
	userKey := user.WebAuthnName()
	//check username exist on DB again
	userExist, existErr := s.H.CheckUserExist(userKey)
	if existErr != nil {
		return ResponseError("Check exist user on DB failed", utils.GetFuncName(), existErr)
	}
	if userExist {
		return ResponseError("Username already exists. Unable to register", utils.GetFuncName(), nil)
	}
	//init request body
	request, parseRequestErr := utils.ConvertBodyJsonToRequest(reqData.Request.BodyJson)
	if parseRequestErr != nil {
		return ResponseError(parseRequestErr.Error(), utils.GetFuncName(), parseRequestErr)
	}
	credential, err := passkey.WebAuthn.FinishRegistration(user, session, request)
	if err != nil {
		return ResponseError("can't finish registration", utils.GetFuncName(), err)
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
		tx.Rollback()
		return ResponseError("Get old user from DB failed", utils.GetFuncName(), oldUErr)
	}

	oldUser.Username = userKey
	oldUser.CredsArrJson = user.GetUserCredsJson()
	oldUser.Updatedt = time.Now().Unix()
	updateErr := tx.Save(oldUser).Error
	if updateErr != nil {
		tx.Rollback()
		return ResponseError("Update username failed. Try again!", utils.GetFuncName(), updateErr)
	}
	//Login after registration
	tokenString, authClaim, err := s.Jwt.CreateAuthClaimSession(oldUser)
	if err != nil {
		tx.Rollback()
		return ResponseError("Creating login session token failed", utils.GetFuncName(), err)
	}
	tx.Commit()
	loginResponse := map[string]any{
		"token": tokenString,
		"user":  *authClaim,
	}
	return ResponseSuccessfullyWithAnyData("", "Finish registration successfully", utils.GetFuncName(), loginResponse)
}

func (s *Server) SyncUsernameDB(ctx context.Context, reqData *pb.SyncUsernameDBRequest) (*pb.ResponseData, error) {
	newUsername := reqData.NewUsername
	oldUsername := reqData.OldUsername
	if utils.IsEmpty(newUsername) || utils.IsEmpty(oldUsername) {
		return ResponseError("Get Param not found", utils.GetFuncName(), nil)
	}
	//check logging in
	authClaims, isLoggingin := s.Jwt.HanlderCheckLogin(reqData.Common.AuthToken)
	if !isLoggingin || authClaims.Username != newUsername {
		return ResponseError("Target user is not logging in", utils.GetFuncName(), fmt.Errorf("Target user is not logging in"))
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
	return ResponseSuccessfully("", "Sync user table successfully", utils.GetFuncName())
}
