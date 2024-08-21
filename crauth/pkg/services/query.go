package services

import (
	"crauth/pkg/models"
	"crauth/pkg/pb"
	"crauth/pkg/utils"
	"fmt"
	"time"

	"gorm.io/gorm"
)

func (s *Server) GetAdminUserList(reqData *pb.RequestData) *pb.ResponseData {
	authClaims, isLogin := s.Jwt.HanlderCheckLogin(reqData.AuthToken)
	if !isLogin {
		return pb.ResponseError("User is not login", utils.GetFuncName(), fmt.Errorf("User is not login"))
	}
	//if is not superadmin, ignore
	if authClaims.Role != int(utils.RoleSuperAdmin) {
		return pb.ResponseError("There is no permission to access this feature", utils.GetFuncName(), fmt.Errorf("There is no permission to access this feature"))
	}
	userList, listErr := s.H.GetUserListWithExcludeId(authClaims.Id)
	if listErr != nil && listErr != gorm.ErrRecordNotFound {
		return pb.ResponseError("Get user list failed", utils.GetFuncName(), fmt.Errorf("Get user list failed"))
	}
	return pb.ResponseSuccessfullyWithAnyData(0, "Get user list successfully", utils.GetFuncName(), userList)
}

func (s *Server) GetUserInfoByUsername(reqData *pb.RequestData) *pb.ResponseData {
	usernameAny, usernameExist := reqData.DataMap["username"]
	if !usernameExist {
		return pb.ResponseError("Param not found", utils.GetFuncName(), nil)
	}
	username := usernameAny.(string)
	user, err := s.H.GetUserByUsername(username)
	if err != nil {
		return pb.ResponseError("Get user by username failed", utils.GetFuncName(), err)
	}
	return pb.ResponseSuccessfullyWithAnyData(0, "Get user info successfully", utils.GetFuncName(), models.UserInfo{
		Id:       user.Id,
		Username: user.Username,
		Role:     user.Role,
	})
}

func (s *Server) GetAdminUserInfo(reqData *pb.RequestData) *pb.ResponseData {
	authClaims, isLogin := s.Jwt.HanlderCheckLogin(reqData.AuthToken)
	if !isLogin {
		return pb.ResponseError("User is not login", utils.GetFuncName(), fmt.Errorf("User is not login"))
	}
	//if is not superadmin, ignore
	if authClaims.Role != int(utils.RoleSuperAdmin) {
		return pb.ResponseError("There is no permission to access this feature", utils.GetFuncName(), fmt.Errorf("There is no permission to access this feature"))
	}
	userIdAny, exist := reqData.DataMap["userId"]
	if !exist {
		return pb.ResponseError("User id param not found", utils.GetFuncName(), nil)
	}
	userId := userIdAny.(int64)
	//get user by id
	user, err := s.H.GetUserFromId(userId)
	if err != nil {
		return pb.ResponseError("Retrieve user data failed", utils.GetFuncName(), err)
	}
	return pb.ResponseSuccessfullyWithAnyData(0, "Get admin user info successfully", utils.GetFuncName(), user)
}

func (s *Server) GetExcludeLoginUserNameList(reqData *pb.RequestData) *pb.ResponseData {
	authClaims, isLogin := s.Jwt.HanlderCheckLogin(reqData.AuthToken)
	if !isLogin {
		return pb.ResponseError("User is not login", utils.GetFuncName(), fmt.Errorf("User is not login"))
	}

	listName := s.H.GetUsernameListExcludeId(authClaims.Id)
	return pb.ResponseSuccessfullyWithAnyData(authClaims.Id, "Get user name list successfully", utils.GetFuncName(), listName)
}

func (s *Server) ChangeUserStatus(reqData *pb.RequestData) *pb.ResponseData {
	authClaims, isLogin := s.Jwt.HanlderCheckLogin(reqData.AuthToken)
	if !isLogin {
		return pb.ResponseError("User is not login", utils.GetFuncName(), fmt.Errorf("User is not login"))
	}
	//if is not superadmin, ignore
	if authClaims.Role != int(utils.RoleSuperAdmin) {
		return pb.ResponseError("There is no permission to access this feature", utils.GetFuncName(), fmt.Errorf("There is no permission to access this feature"))
	}
	userIdAny, idExist := reqData.DataMap["userId"]
	activeAny, activeExist := reqData.DataMap["active"]
	if !idExist || !activeExist {
		return pb.ResponseError("Param not found", utils.GetFuncName(), nil)
	}

	userIdParam := userIdAny.(int64)
	activeFlg := activeAny.(int)
	user, err := s.H.GetUserFromId(userIdParam)
	if err != nil {
		return pb.ResponseLoginError(authClaims.Id, "Get user from DB error. Please try again!", utils.GetFuncName(), err)
	}

	tx := s.H.DB.Begin()
	user.Status = activeFlg
	user.Updatedt = time.Now().Unix()
	//update user
	updateErr := tx.Save(&user).Error
	if updateErr != nil {
		return pb.ResponseLoginRollbackError(authClaims.Id, tx, "Update User failed. Please try again!", utils.GetFuncName(), updateErr)
	}
	tx.Commit()
	return pb.ResponseSuccessfully(authClaims.Id, "Update User successfully!", utils.GetFuncName())
}

// login check
func (s *Server) IsLoggingOn(reqData *pb.RequestData) *pb.ResponseData {
	authClaims, isLogin := s.Jwt.HanlderCheckLogin(reqData.AuthToken)
	if !isLogin {
		return pb.ResponseError("User is not login", utils.GetFuncName(), fmt.Errorf("User is not login"))
	}
	return pb.ResponseSuccessfullyWithAnyData(0, fmt.Sprintf("LoginUser Id: %d", authClaims.Id), utils.GetFuncName(), authClaims)
}
func (s *Server) GenRandomUsername(reqData *pb.RequestData) *pb.ResponseData {
	username, err := s.H.GetNewRandomUsername()
	if err != nil {
		return pb.ResponseError("Create new username failed", utils.GetFuncName(), err)
	}
	result := make(map[string]string)
	result["username"] = username
	return pb.ResponseSuccessfullyWithAnyData(0, fmt.Sprintf("Get random username: %s", username), utils.GetFuncName(), utils.ObjectToJsonString(result))
}
