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
