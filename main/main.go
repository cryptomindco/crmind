package main

import (
	_ "crmind/routers"
	"crmind/services"
	"crmind/utils"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	beego "github.com/beego/beego/v2/adapter"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	addFuncMap()
	fileName := "logs/crmind_op.log"
	var exist bool
	for !exist {
		err := os.MkdirAll("logs", os.ModePerm)
		if err == nil {
			exist = true
		}
	}
	// open log file
	logFile, logErr := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE, 0644)
	if logErr != nil {
		log.Panic(logErr)
	}
	defer logFile.Close()
	// redirect all the output to file
	wrt := io.MultiWriter(os.Stdout, logFile)

	// set log out put
	log.SetOutput(wrt)
	// optional: log date-time, filename, and line number
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	beego.BConfig.AppName = "crmind"
	beego.BConfig.Log.AccessLogs = true
	//set allow assets
	utils.AllowAssets, _ = utils.GetAssetStrFromSettings()
	initServiceConfig()
	//TODO: Display follow settings on DB. Default if off
	initMicroserviceClient()
	beego.Run()
}

func initMicroserviceClient() {
	//init auth client
	services.CheckAndInitAuthClient()
	//init assets client
	services.CheckAndInitAssetsClient()
	//init chat client
	services.CheckAndInitChatClient()
}

func initServiceConfig() {
	utils.GetAuthHost()
	utils.GetChatHost()
	utils.GetAssetsHost()
}

// add func to func map of beego
func addFuncMap() {
	beego.AddFuncMap("minusInt", minusInt)
	beego.AddFuncMap("addInt", addInt)
	beego.AddFuncMap("minusFloat", minusFloat)
	beego.AddFuncMap("addFloat", addFloat)
	beego.AddFuncMap("round2Decimals", round2Decimals)
	beego.AddFuncMap("round8Decimals", round8Decimals)
	beego.AddFuncMap("dispFloatAsString", dispFloatAsString)
	beego.AddFuncMap("disp8FloatAsString", disp8FloatAsString)
	beego.AddFuncMap("dispDateTime", dispDateTime)
	beego.AddFuncMap("dispDaysLeft", dispDaysLeft)
	beego.AddFuncMap("tradingStatus", tradingStatus)
	beego.AddFuncMap("tradingStatusBadge", tradingStatusBadge)
	beego.AddFuncMap("toLowercase", toLowercase)
	beego.AddFuncMap("toUppercase", toUppercase)
	beego.AddFuncMap("includeStringArray", includeStringArray)
	beego.AddFuncMap("chatTime", chatTime)
	beego.AddFuncMap("codeStatusColor", codeStatusColor)
	beego.AddFuncMap("upperFirstLetter", upperFirstLetter)
	beego.AddFuncMap("upperFirstCase", upperFirstCase)
	beego.AddFuncMap("dispDate", dispDate)
	beego.AddFuncMap("assetColor", assetColor)
	beego.AddFuncMap("roundDecimalClassWithAsset", roundDecimalClassWithAsset)
	beego.AddFuncMap("assetName", assetName)
}

func assetColor(assetType string) string {
	return utils.GetAssetColor(assetType)
}

func roundDecimalClassWithAsset(assetType string) string {
	switch assetType {
	case string(utils.USDWalletAsset):
		return "amount-number"
	case string(utils.BTCWalletAsset):
		return "btc-amount-number"
	case string(utils.DCRWalletAsset):
		return "dcr-amount-number"
	case string(utils.LTCWalletAsset):
		return "ltc-amount-number"
	default:
		return "amount-number"
	}
}

func codeStatusColor(status int) string {
	statusObj, err := utils.GetUrlCodeStatusFromValue(status)
	if err != nil {
		return "#000"
	}
	return statusObj.CodeStatusColor()
}

func assetName(asset string) string {
	return utils.GetAssetFullName(asset)
}

func upperFirstLetter(str string) string {
	if utils.IsEmpty(str) {
		return str
	}
	return strings.ToUpper(str[:1]) + strings.ToLower(str[1:])
}

func chatTime(createdt int64) string {
	createDate := time.Unix(createdt, 0)
	//if is tody, only display time. else, display month, day, year and time
	now := time.Now()
	if createDate.Year() == now.Year() && createDate.Month() == now.Month() && createDate.Day() == now.Day() {
		return createDate.Format("15:04")
	}
	return createDate.Format("2006/01/02, 15:04")
}

func upperFirstCase(text string) string {
	if utils.IsEmpty(text) {
		return text
	}
	return strings.ToUpper(text[:1]) + text[1:]
}

func includeStringArray(array []string, value string) bool {
	return slices.Contains(array, value)
}

func toLowercase(input string) string {
	return strings.ToLower(input)
}

func toUppercase(input string) string {
	return strings.ToUpper(input)
}

func tradingStatus(statusStr string) string {
	switch statusStr {
	case "trading":
		return "Trading"
	case "denied":
		return "Denied"
	case "success":
		return "Successful Bidding"
	case "completed":
		return "Completed"
	case "expired":
		return "Expired"
	case "comming":
		return "Comming Soon"
	default:
		return ""
	}
}

func tradingStatusBadge(statusStr string) string {
	switch statusStr {
	case "trading":
		return "info"
	case "denied":
		return "danger"
	case "success":
		return "success"
	case "completed":
		return "success"
	case "expired":
		return "danger"
	case "comming":
		return "secondary"
	default:
		return "light"
	}
}

func dispDaysLeft(dateUnix int64) string {
	date := time.Unix(dateUnix, 0)
	now := time.Now()
	endCompare := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	nowCompare := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	days := int64(math.Round(endCompare.Sub(nowCompare).Hours() / 24))
	if days > 1 {
		return fmt.Sprintf("%d days", days)
	} else {
		return fmt.Sprintf("%d day", days)
	}
}

func dispDateTime(dateUnix int64) string {
	date := time.Unix(dateUnix, 0)
	return date.Format("2006/01/02, 15:04:05")
}

func dispFloatAsString(a float64) string {
	return fmt.Sprintf("%.2f", a)
}

func disp8FloatAsString(a float64) string {
	return fmt.Sprintf("%.8f", a)
}

func round2Decimals(a float64) float64 {
	return math.Round((a * 100)) / 100
}

func round8Decimals(a float64) float64 {
	return math.Round((a * 1e8)) / 1e8
}

func minusInt(a, b int) string {
	return strconv.FormatInt(int64(a-b), 10)
}

func addInt(a, b int) string {
	return strconv.FormatInt(int64(a+b), 10)
}

func minusFloat(a, b float64) string {
	return strconv.FormatFloat(a-b, 'f', 2, 64)
}

func addFloat(a, b float64) string {
	return strconv.FormatFloat(a+b, 'f', 2, 64)
}

func dispDate(timeInt int64) string {
	if timeInt == 0 {
		return ""
	}
	return time.Unix(timeInt, 0).Format("2006/01/02, 15:04:05")
}
