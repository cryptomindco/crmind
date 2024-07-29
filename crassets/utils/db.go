package utils

import (
	"crassets/models"
	"crassets/walletlib/assets"
	"encoding/json"
	"fmt"

	"github.com/beego/beego/v2/client/orm"
)

func GetSuperadminSystemAddress(assetType string) (string, error) {
	return "", nil
}

func ReadRateFromDB() *models.RateObject {
	//get rate String
	usdResult := make(map[string]float64)
	allResult := make(map[string]float64)
	rateJsonStr, allRate, readErr := ReadRateJsonStrFromDB()
	if readErr != nil || IsEmpty(rateJsonStr) {
		return nil
	}
	//Unamrshal json
	json.Unmarshal([]byte(rateJsonStr), &usdResult)
	json.Unmarshal([]byte(allRate), &allResult)
	return &models.RateObject{
		UsdRates: usdResult,
		AllRates: allResult,
	}
}

// return: usdRate, allRate, error
func ReadRateJsonStrFromDB() (string, string, error) {
	settings := models.Settings{}
	o := orm.NewOrm()
	queryBuilder := fmt.Sprintf("SELECT * from settings")
	settingsErr := o.Raw(queryBuilder).QueryRow(&settings)
	if settingsErr != nil {
		return "", "", settingsErr
	}
	return settings.UsdRate, settings.AllRate, nil
}

func GetUnconfirmedTxHistoryList() ([]models.TxHistory, error) {
	o := orm.NewOrm()
	result := make([]models.TxHistory, 0)
	_, listErr := o.QueryTable(new(models.TxHistory)).Filter("confirmed", false).Exclude("currency", assets.USDWalletAsset.String()).All(&result)
	return result, listErr
}

// Check user exist with username and status active
func GetAssetByOwner(ownerId int64, o orm.Ormer, asseetType string) (*models.Asset, error) {
	asset := models.Asset{}
	queryErr := o.QueryTable(new(models.Asset)).Filter("user_id", ownerId).Filter("type", asseetType).Filter("status", int(AssetStatusActive)).Limit(1).One(&asset)
	return &asset, queryErr
}

func GetAddress(address string) (*models.Addresses, error) {
	addressObj := models.Addresses{}
	o := orm.NewOrm()
	queryErr := o.QueryTable(new(models.Addresses)).Filter("address", address).Limit(1).One(&addressObj)
	if queryErr != nil {
		return nil, queryErr
	}
	return &addressObj, nil
}

func GetAssetById(assetId int64) (*models.Asset, error) {
	asset := models.Asset{}
	o := orm.NewOrm()
	queryErr := o.QueryTable(new(models.Asset)).Filter("id", assetId).Filter("status", int(AssetStatusActive)).Limit(1).One(&asset)
	if queryErr != nil {
		return nil, queryErr
	}
	return &asset, nil
}

func GetUserAsset(userId int64, assetType string) (*models.Asset, error) {
	asset := models.Asset{}
	o := orm.NewOrm()
	queryErr := o.QueryTable(new(models.Asset)).Filter("user_id", userId).Filter("type", assetType).Limit(1).One(&asset)
	if queryErr != nil {
		if queryErr == orm.ErrNoRows {
			return nil, nil
		}
		return nil, queryErr
	}
	return &asset, nil
}

func GetTxHistoryByTxid(txid string) (*models.TxHistory, error) {
	o := orm.NewOrm()
	//Get txHistory by id
	history := models.TxHistory{}
	err := o.QueryTable(new(models.TxHistory)).Filter("txid", txid).Limit(1).One(&history)
	return &history, err
}

func WriteRateToDB(settings *models.Settings, usdRateMap map[string]float64, allRateMap map[string]float64) {
	usdResultString, jsonErr := json.Marshal(usdRateMap)
	allRateString, allJsonErr := json.Marshal(allRateMap)
	if jsonErr != nil || allJsonErr != nil {
		return
	}
	settings.UsdRate = string(usdResultString)
	settings.AllRate = string(allRateString)
	o := orm.NewOrm()
	o.Update(settings)
}

func GetAssetFromAddress(address string, assetType string) (*models.Asset, error) {
	o := orm.NewOrm()
	//Check asset exist on assets table
	assets := models.Asset{}
	queryBuilder := fmt.Sprintf("SELECT * FROM %sasset WHERE type='%s' AND status=%d AND id IN (SELECT asset_id FROM %saddresses WHERE address='%s')", GetAssetRelatedTablePrefix(), assetType, int(AssetStatusActive), GetAssetRelatedTablePrefix(), address)
	err := o.Raw(queryBuilder).QueryRow(&assets)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, err
		}
		return nil, nil
	}

	return &assets, nil
}

func GetSystemUserAsset(assetType string) (*models.Asset, error) {
	systemUser, userErr := GetSystemUser()
	if userErr != nil {
		return nil, userErr
	}
	o := orm.NewOrm()
	return GetAssetByOwner(systemUser, o, assetType)
}

func GetTxcode(code string) (*models.TxCode, bool) {
	if IsEmpty(code) {
		return nil, false
	}
	breakLoop := 0
	var txCode models.TxCode
	var exist bool
	//Try up to 10 times if code fails
	for breakLoop < 10 {
		breakLoop++
		//check code exist on txcode table
		o := orm.NewOrm()
		queryErr := o.QueryTable(new(models.TxCode)).Filter("code", code).Filter("status", int(UrlCodeStatusCreated)).Limit(1).One(&txCode)
		if queryErr != nil {
			continue
		}
		exist = true
		break
	}
	return &txCode, exist
}

func GetRateFromDBByAsset(assetType string) float64 {
	settings := models.Settings{}
	o := orm.NewOrm()
	queryBuilder := fmt.Sprintf("SELECT * from settings")
	settingsErr := o.Raw(queryBuilder).QueryRow(&settings)
	if settingsErr != nil {
		return 0
	}
	//get rate String
	result := make(map[string]float64)
	rateJsonStr := settings.UsdRate
	if IsEmpty(rateJsonStr) {
		return 0
	}
	//Unamrshal json
	json.Unmarshal([]byte(rateJsonStr), &result)
	return result[assetType]
}

func CheckMatchAddressWithUser(assetId, addressId, userId int64, archived bool) bool {
	o := orm.NewOrm()
	queryBuilder := fmt.Sprintf("SELECT count(*) from %sasset as aet where id = %d AND user_id = %d AND EXISTS(SELECT 1 FROM %saddresses WHERE id = %d AND asset_id = aet.id AND archived=%v)", GetAssetRelatedTablePrefix(),
		assetId, userId, GetAssetRelatedTablePrefix(), addressId, archived)
	var count int64
	countErr := o.Raw(queryBuilder).QueryRow(&count)
	return countErr == nil && count > 0
}

func GetAddressById(addressId int64) (*models.Addresses, error) {
	address := models.Addresses{}
	o := orm.NewOrm()
	queryErr := o.QueryTable(new(models.Addresses)).Filter("id", addressId).Limit(1).One(&address)
	if queryErr != nil {
		return nil, queryErr
	}
	return &address, nil
}

func CheckAssetMatchWithUser(assetId, userId int64) bool {
	o := orm.NewOrm()
	count, countErr := o.QueryTable(new(models.Asset)).Filter("user_id", userId).Filter("id", assetId).Count()
	return countErr == nil && count > 0
}

func FilterAddressList(assetId int64, status string) ([]models.Addresses, error) {
	var checkArchived, archived bool
	switch status {
	case "all":
		checkArchived = false
	case "active":
		checkArchived = true
		archived = false
	case "archived":
		checkArchived = true
		archived = true
	default:
		checkArchived = true
		archived = false
	}
	o := orm.NewOrm()
	result := make([]models.Addresses, 0)
	var listErr error
	if checkArchived {
		_, listErr = o.QueryTable(new(models.Addresses)).Filter("asset_id", assetId).Filter("archived", archived).OrderBy("-createdt").All(&result)
	} else {
		_, listErr = o.QueryTable(new(models.Addresses)).Filter("asset_id", assetId).OrderBy("-createdt").All(&result)
	}
	if listErr != nil && listErr != orm.ErrNoRows {
		return nil, listErr
	}
	return result, nil
}

func FilterUrlCodeList(assetType string, status string, userId int64) ([]models.TxCode, error) {
	var statusInt int
	switch status {
	case "unconfirmed":
		statusInt = int(UrlCodeStatusCreated)
	case "confirmed":
		statusInt = int(UrlCodeStatusConfirmed)
	case "cancelled":
		statusInt = int(UrlCodeStatusCancelled)
	default:
		statusInt = -1
	}
	o := orm.NewOrm()
	result := make([]models.TxCode, 0)
	var listErr error
	if statusInt >= 0 {
		_, listErr = o.QueryTable(new(models.TxCode)).Filter("asset", assetType).Filter("ownerId", userId).Filter("status", statusInt).All(&result)
	} else {
		_, listErr = o.QueryTable(new(models.TxCode)).Filter("asset", assetType).Filter("ownerId", userId).All(&result)
	}
	if listErr != nil && listErr != orm.ErrNoRows {
		return nil, listErr
	}
	return result, nil
}

func GetTxHistoryById(txHistoryId int64) (*models.TxHistory, error) {
	o := orm.NewOrm()
	//Get txHistory by id
	history := models.TxHistory{}
	err := o.QueryTable(new(models.TxHistory)).Filter("id", txHistoryId).Limit(1).One(&history)
	return &history, err
}

func CancelTxCodeById(ownerId int64, codeId int64) error {
	o := orm.NewOrm()
	txCode := models.TxCode{}
	queryErr := o.QueryTable(new(models.TxCode)).Filter("id", codeId).Limit(1).One(&txCode)
	if queryErr != nil {
		return queryErr
	}
	//if ownerId not match
	if ownerId != txCode.OwnerId {
		return fmt.Errorf("%s", "Owner not match. There is no right to cancel this Code")
	}
	txCode.Status = int(UrlCodeStatusCancelled)
	//update txCode
	_, updateErr := o.Update(&txCode)
	return updateErr
}

// Create new code for withdraw url, 32 characters
func CreateNewUrlCode() (string, bool) {
	breakLoop := 0
	//Try up to 10 times if token creation fails
	for breakLoop <= 10 {
		newCode := RandSeq(32)
		breakLoop++
		//check token exist on user table
		o := orm.NewOrm()
		codeCount, queryErr := o.QueryTable(new(models.TxCode)).Filter("code", newCode).Count()
		if queryErr != nil {
			continue
		}
		if codeCount == 0 {
			return newCode, true
		}
	}
	return "", false
}
