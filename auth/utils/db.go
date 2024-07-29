package utils

import (
	"auth/models"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

// Check user exist with username and status active
func CheckUserExist(username string, o orm.Ormer) (bool, error) {
	count, err := o.QueryTable(new(models.User)).Filter("username", username).Filter("status", int(StatusActive)).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Check user exist with username and status active
func GetUserByUsername(username string, o orm.Ormer) (*models.User, error) {
	user := models.User{}
	queryErr := o.QueryTable(new(models.User)).Filter("username", username).Filter("status", int(StatusActive)).Limit(1).One(&user)
	return &user, queryErr
}

func GetUserFromToken(token string) (*models.User, error) {
	o := orm.NewOrm()
	user := models.User{}
	queryErr := o.QueryTable(new(models.User)).Filter("token", token).Filter("status", int(StatusActive)).Limit(1).One(&user)
	return &user, queryErr
}

func GetTokenFromUserId(userId int64) string {
	user, err := GetUserFromId(userId)
	if err != nil {
		return ""
	}
	return user.Token
}

func GetHasCredUserList() ([]models.User, error) {
	o := orm.NewOrm()
	queryBuilder := "SELECT * FROM public.user WHERE creds_arr_json <> ''"
	var res []models.User
	_, err := o.Raw(queryBuilder).QueryRows(&res)
	return res, err
}

func GetUserFromId(userId int64) (*models.User, error) {
	user := models.User{}
	o := orm.NewOrm()
	queryErr := o.QueryTable(new(models.User)).Filter("id", userId).Filter("status", int(StatusActive)).Limit(1).One(&user)
	if queryErr != nil {
		return nil, queryErr
	}
	return &user, nil
}

func CheckValidToken(token string) bool {
	breakLoop := 0
	//Try up to 10 times if token fails
	for breakLoop < 10 {
		breakLoop++
		//check token exist on user table
		o := orm.NewOrm()
		userCount, queryErr := o.QueryTable(new(models.User)).Filter("token", token).Count()
		if queryErr != nil {
			continue
		}
		return userCount > 0
	}
	return false
}

func GetUserOfToken(token string) *models.User {
	breakLoop := 0
	//Try up to 10 times if token creation fails
	for breakLoop < 10 {
		breakLoop++
		//check token exist on user table
		o := orm.NewOrm()
		var user models.User
		queryErr := o.QueryTable(new(models.User)).Filter("token", token).Filter("status", int(StatusActive)).Limit(1).One(&user)
		if queryErr != nil && queryErr != orm.ErrNoRows {
			continue
		}
		if queryErr == orm.ErrNoRows {
			return nil
		}
		return &user
	}
	return nil
}

// Check and create new token for user, if exist, ignore
func CheckAndCreateUserToken(user models.User) (token string, updated bool, err error) {
	if !IsEmpty(user.Token) {
		token = user.Token
		updated = false
	}
	//get user
	currentUser, userErr := GetUserFromId(user.Id)
	if userErr != nil {
		err = userErr
		return
	}
	if !IsEmpty(currentUser.Token) {
		token = currentUser.Token
		updated = false
		return
	}
	//Create new token
	newToken, ok := CreateNewUserToken()
	if !ok {
		err = fmt.Errorf("%s", "Create new token failed")
		return
	}
	currentUser.Token = newToken
	//update new Token
	o := orm.NewOrm()
	tx, beginErr := o.Begin()
	if beginErr != nil {
		err = beginErr
		return
	}
	currentUser.Updatedt = time.Now().Unix()
	_, updateErr := tx.Update(currentUser)
	if updateErr != nil {
		tx.Rollback()
		err = updateErr
		return
	}
	token = newToken
	updated = true
	tx.Commit()
	return
}

func GetNewRandomUsername() (string, error) {
	breakLoop := 0
	//Try up to 10 times if username creations failed
	for breakLoop < 10 {
		newUsername := RandSeq(8)
		breakLoop++
		//check token exist on user table
		o := orm.NewOrm()
		exist, err := CheckUserExist(newUsername, o)
		if err != nil || exist {
			continue
		}
		return newUsername, nil
	}
	return "", fmt.Errorf("%s", "Create new username failed")
}

// Create new user token, 6 characters
func CreateNewUserToken() (string, bool) {
	breakLoop := 0
	//Try up to 10 times if token creation fails
	for breakLoop < 10 {
		newToken := RandSeq(6)
		breakLoop++
		//check token exist on user table
		o := orm.NewOrm()
		userCount, queryErr := o.QueryTable(new(models.User)).Filter("token", newToken).Count()
		if queryErr != nil {
			continue
		}
		if userCount == 0 {
			return newToken, true
		}
	}
	return "", false
}

func GetContactListOfUser(userId int64) []string {
	result := make([]string, 0)
	user, userErr := GetUserFromId(userId)
	if userErr != nil {
		return result
	}

	if IsEmpty(user.Contacts) {
		return result
	}

	var contacts []models.ContactItem
	err := json.Unmarshal([]byte(user.Contacts), &contacts)
	if err != nil {
		return result
	}
	for _, contact := range contacts {
		result = append(result, contact.UserName)
	}
	return result
}

func GetSystemUser() (*models.User, error) {
	o := orm.NewOrm()
	user := models.User{}
	queryErr := o.QueryTable(new(models.User)).Filter("role", int(RoleSuperAdmin)).Filter("status", int(StatusActive)).Limit(1).One(&user)
	return &user, queryErr
}

// Check and get user from address label
func GetUserFromLabel(label string) (*models.User, bool) {
	if IsEmpty(label) {
		return nil, false
	}

	//split label
	labelArr := strings.Split(label, "_")
	if len(labelArr) <= 1 {
		return nil, false
	}

	//token
	token := labelArr[0]
	//check token from user
	user, userErr := GetUserFromToken(token)
	if userErr != nil {
		return nil, false
	}
	return user, true
}

func GetContactListFromUser(userId int64) ([]models.ContactItem, error) {
	user, userErr := GetUserFromId(userId)
	if userErr != nil {
		return nil, userErr
	}
	result := make([]models.ContactItem, 0)
	if IsEmpty(user.Contacts) {
		return result, nil
	}
	err := json.Unmarshal([]byte(user.Contacts), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
