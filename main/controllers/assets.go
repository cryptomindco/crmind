package controllers

import (
	"crmind/logpack"
	"crmind/models"
	"crmind/services"
	"crmind/utils"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type AssetsController struct {
	BaseController
}

func (this *AssetsController) FetchRate() {
	var response utils.ResponseData
	req := &services.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: fmt.Sprintf("%s%s", this.AssetsSite(), "/fetch-rate"),
		Payload: map[string]string{},
	}
	if err := services.HttpRequest(req, &response); err != nil {
		this.ResponseError("Fetch rate failed", utils.GetFuncName(), err)
		return
	}
	this.Data["json"] = response
	this.ServeJSON()
}

func (this *AssetsController) GetCodeListData() {
	loginUser, err := this.GetLoginUser()
	if err != nil {
		this.Data["json"] = nil
		this.ServeJSON()
		return
	}
	assetType := strings.TrimSpace(this.GetString("asset"))
	status := strings.TrimSpace(this.GetString("codeStatus"))

	//Get url code list
	urlCodeList, urlCodeErr := this.FilterUrlCodeList(assetType, status)
	if urlCodeErr != nil {
		this.Data["json"] = nil
		this.ServeJSON()
		return
	}
	urlCodeDisplayList := make([]*models.TxCodeDisplay, 0)
	for _, urlCode := range urlCodeList {
		urlCodeStatus, statusErr := utils.GetUrlCodeStatusFromValue(urlCode.Status)
		if statusErr != nil {
			logpack.FError(statusErr.Error(), loginUser.Id, utils.GetFuncName(), nil)
			continue
		}
		txCodeDisp := &models.TxCodeDisplay{
			TxCode:           urlCode,
			CreatedtDisplay:  utils.GetDateTimeDisplay(urlCode.Createdt),
			ConfirmdtDisplay: utils.GetDateTimeDisplay(urlCode.Confirmdt),
			StatusDisplay:    urlCodeStatus.ToString(),
			IsCancelled:      urlCode.Status == int(utils.UrlCodeStatusCancelled),
			IsConfirmed:      urlCode.Status == int(utils.UrlCodeStatusConfirmed),
			IsCreatedt:       urlCode.Status == int(utils.UrlCodeStatusCreated),
		}
		if urlCode.Status == int(utils.UrlCodeStatusConfirmed) && urlCode.HistoryId > 0 {
			history, err := this.GetTxHistory(urlCode.HistoryId)
			if err == nil {
				txCodeDisp.TxHistory = history
			}

		}
		urlCodeDisplayList = append(urlCodeDisplayList, txCodeDisp)
	}

	resultMap := make(map[string]any)
	resultMap["list"] = urlCodeDisplayList
	this.Data["json"] = resultMap
	this.ServeJSON()
}

func (this *AssetsController) GetAddressListData() {
	loginUser, err := this.GetLoginUser()
	if err != nil {
		this.Data["json"] = nil
		this.ServeJSON()
		return
	}
	assetId, assetErr := this.GetInt64("assetId")
	if assetErr != nil {
		this.Data["json"] = nil
		this.ServeJSON()
		return
	}

	//check asset id match with loginUser
	assetMatch := this.CheckAssetMatchUser(assetId)
	if !assetMatch {
		this.Data["json"] = nil
		this.ServeJSON()
		return
	}
	status := strings.TrimSpace(this.GetString("status"))
	//Get url code list
	addressList, addressErr := this.FilterAddressList(assetId, status)
	if addressErr != nil {
		this.Data["json"] = nil
		this.ServeJSON()
		return
	}
	//user token
	token, tokenErr := this.CheckAndCreateAccountToken(loginUser.Id, loginUser.Username, loginUser.Role)
	if tokenErr != nil {
		this.Data["json"] = nil
		this.ServeJSON()
		return
	}
	addressDisplayList := make([]*models.AddressesDisplay, 0)
	for _, address := range addressList {
		//check label of address
		var mainLabel = ""
		if !utils.IsEmpty(address.Label) && strings.Contains(address.Label, fmt.Sprintf("%s_", token)) {
			mainLabel = strings.ReplaceAll(address.Label, fmt.Sprintf("%s_", token), "")
		}
		addressDisp := &models.AddressesDisplay{
			CreatedtDisplay: utils.GetDateTimeDisplay(address.Createdt),
			Addresses:       address,
			TotalReceived:   address.ChainReceived + address.LocalReceived,
			LabelMainPart:   mainLabel,
		}
		addressDisplayList = append(addressDisplayList, addressDisp)
	}

	resultMap := make(map[string]any)
	resultMap["list"] = addressDisplayList
	resultMap["userToken"] = token
	this.Data["json"] = resultMap
	this.ServeJSON()
}

func (this *AssetsController) AssetsDetail() {
	loginUser, err := this.AuthCheck()
	if err != nil {
		this.TplName = "err_403.html"
		return
	}

	this.TplName = "wallet/wallet_detail.html"
	assetType := this.Ctx.Input.Query("type")
	if utils.IsEmpty(assetType) {
		this.TplName = "err_403.html"
		return
	}

	tempRes, assetErr := this.GetAssetByUser(loginUser.Id, assetType)
	if assetErr != nil {
		this.TplName = "err_403.html"
		return
	}
	fmt.Println("Debug 0001")
	var asset *models.Asset
	if !tempRes.Exist {
		asset = utils.CreateNewAsset(assetType, loginUser.Id, loginUser.Username)
	} else {
		asset = tempRes.Asset
	}
	fmt.Println("Debug 0002")
	//Get address list
	addressList := make([]string, 0)
	if asset.Id > 0 {
		var addrErr error
		addressList, addrErr = this.GetAddressListByAssetId(asset.Id)
		if addrErr != nil {
			this.TplName = "err_403.html"
			return
		}
	}
	fmt.Println("Debug 0003")
	assetList, err := this.GetUserAssetList()
	if err != nil {
		logpack.FError("Get Asset List for user failed", loginUser.Id, utils.GetFuncName(), err)
		this.TplName = "err_403.html"
		return
	}
	firstType := ""
	firstBalance := float64(0)
	for _, tmpAsset := range assetList {
		if tmpAsset.Type != assetType {
			firstType = tmpAsset.Type
			firstBalance = tmpAsset.Balance
			break
		}
	}
	fmt.Println("Debug 0004")
	activeAddressCount := this.CountAddressesWithStatus(asset.Id, true)
	archivedAddressCount := this.CountAddressesWithStatus(asset.Id, false)
	//check have code list
	hasCodeList := this.CheckHasCodeList(assetType)
	this.Data["ContactList"] = this.GetContactList()
	this.Data["HasAddress"] = len(addressList) > 0
	this.Data["AddressList"] = addressList
	this.Data["HasCodeList"] = hasCodeList
	this.Data["AssetList"] = assetList
	this.Data["FirstType"] = firstType
	this.Data["FirstBalance"] = firstBalance
	this.Data["ActiveAddressCount"] = activeAddressCount
	this.Data["ArchivedAddressCount"] = archivedAddressCount
	this.Data["TotalAddressesCount"] = archivedAddressCount + activeAddressCount
	this.Data["Asset"] = asset
}

func (this *AssetsController) ConfirmAddressAction() {
	loginUser, err := this.GetLoginUser()
	if err != nil {
		this.ResponseError("An error has occurred. Please try again!", utils.GetFuncName(), nil)
		return
	}
	assetId, parseAssetErr := this.GetInt64("assetId")
	addressId, parseAddrErr := this.GetInt64("addressId")
	if parseAddrErr != nil || parseAssetErr != nil {
		this.ResponseLoginError(loginUser.Id, "Check param error. Please try again!", utils.GetFuncName(), nil)
		return
	}
	action := this.GetString("action")
	var response utils.ResponseData
	formData := url.Values{
		"authorization": {this.GetLoginToken()},
		"assetId":       {fmt.Sprintf("%d", assetId)},
		"addressId":     {fmt.Sprintf("%d", addressId)},
		"action":        {action},
	}
	if err := services.HttpPost(fmt.Sprintf("%s%s", this.AssetsSite(), "/assets/confirm-address-action"), formData, &response); err != nil {
		this.ResponseError("Confirm address action failed", utils.GetFuncName(), err)
		return
	}
	if response.IsError {
		this.ResponseError(response.Msg, utils.GetFuncName(), fmt.Errorf(response.Msg))
		return
	}
	this.ResponseSuccessfully(loginUser.Id, "Confirm address action successfully", utils.GetFuncName())
}

func (this *AssetsController) CancelUrlCode() {
	loginUser, err := this.GetLoginUser()
	if err != nil {
		this.ResponseError("An error has occurred. Please try again!", utils.GetFuncName(), nil)
		return
	}
	//get codeid
	codeIdStr := strings.TrimSpace(this.GetString("codeId"))
	codeId, err := strconv.ParseInt(codeIdStr, 0, 32)
	if err != nil {
		this.ResponseLoginError(loginUser.Id, "Parse codeId to cancel failed. Please try again!", utils.GetFuncName(), nil)
		return
	}

	var response utils.ResponseData
	formData := url.Values{
		"authorization": {this.GetLoginToken()},
		"codeId":        {fmt.Sprintf("%d", codeId)},
	}
	if err := services.HttpPost(fmt.Sprintf("%s%s", this.AssetsSite(), "/assets/cancel-url-code"), formData, &response); err != nil {
		this.ResponseError("cancel url code failed", utils.GetFuncName(), err)
		return
	}
	if response.IsError {
		this.ResponseError(response.Msg, utils.GetFuncName(), fmt.Errorf(response.Msg))
		return
	}
	this.ResponseSuccessfully(loginUser.Id, "Cancel url code successfully", utils.GetFuncName())
}
