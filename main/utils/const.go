package utils

const Tokenkey = "Token"

var AuthHost, AuthPort, AssetsHost, AssetsPort string

var LoginExcludeUrl = []string{"/404", "/exit", "/login", "/LoginSubmit", "/checkLogin", "/register", "/RegisterSubmit", "/walletSocket", "/withdrawl",
	"/confirmWithdraw", "/passkey/registerStart", "/passkey/registerFinish", "/assertion/options",
	"/assertion/result", "/passkey/cancelRegister", "/assertion/withdrawConfirmLoginResult", "/passkey/withdrawWithNewAccountFinish", "/gen-random-username",
	"/check-user"}

type ResponseData struct {
	IsError bool        `json:"error"`
	Msg     string      `json:"msg"`
	Data    interface{} `json:"data"`
}
