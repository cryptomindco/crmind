package controllers

import (
	"crassets/dohttp"
	"crassets/handler"
	"crassets/logpack"
	"crassets/models"
	"crassets/utils"
	"crassets/walletlib/assets"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	beego "github.com/beego/beego/v2/adapter"
	"github.com/beego/beego/v2/client/orm"
)

var (
	txHistoryModel = new(models.TxHistory)
	assetsModel    = new(models.Asset)
	addressesModel = new(models.Addresses)
	ratesModel     = new(models.Rates)
	accountsModel  = new(models.Accounts)
	txCodeModel    = new(models.TxCode)
)

var (
	leftTreeResultMap = make(map[int][]orm.Params)
)

type BaseController struct {
	beego.Controller
}

func (this *BaseController) ResponseSuccessfully(loginId int64, msg string, funcName string) {
	if loginId <= 0 {
		logpack.Info(msg, funcName)
	} else {
		logpack.FInfo(msg, loginId, funcName)
	}
	this.Data["json"] = utils.ResponseData{
		IsError: false,
		Msg:     msg,
	}
	this.ServeJSON()
}

func (this *BaseController) ResponseLoginRollbackError(loginId int64, tx orm.TxOrmer, msg string, funcName string, err error) {
	tx.Rollback()
	this.ResponseLoginError(loginId, msg, funcName, err)
}

func (this *BaseController) ResponseLoginError(loginId int64, msg string, funcName string, err error) {
	if loginId <= 0 {
		logpack.Error(msg, funcName, err)
	} else {
		logpack.FError(msg, loginId, funcName, err)
	}
	this.Data["json"] = utils.ResponseData{
		IsError: true,
		Msg:     msg,
	}
	this.ServeJSON()
}

func (this *BaseController) ResponseLoginErrorWithCode(loginId int64, errCode string, msg string, funcName string, err error) {
	if loginId <= 0 {
		logpack.Error(msg, funcName, err)
	} else {
		logpack.FError(msg, loginId, funcName, err)
	}
	this.Data["json"] = utils.ResponseData{
		IsError:   true,
		ErrorCode: errCode,
		Msg:       msg,
	}
	this.ServeJSON()
}

func (this *BaseController) ResponseError(msg string, funcName string, err error) {
	logpack.Error(msg, funcName, err)
	this.Data["json"] = utils.ResponseData{
		IsError: true,
		Msg:     msg,
	}
	this.ServeJSON()
}

func (this *BaseController) ResponseRollbackError(tx orm.TxOrmer, msg string, funcName string, err error) {
	tx.Rollback()
	this.ResponseError(msg, funcName, err)
}

func (this *BaseController) ResponseSuccessfullyWithAnyData(loginId int64, msg, funcName string, result any) {
	if loginId <= 0 {
		logpack.Info(msg, funcName)
	} else {
		logpack.FInfo(msg, loginId, funcName)
	}
	this.Data["json"] = utils.ResponseData{
		IsError: false,
		Data:    result,
	}
	this.ServeJSON()
}

func (this *BaseController) ResponseSuccessfullyWithAnyDataNoLog(result any) {
	this.Data["json"] = utils.ResponseData{
		IsError: false,
		Data:    result,
	}
	this.ServeJSON()
}

func (this *BaseController) GetUserInfoByName(username string) (*models.UserInfo, error) {
	checkUrl := fmt.Sprintf("%s%s", utils.AuthSite(), "/user-by-name")
	req := &dohttp.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: checkUrl,
		Payload: map[string]string{
			"username": username,
		},
	}
	var response utils.ResponseData
	err := dohttp.HttpRequest(req, &response)
	if err != nil {
		return nil, err
	}
	if response.IsError {
		return nil, fmt.Errorf("Get user info by username failed")
	}
	var userInfo models.UserInfo
	err = utils.CatchObject(response.Data, &userInfo)
	if err != nil {
		return nil, err
	}
	return &userInfo, nil
}

func (this *BaseController) AuthTokenCheck(token string) (*models.AuthClaims, error) {
	var response utils.ResponseData
	req := &dohttp.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: fmt.Sprintf("%s%s", utils.AuthSite(), "/is-logging"),
		Payload: map[string]string{},
		Header: map[string]string{
			"Authorization": token,
		},
	}

	err := dohttp.HttpRequest(req, &response)
	if err != nil {
		return nil, err
	}

	if response.IsError {
		return nil, fmt.Errorf(response.Msg)
	}

	bytes, err := json.Marshal(response.Data)
	if err != nil {
		return nil, err
	}
	var authRes models.AuthClaims
	err = json.Unmarshal(bytes, &authRes)
	if err != nil {
		return nil, err
	}
	return &authRes, nil
}

// func (this *BaseController) CheckExistChat(fromId, toId int64) (bool, error) {
// 	authToken := this.GetString("authorization")
// 	checkUrl := fmt.Sprintf("%s%s", utils.AuthSite(), "/chat-exist")
// 	req := &dohttp.ReqConfig{
// 		Method:  http.MethodGet,
// 		HttpUrl: checkUrl,
// 		Payload: map[string]string{
// 			"fromId": fmt.Sprintf("%d", fromId),
// 			"toId":   fmt.Sprintf("%d", toId),
// 		},
// 		Header: map[string]string{
// 			"Authorization": authToken,
// 		},
// 	}
// 	var response utils.ResponseData
// 	err := dohttp.HttpRequest(req, &response)
// 	if err != nil {
// 		return false, err
// 	}
// 	if response.IsError {
// 		return false, fmt.Errorf(response.Msg)
// 	}
// 	exist := response.Data.(bool)
// 	return exist, nil
// }

func (this *BaseController) AuthCheck() (*models.AuthClaims, error) {
	authen := this.Ctx.Request.Header.Get("Authorization")
	return this.AuthTokenCheck(authen)
}

func (this *BaseController) CreateNewAddressForAsset(userId int64, username string, isAdmin bool, assetObject assets.AssetType) (*models.Addresses, *models.Asset, error) {
	o := orm.NewOrm()
	tx, beginErr := o.Begin()
	if beginErr != nil {
		return nil, nil, fmt.Errorf("An error has occurred. Please try again!")
	}
	handler.UpdateAssetManagerByType(assetObject.String())
	assetObj, assetMgrExist := utils.GlobalItem.AssetMgrMap[assetObject.String()]
	if !assetMgrExist {
		return nil, nil, fmt.Errorf("RPC Client failed at the server. Please contact admin!")
	}
	assetObj.MutexLock()
	defer assetObj.MutexUnlock()
	var assetLabel string
	if assetObject == assets.DCRWalletAsset {
		assetLabel = username
	} else {
		//Check and get user token
		var role int
		if isAdmin {
			role = int(utils.RoleSuperAdmin)
		} else {
			role = int(utils.RoleRegular)
		}
		token, _, err := utils.CheckAndCreateAccountToken(userId, username, role)
		if err != nil {
			return nil, nil, fmt.Errorf("Check or create user token failed")
		}
		//default label format: token_%label%
		assetLabel = fmt.Sprintf("%s%s", token, utils.CreateDefaultAddressLabelPostfix(assetObject.String()))
	}
	//Create new address with label. Label form is: btc_address_$username
	newAddress, addrErr := assetObj.CreateNewAddressWithLabel(username, assetLabel)
	if addrErr != nil {
		return nil, nil, fmt.Errorf("Creating an address with Label failed. Please try again!")
	}

	//Get asset from DB
	userAsset, err := utils.GetUserAsset(userId, assetObject.String())
	if err != nil {
		return nil, nil, fmt.Errorf("Get Asset from DB failed. Please try again!")
	}
	//if user asset is nil, insert new asset to DB
	var assetId int64
	if userAsset == nil {
		asset := models.Asset{
			DisplayName: assetObject.ToFullName(),
			UserId:      userId,
			UserName:    username,
			Type:        assetObject.String(),
			Sort:        assetObject.AssetSortInt(),
			Status:      int(utils.AssetStatusActive),
			IsAdmin:     isAdmin,
			Createdt:    time.Now().Unix(),
			Updatedt:    time.Now().Unix(),
		}
		//update user
		id, insertErr := tx.Insert(&asset)
		if insertErr != nil {
			tx.Rollback()
			return nil, nil, fmt.Errorf("Insert new asset failed. Please try again!")
		}
		userAsset = &asset
		assetId = id
	} else {
		assetId = userAsset.Id
	}

	//insert new Address
	insertAddress := models.Addresses{
		AssetId:  assetId,
		Address:  newAddress,
		Label:    assetLabel,
		Createdt: time.Now().Unix(),
	}
	_, insertAddressErr := tx.Insert(&insertAddress)
	if insertAddressErr != nil {
		tx.Rollback()
		return nil, nil, fmt.Errorf("Insert new address label failed. Please try again!")
	}
	//Insert to asset table
	tx.Commit()
	return &insertAddress, userAsset, nil
}

func (this *BaseController) IsCryptoCurrency(assetType string) bool {
	var lowercaseType = strings.ToLower(assetType)
	return lowercaseType == assets.BTCWalletAsset.String() || lowercaseType == assets.DCRWalletAsset.String() || lowercaseType == assets.LTCWalletAsset.String()
}

func (this *BaseController) GetAddressListByAssetId(assetId int64) ([]string, error) {
	addressList := make([]*models.Addresses, 0)
	o := orm.NewOrm()
	_, queryErr := o.QueryTable(addressesModel).Filter("asset_id", assetId).Filter("archived", false).OrderBy("-createdt").All(&addressList)
	if queryErr != nil {
		return make([]string, 0), queryErr
	}
	result := make([]string, 0)
	for _, address := range addressList {
		result = append(result, address.Address)
	}
	return result, nil
}

func (this *BaseController) HandlerInternalWithdrawl(txCode *models.TxCode, user models.AuthClaims, rateSend float64, o orm.Ormer) bool {
	tx, beginErr := o.Begin()
	if beginErr != nil {
		this.ResponseError("An error has occurred. Please try again!", utils.GetFuncName(), beginErr)
		return false
	}
	//get assets of sender
	assetObj := assets.StringToAssetType(txCode.Asset)
	senderAsset, senderAssetErr := utils.GetUserAsset(txCode.OwnerId, txCode.Asset)
	if senderAssetErr != nil || senderAsset == nil {
		this.ResponseError("Error getting Asset data from DB or sender asset does not exist. Please try again!", utils.GetFuncName(), nil)
		return false
	}
	//if balance is not enough to withdraw
	if senderAsset.Balance < txCode.Amount {
		this.ResponseError("Balance is not enough to withdraw", utils.GetFuncName(), nil)
		return false
	}

	//Deduct money from balance and update local transfer total
	senderAsset.Balance -= txCode.Amount
	senderAsset.LocalSent += txCode.Amount

	//get assets of receiver
	receiverAsset, receiverAssetErr := utils.GetUserAsset(user.Id, txCode.Asset)
	if receiverAssetErr != nil {
		this.ResponseError("Retrieve recipient asset data failed. Please try again!", utils.GetFuncName(), receiverAssetErr)
		return false
	}

	//update sender asset
	_, senderAssetUpdateErr := tx.Update(senderAsset)
	if senderAssetUpdateErr != nil {
		this.ResponseRollbackError(tx, "Update Sender failed. Please try again!", utils.GetFuncName(), senderAssetUpdateErr)
		return false
	}
	receiverAssetCreate := receiverAsset == nil

	//if receiver create asset
	if receiverAssetCreate {
		_, newReceiverAsset, newErr := this.CreateNewAddressForAsset(user.Id, user.Username, utils.IsSuperAdmin(user.Role), assetObj)
		if newErr != nil {
			this.ResponseError("Create new asset and address failed. Please check again!", utils.GetFuncName(), newErr)
			return false
		}
		newReceiverAsset.Balance = txCode.Amount
		newReceiverAsset.LocalReceived = txCode.Amount
		newReceiverAsset.Updatedt = time.Now().Unix()

		_, receiverAssetUpdateErr := tx.Update(newReceiverAsset)
		if receiverAssetUpdateErr != nil {
			this.ResponseRollbackError(tx, "Update balance for asset failed. Please check again!", utils.GetFuncName(), receiverAssetUpdateErr)
			return false
		}
	} else {
		//update receiver asset
		receiverAsset.Balance += txCode.Amount
		receiverAsset.LocalReceived += txCode.Amount
		_, receiverAssetUpdateErr := tx.Update(receiverAsset)
		if receiverAssetUpdateErr != nil {
			this.ResponseRollbackError(tx, "Update recipient assets failed. Please try again!", utils.GetFuncName(), receiverAssetUpdateErr)
			return false
		}
	}

	//insert to transaction history
	txHistory := models.TxHistory{}
	txHistory.SenderId = txCode.OwnerId
	txHistory.Sender = txCode.OwnerName
	txHistory.ReceiverId = user.Id
	txHistory.Receiver = user.Username
	txHistory.Currency = txCode.Asset
	txHistory.Amount = txCode.Amount
	txHistory.Status = 1
	txHistory.Description = txCode.Note
	txHistory.Createdt = time.Now().UnixNano() / 1e9
	txHistory.TransType = int(utils.TransTypeLocal)
	txHistory.Rate = rateSend

	_, HistoryErr := tx.Insert(&txHistory)
	if HistoryErr != nil {
		this.ResponseRollbackError(tx, "Recorded history is corrupted. Please check your balance again!", utils.GetFuncName(), HistoryErr)
		return false
	}

	//update txCode status
	txCode.Status = int(utils.UrlCodeStatusConfirmed)
	txCode.HistoryId = txHistory.Id
	txCode.Confirmdt = time.Now().Unix()
	_, txUpdateErr := tx.Update(txCode)
	if txUpdateErr != nil {
		this.ResponseRollbackError(tx, "Update Tx Code status failed!", utils.GetFuncName(), txUpdateErr)
		return false
	}
	tx.Commit()
	this.Data["json"] = map[string]string{"error": ""}
	this.ServeJSON()
	return true
}

func (this *BaseController) GetAssetList(userId int64, allowAsset []string) ([]*models.Asset, error) {
	var assetList []*models.Asset
	o := orm.NewOrm()
	tempAllowAssets := make([]string, 0)
	for _, asset := range allowAsset {
		tempAllowAssets = append(tempAllowAssets, fmt.Sprintf("'%s'", asset))
	}
	builderSQL := fmt.Sprintf("SELECT * from %sasset WHERE user_id=%d AND status = %d AND type IN (%s) ORDER BY sort", utils.GetAssetRelatedTablePrefix(), userId, int(utils.AssetStatusActive), strings.Join(tempAllowAssets, ","))
	_, err := o.Raw(builderSQL).QueryRows(&assetList)
	if err != nil {
		return nil, err
	}
	return assetList, nil
}

func (this *BaseController) SyncAssetList(user *models.AuthClaims, summaryList []*models.Asset, allowAssetList []string) []*models.Asset {
	result := make([]*models.Asset, 0)
	for _, allowAsset := range allowAssetList {
		exist := false
		for _, summary := range summaryList {
			if summary.Type == allowAsset {
				result = append(result, summary)
				exist = true
				break
			}
		}
		if !exist {
			summary := this.CreateNewAsset(allowAsset, user.Id, user.Username)
			result = append(result, summary)
		}
	}
	return result
}

func (this *BaseController) CreateNewAsset(assetType string, userId int64, username string) *models.Asset {
	assetObject := assets.StringToAssetType(assetType)
	return &models.Asset{
		Sort:          assetObject.AssetSortInt(),
		DisplayName:   assetObject.ToFullName(),
		UserId:        userId,
		UserName:      username,
		Type:          assetType,
		Balance:       0,
		LocalReceived: 0,
		LocalSent:     0,
		ChainReceived: 0,
		ChainSent:     0,
		TotalFee:      0,
	}
}
