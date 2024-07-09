package passkey

import (
	"auth/models"
	"auth/utils"

	"github.com/beego/beego/v2/client/orm"
)

func InitSessionUser(loginUser models.User) models.SessionUser {
	//get user list that registed on passkey manager
	passkeyAllUsers := Datastore.GetAllUsername()
	validUsers := make([]string, 0)
	o := orm.NewOrm()
	for _, username := range passkeyAllUsers {
		if username == loginUser.Username {
			continue
		}
		//check exist user
		exist, err := utils.CheckUserExist(username, o)
		if err != nil || !exist {
			continue
		}
		validUsers = append(validUsers, username)
	}
	result := models.SessionUser{
		User:       &loginUser,
		OtherUsers: validUsers,
	}
	return result
}
