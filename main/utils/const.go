package utils

var AuthHost, AuthPort, AssetsHost, AssetsPort string

type ResponseData struct {
	IsError bool        `json:"error"`
	Msg     string      `json:"msg"`
	Data    interface{} `json:"data"`
}
