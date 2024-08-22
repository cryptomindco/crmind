package services

import (
	"crauth/pkg/logpack"
	"crauth/pkg/pb"
	"crauth/pkg/utils"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

func ResponseLoginRollbackError(loginName string, tx *gorm.DB, msg string, funcName string, err error) (*pb.ResponseData, error) {
	tx.Rollback()
	return ResponseLoginError(loginName, msg, funcName, err)
}

func ResponseLoginError(loginName string, msg string, funcName string, err error) (*pb.ResponseData, error) {
	if utils.IsEmpty(loginName) {
		logpack.Error(msg, funcName, err)
	} else {
		logpack.FError(msg, loginName, funcName, err)
	}
	return &pb.ResponseData{
		Error: true,
		Msg:   msg,
	}, fmt.Errorf(msg)
}

func ResponseLoginErrorWithCode(loginName string, errCode string, msg string, funcName string, err error) (*pb.ResponseData, error) {
	if utils.IsEmpty(loginName) {
		logpack.Error(msg, funcName, err)
	} else {
		logpack.FError(msg, loginName, funcName, err)
	}
	return &pb.ResponseData{
		Error: true,
		Code:  errCode,
		Msg:   msg,
	}, fmt.Errorf(msg)
}

func ResponseError(msg string, funcName string, err error) (*pb.ResponseData, error) {
	logpack.Error(msg, funcName, err)
	return &pb.ResponseData{
		Error: true,
		Msg:   msg,
	}, fmt.Errorf(msg)
}

func ResponseSuccessfully(loginName string, msg string, funcName string) (*pb.ResponseData, error) {
	if utils.IsEmpty(loginName) {
		logpack.Info(msg, funcName)
	} else {
		logpack.FInfo(msg, loginName, funcName)
	}
	return &pb.ResponseData{
		Error: false,
		Msg:   msg,
	}, nil
}

func ResponseSuccessfullyWithAnyData(loginName string, msg, funcName string, result any) (*pb.ResponseData, error) {
	if utils.IsEmpty(loginName) {
		logpack.Info(msg, funcName)
	} else {
		logpack.FInfo(msg, loginName, funcName)
	}
	//result to json
	b, err := json.Marshal(result)
	data := "{}"
	if err == nil {
		data = string(b)
	}
	return &pb.ResponseData{
		Error: false,
		Data:  data,
	}, nil
}

func ResponseSuccessfullyWithAnyDataNoLog(result any) (*pb.ResponseData, error) {
	//result to json
	b, err := json.Marshal(result)
	data := "{}"
	if err == nil {
		data = string(b)
	}
	return &pb.ResponseData{
		Error: false,
		Data:  data,
	}, nil
}
