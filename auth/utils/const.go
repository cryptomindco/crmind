package utils

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

type UserRole int
type UserStatus int

const (
	UserListSessionKey = "userList"
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
	IsError bool        `json:"error"`
	Msg     string      `json:"msg"`
	Data    interface{} `json:"data"`
}
