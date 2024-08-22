package db

import (
	"crauth/pkg/config"
	"crauth/pkg/models"
	"crauth/pkg/utils"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

func Init(config config.Config) Handler {
	db, err := gorm.Open(postgres.Open(config.DBUrl), &gorm.Config{})

	if err != nil {
		log.Fatalln(err)
	}
	db.AutoMigrate(
		&models.User{},
	)
	return Handler{DB: db}
}

// Check user exist with username and status active
func (h *Handler) CheckUserExist(username string) (bool, error) {
	var count int64
	if result := h.DB.Model(&models.User{}).Where("username = ? AND status = ?", username, int(utils.StatusActive)).Count(&count); result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// Check user exist with username and status active
func (h *Handler) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	result := h.DB.Where(&models.User{Username: username, Status: int(utils.StatusActive)}).First(&user)
	return &user, result.Error
}

func (h *Handler) GetHasCredUserList() ([]models.User, error) {
	queryBuilder := "SELECT * FROM public.user WHERE creds_arr_json <> ''"
	var res []models.User
	err := h.DB.Raw(queryBuilder).Scan(&res).Error
	return res, err
}

func (h *Handler) GetUserFromId(userId int64) (*models.User, error) {
	user := models.User{}
	result := h.DB.Where(&models.User{Id: userId, Status: int(utils.StatusActive)}).First(&user)
	return &user, result.Error
}

func (h *Handler) GetNewRandomUsername() (string, error) {
	breakLoop := 0
	//Try up to 10 times if username creations failed
	for breakLoop < 10 {
		newUsername := utils.RandSeq(8)
		fmt.Println("Check hreere", newUsername)
		breakLoop++
		//check token exist on user table
		exist, err := h.CheckUserExist(newUsername)
		fmt.Println("Check err : ", err)
		if err != nil || exist {
			continue
		}
		return newUsername, nil
	}
	return "", fmt.Errorf("%s", "Create new username failed")
}

func (h *Handler) GetSystemUser() (*models.User, error) {
	user := models.User{}
	result := h.DB.Where(&models.User{Role: int(utils.RoleSuperAdmin), Status: int(utils.StatusActive)}).First(&user)
	return &user, result.Error
}

func (h *Handler) GetUserListWithExcludeId(excludeId int64) ([]models.User, error) {
	userList := make([]models.User, 0)
	listErr := h.DB.Where("id <> ?", excludeId).Order("createdt").Find(&userList).Error
	return userList, listErr
}

func (h *Handler) GetUsernameListExcludeId(loginUserId int64) []*models.UserInfo {
	result := make([]*models.UserInfo, 0)
	userList, err := h.GetUserListWithExcludeId(loginUserId)
	if err != nil {
		return result
	}
	for _, user := range userList {
		result = append(result, &models.UserInfo{
			Id:       user.Id,
			Username: user.Username,
			Role:     user.Role,
		})
	}
	return result
}
