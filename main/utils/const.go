package utils

const (
	Tokenkey           = "Token"
	LoginUserKey       = "AuthClaims"
	UserListSessionKey = "userList"
)

var AuthHost, AuthPort, AssetsHost, AssetsPort, ChatHost, ChatPort string

var LoginExcludeUrl = []string{"/404", "/exit", "/login", "/LoginSubmit", "/checkLogin", "/register", "/RegisterSubmit", "/walletSocket", "/withdrawl",
	"/confirmWithdraw", "/passkey/registerStart", "/passkey/registerFinish", "/assertion/options",
	"/assertion/result", "/passkey/cancelRegister", "/assertion/withdrawConfirmLoginResult", "/passkey/withdrawWithNewAccountFinish", "/gen-random-username",
	"/check-user"}

type ResponseData struct {
	IsError bool        `json:"error"`
	Msg     string      `json:"msg"`
	Data    interface{} `json:"data"`
}

type UserRole int
type UrlCodeStatus int

const (
	UrlCodeStatusCreated UrlCodeStatus = iota
	UrlCodeStatusConfirmed
	UrlCodeStatusCancelled
)

const (
	RoleSuperAdmin UserRole = iota
	RoleRegular
)
