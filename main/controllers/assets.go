package controllers

import (
	"crmind/logpack"
	"crmind/models"
	"crmind/pb/assetspb"
	"crmind/services"
	"crmind/utils"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type AssetsController struct {
	BaseController
}

func (this *AssetsController) FetchRate() {
	res, err := services.FetchRateHandler(this.Ctx.Request.Context())
	if err != nil {
		this.ResponseError("Fetch rate failed", utils.GetFuncName(), err)
		return
	}
	this.Data["json"] = res
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
	urlCodeList, urlCodeErr := this.FilterUrlCodeList(loginUser.Username, assetType, status)
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
			history, err := this.GetTxHistory(loginUser.Username, urlCode.HistoryId)
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
	assetMatch := this.CheckAssetMatchUser(loginUser.Username, assetId)
	if !assetMatch {
		this.Data["json"] = nil
		this.ServeJSON()
		return
	}
	status := strings.TrimSpace(this.GetString("status"))
	//Get url code list
	addressList, addressErr := this.FilterAddressList(loginUser.Username, assetId, status)
	if addressErr != nil {
		this.Data["json"] = nil
		this.ServeJSON()
		return
	}
	//user token
	token, tokenErr := this.CheckAndCreateAccountToken(loginUser.Username, loginUser.Username, loginUser.Role)
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

	tempRes, assetErr := this.GetAssetByUser(loginUser.Username, loginUser.Username, assetType)
	if assetErr != nil {
		this.TplName = "err_403.html"
		return
	}
	//check exist
	var asset *models.Asset
	if !tempRes.Exist {
		asset = utils.CreateNewAsset(assetType, loginUser.Username)
	} else {
		asset = tempRes.Asset
	}
	//Get address list
	addressList := make([]string, 0)
	if asset.Id > 0 {
		var addrErr error
		addressList, addrErr = this.GetAddressListByAssetId(loginUser.Username, asset.Id)
		if addrErr != nil {
			this.TplName = "err_403.html"
			return
		}
	}
	assetList, err := this.GetUserAssetList(loginUser.Username, loginUser.Username)
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
	activeAddressCount := int64(0)
	archivedAddressCount := int64(0)
	if asset.Id > 0 {
		activeAddressCount = this.CountAddressesWithStatus(loginUser.Username, asset.Id, true)
		archivedAddressCount = this.CountAddressesWithStatus(loginUser.Username, asset.Id, false)
	}
	priceSpread, _ := utils.GetPriceSpread()
	//check have code list
	hasCodeList := this.CheckHasCodeList(loginUser.Username, assetType)
	this.Data["ContactList"] = this.GetContactList(loginUser.Username)
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
	this.Data["PriceSpread"] = priceSpread
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
	//confirm address action
	_, err = services.ConfirmAddressActionHandler(this.Ctx.Request.Context(), &assetspb.ConfirmAddressActionRequest{
		Common: &assetspb.CommonRequest{
			LoginName: loginUser.Username,
		},
		AssetId:   assetId,
		AddressId: addressId,
		Action:    action,
	})

	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
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

	_, err = services.CancelUrlCodeHandler(this.Ctx.Request.Context(), &assetspb.OneIntegerRequest{
		Common: &assetspb.CommonRequest{
			LoginName: loginUser.Username,
		},
		Data: codeId,
	})
	if err != nil {
		this.ResponseError(err.Error(), utils.GetFuncName(), err)
		return
	}
	this.ResponseSuccessfully(loginUser.Id, "Cancel url code successfully", utils.GetFuncName())
}

func (this *AssetsController) TransactionDetail() {
	loginUser, check := this.AuthCheck()
	if check != nil {
		this.TplName = "err_403.html"
		return
	}
	var historyId int64
	if err := this.Ctx.Input.Bind(&historyId, "id"); err != nil {
		this.Redirect("/", http.StatusFound)
	}

	res, err := services.TransactionDetailHandler(this.Ctx.Request.Context(), &assetspb.OneIntegerRequest{
		Common: &assetspb.CommonRequest{
			LoginName: loginUser.Username,
		},
		Data: historyId,
	})
	if err != nil {
		logpack.Error("Get Transaction history failed", utils.GetFuncName(), err)
		this.TplName = "err_403.html"
		return
	}
	var txHistoryDisp models.TxHistoryDisplay
	parseErr := utils.JsonStringToObject(res.Data, &txHistoryDisp)
	if parseErr != nil {
		logpack.Error("Parse tx history data failed", utils.GetFuncName(), err)
		this.TplName = "err_403.html"
		return
	}
	this.Data["TransactionHistory"] = txHistoryDisp
	this.TplName = "transactions/detail.html"
}
