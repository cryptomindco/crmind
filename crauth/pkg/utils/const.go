package utils

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

type UserRole int
type UserStatus int

const (
	UserListSessionKey = "userList"
	AliveSessionHours  = 24
)

const (
	RoleSuperAdmin UserRole = iota
	RoleRegular
)

const (
	StatusDeactive UserStatus = iota
	StatusActive
)

type ResponseData struct {
	IsError   bool        `json:"error"`
	ErrorCode string      `json:"errorCode"`
	Msg       string      `json:"msg"`
	Data      interface{} `json:"data"`
}
