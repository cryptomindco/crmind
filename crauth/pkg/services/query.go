package services

import (
	"context"
	"crauth/pkg/models"
	"crauth/pkg/pb"
	"crauth/pkg/utils"
	"fmt"
	"time"

	"gorm.io/gorm"
)

func (s *Server) GetAdminUserList(ctx context.Context, reqData *pb.CommonRequest) (*pb.ResponseData, error) {
	authClaims, isLogin := s.Jwt.HanlderCheckLogin(reqData.AuthToken)
	if !isLogin {
		return ResponseError("User is not login", utils.GetFuncName(), fmt.Errorf("User is not login"))
	}
	//if is not superadmin, ignore
	if authClaims.Role != int(utils.RoleSuperAdmin) {
		return ResponseError("There is no permission to access this feature", utils.GetFuncName(), fmt.Errorf("There is no permission to access this feature"))
	}
	userList, listErr := s.H.GetUserListWithExcludeId(authClaims.Id)
	if listErr != nil && listErr != gorm.ErrRecordNotFound {
		return ResponseError("Get user list failed", utils.GetFuncName(), fmt.Errorf("Get user list failed"))
	}
	return ResponseSuccessfullyWithAnyData(authClaims.Username, "Get user list successfully", utils.GetFuncName(), userList)
}

func (s *Server) GetUserInfoByUsername(ctx context.Context, reqData *pb.WithUsernameRequest) (*pb.ResponseData, error) {
	username := reqData.Username
	if utils.IsEmpty(username) {
		return ResponseError("Param not found", utils.GetFuncName(), nil)
	}
	user, err := s.H.GetUserByUsername(username)
	if err != nil {
		return ResponseError("Get user by username failed", utils.GetFuncName(), err)
	}
	return ResponseSuccessfullyWithAnyData("", "Get user info successfully", utils.GetFuncName(), models.UserInfo{
		Id:       user.Id,
		Username: user.Username,
		Role:     user.Role,
	})
}

func (s *Server) GetAdminUserInfo(ctx context.Context, reqData *pb.WithUserIdRequest) (*pb.ResponseData, error) {
	authClaims, isLogin := s.Jwt.HanlderCheckLogin(reqData.Common.AuthToken)
	if !isLogin {
		return ResponseError("User is not login", utils.GetFuncName(), fmt.Errorf("User is not login"))
	}
	//if is not superadmin, ignore
	if authClaims.Role != int(utils.RoleSuperAdmin) {
		return ResponseError("There is no permission to access this feature", utils.GetFuncName(), fmt.Errorf("There is no permission to access this feature"))
	}
	userId := reqData.UserId
	if userId < 1 {
		return ResponseError("User id param not found", utils.GetFuncName(), nil)
	}
	//get user by id
	user, err := s.H.GetUserFromId(userId)
	if err != nil {
		return ResponseError("Retrieve user data failed", utils.GetFuncName(), err)
	}
	return ResponseSuccessfullyWithAnyData(authClaims.Username, "Get admin user info successfully", utils.GetFuncName(), user)
}

func (s *Server) GetExcludeLoginUserNameList(ctx context.Context, reqData *pb.CommonRequest) (*pb.ResponseData, error) {
	authClaims, isLogin := s.Jwt.HanlderCheckLogin(reqData.AuthToken)
	if !isLogin {
		return ResponseError("User is not login", utils.GetFuncName(), fmt.Errorf("User is not login"))
	}

	listName := s.H.GetUsernameListExcludeId(authClaims.Id)
	return ResponseSuccessfullyWithAnyData(authClaims.Username, "Get user name list successfully", utils.GetFuncName(), listName)
}

func (s *Server) ChangeUserStatus(ctx context.Context, reqData *pb.ChangeUserStatusRequest) (*pb.ResponseData, error) {
	authClaims, isLogin := s.Jwt.HanlderCheckLogin(reqData.Common.AuthToken)
	if !isLogin {
		return ResponseError("User is not login", utils.GetFuncName(), fmt.Errorf("User is not login"))
	}
	//if is not superadmin, ignore
	if authClaims.Role != int(utils.RoleSuperAdmin) {
		return ResponseError("There is no permission to access this feature", utils.GetFuncName(), fmt.Errorf("There is no permission to access this feature"))
	}

	userIdParam := reqData.UserId
	activeFlg := reqData.Active
	if userIdParam < 1 {
		return ResponseError("Param not found", utils.GetFuncName(), nil)
	}
	user, err := s.H.GetUserFromId(userIdParam)
	if err != nil {
		return ResponseLoginError(authClaims.Username, "Get user from DB error. Please try again!", utils.GetFuncName(), err)
	}

	tx := s.H.DB.Begin()
	user.Status = int(activeFlg)
	user.Updatedt = time.Now().Unix()
	//update user
	updateErr := tx.Save(&user).Error
	if updateErr != nil {
		return ResponseLoginRollbackError(authClaims.Username, tx, "Update User failed. Please try again!", utils.GetFuncName(), updateErr)
	}
	tx.Commit()
	return ResponseSuccessfully(authClaims.Username, "Update User successfully!", utils.GetFuncName())
}

// login check
func (s *Server) IsLoggingOn(ctx context.Context, reqData *pb.CommonRequest) (*pb.ResponseData, error) {
	authClaims, isLogin := s.Jwt.HanlderCheckLogin(reqData.AuthToken)
	if !isLogin {
		return ResponseError("User is not login", utils.GetFuncName(), fmt.Errorf("User is not login"))
	}
	return ResponseSuccessfullyWithAnyData(authClaims.Username, fmt.Sprintf("LoginUser Id: %d", authClaims.Id), utils.GetFuncName(), authClaims)
}

func (s *Server) GenRandomUsername(ctx context.Context, reqData *pb.CommonRequest) (*pb.ResponseData, error) {
	username, err := s.H.GetNewRandomUsername()
	if err != nil {
		return ResponseError("Create new username failed", utils.GetFuncName(), err)
	}
	result := make(map[string]string)
	result["username"] = username
	return ResponseSuccessfullyWithAnyData("", fmt.Sprintf("Get random username: %s", username), utils.GetFuncName(), utils.ObjectToJsonString(result))
}
