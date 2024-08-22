package services

import (
	"crassets/pkg/dohttp"
	"crassets/pkg/utils"
	"crassets/pkg/walletlib/assets"
	"fmt"
	"net/http"
	"slices"
	"strings"
)

const (
	binancePriceURL         = "https://api.binance.com/api/v3/ticker/price"
	kucoinPriceURL          = "https://api.kucoin.com/api/v1/market/orderbook/level1"
	kucoinAllTickerPriceURL = "https://api.kucoin.com/api/v1/market/allTickers"
	mexcPriceURL            = "https://api.mexc.com/api/v3/ticker/price"
)

type ticker struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price,string"`
}

const (
	Binance = "binance"
	Kucoin  = "kucoin"
	Mexc    = "mexc"
)

type KucoinPriceResponse struct {
	Code string             `json:"code"`
	Data KucoinResponseData `json:"data"`
}

type KucoinAllTickersResponse struct {
	Code string                       `json:"code"`
	Data KucoinAllTickersResponseData `json:"data"`
}

type KucoinAllTickersResponseData struct {
	Time   int64                  `json:"time"`
	Ticker []KucoinAllTickersData `json:"ticker"`
}

type KucoinAllTickersData struct {
	Symbol           string  `json:"symbol"`
	SymbolName       string  `json:"symbolName"`
	Buy              string  `json:"buy"`
	Sell             string  `json:"sell"`
	BestBidSize      string  `json:"bestBidSize"`
	BestAskSize      string  `json:"bestAskSize"`
	ChangeRate       string  `json:"changeRate"`
	ChangePrice      string  `json:"changePrice"`
	High             string  `json:"high"`
	Low              string  `json:"low"`
	Vol              string  `json:"vol"`
	VolValue         string  `json:"volValue"`
	Last             float64 `json:"last,string"`
	AveragePrice     string  `json:"averagePrice"`
	TakerFeeRate     string  `json:"takerFeeRate"`
	MakerFeeRate     string  `json:"makerFeeRate"`
	TakerCoefficient string  `json:"takerCoefficient"`
	MakerCoefficient string  `json:"makerCoefficient"`
}

type KucoinResponseData struct {
	Time        int64   `json:"time"`
	Sequence    string  `json:"sequence"`
	Price       float64 `json:"price,string"`
	Size        string  `json:"size"`
	BestBid     string  `json:"bestBid"`
	BestBidSize string  `json:"bestBidSize"`
	BestAsk     string  `json:"bestAsk"`
	BestAskSize string  `json:"bestAskSize"`
}

func (s *Server) GetExchangePrice(currency string) (float64, error) {
	//get exchange price on config
	exchangeConfig := s.Conf.Exchange
	if utils.IsEmpty(exchangeConfig) {
		return 0, fmt.Errorf("%s", "The exchange has not been established yet. Check the config file!")
	}
	switch exchangeConfig {
	case Binance:
		return s.GetBinancePrice(currency)
	case Kucoin:
		return s.GetKucoinPrice(currency, "USDT")
	case Mexc:
		return s.GetMexcPrice(currency, "USDT")
	default:
		return 0, fmt.Errorf("%s", "Exhchange failed!")
	}
}

func (s *Server) GetExchangeMultilPrice(currencies []string) (map[string]float64, error) {
	//get exchange price on config
	exchangeConfig := s.Conf.Exchange
	if utils.IsEmpty(exchangeConfig) {
		return nil, fmt.Errorf("%s", "The exchange has not been established yet. Check the config file!")
	}
	switch exchangeConfig {
	case Binance:
		return s.GetBinanceMultilPrice(currencies)
	case Kucoin:
		return s.GetKucoinMultilPrice(currencies)
	case Mexc:
		return s.GetMexcMultilPrice(currencies)
	default:
		return nil, fmt.Errorf("%s", "Exhchange failed!")
	}
}

func (s *Server) GetKucoinPrice(firstCurrency string, secondCurrency string) (float64, error) {
	var symbol = fmt.Sprintf("%s-%s", strings.ToUpper(firstCurrency), strings.ToUpper(secondCurrency))
	query := map[string]string{
		"symbol": symbol,
	}
	req := &dohttp.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: kucoinPriceURL,
		Payload: query,
	}
	var kuCoinRes KucoinPriceResponse
	if err := dohttp.HttpRequest(req, &kuCoinRes); err != nil {
		return 0, err
	}

	if kuCoinRes.Code != "200000" {
		return 0, fmt.Errorf("Get Kucoin %s-%s price failed", firstCurrency, secondCurrency)
	}
	return kuCoinRes.Data.Price, nil
}

func (s *Server) GetMexcPrice(firstCurrency string, secondCurrency string) (float64, error) {
	var symbol = fmt.Sprintf("%s%s", strings.ToUpper(firstCurrency), strings.ToUpper(secondCurrency))
	query := map[string]string{
		"symbol": symbol,
	}
	req := &dohttp.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: mexcPriceURL,
		Payload: query,
	}
	var t ticker
	if err := dohttp.HttpRequest(req, &t); err != nil {
		return 0, err
	}

	return t.Price, nil
}

func (s *Server) GetBinancePrice(currency string) (float64, error) {
	var symbol = fmt.Sprintf("%sUSDT", strings.ToUpper(currency))
	query := map[string]string{
		"symbol": symbol,
	}
	req := &dohttp.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: binancePriceURL,
		Payload: query,
	}
	var t ticker
	if err := dohttp.HttpRequest(req, &t); err != nil {
		return 0, err
	}

	return t.Price, nil
}

func (s *Server) GetBinanceMultilPrice(currencies []string) (map[string]float64, error) {
	symbolArr := make([]string, 0)
	for _, currency := range currencies {
		if currency == assets.USDWalletAsset.String() {
			continue
		}
		symbol := fmt.Sprintf("%sUSDT", strings.ToUpper(currency))
		symbolArr = append(symbolArr, fmt.Sprintf("\"%s\"", symbol))
	}

	symbolParam := fmt.Sprintf("[%s]", strings.Join(symbolArr, ","))
	query := map[string]string{
		"symbols": symbolParam,
	}
	var binancePriceURL = "https://api.binance.com/api/v3/ticker/price"
	req := &dohttp.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: binancePriceURL,
		Payload: query,
	}
	var tickerList []ticker
	if err := dohttp.HttpRequest(req, &tickerList); err != nil {
		return make(map[string]float64), err
	}

	result := make(map[string]float64)
	for _, ticker := range tickerList {
		currency := strings.ReplaceAll(ticker.Symbol, "USDT", "")
		result[strings.ToLower(currency)] = ticker.Price
	}
	return result, nil
}

func (s *Server) GetKucoinMultilPrice(currencies []string) (map[string]float64, error) {
	result := make(map[string]float64)
	for _, currency := range currencies {
		if currency == assets.USDWalletAsset.String() {
			continue
		}
		price, err := s.GetKucoinPrice(currency, "USDT")
		if err != nil {
			return nil, err
		}
		result[currency] = price
	}
	return result, nil
}

func (ts *Server) GetMexcMultilPrice(currencies []string) (map[string]float64, error) {
	result := make(map[string]float64)
	for _, currency := range currencies {
		if currency == assets.USDWalletAsset.String() {
			continue
		}
		price, err := ts.GetMexcPrice(currency, "USDT")
		if err != nil {
			return nil, err
		}
		result[currency] = price
	}
	return result, nil
}

func (s *Server) GetSymbolByExchange(exchange string, currency1 string, currency2 string) string {
	switch exchange {
	case Kucoin:
		return fmt.Sprintf("%s-%s", currency1, currency2)
	default:
		return fmt.Sprintf("%s%s", currency1, currency2)
	}
}

// Get all symbol from crytocurrency list, return: usd rate symbol, all rate symbol
func (s *Server) GetAllSymbolFromChainList(currencies []string, exchange string) ([]string, []string) {
	result := make([]string, 0)
	usdResult := make([]string, 0)
	for _, currency := range currencies {
		for _, currency2 := range currencies {
			if currency2 == currency {
				continue
			}
			var symbol1 = strings.ReplaceAll(currency, assets.USDWalletAsset.String(), "usdt")
			var symbol2 = strings.ReplaceAll(currency2, assets.USDWalletAsset.String(), "usdt")
			var symbol = s.GetSymbolByExchange(exchange, strings.ToUpper(symbol1), strings.ToUpper(symbol2))
			if !slices.Contains(result, symbol) {
				result = append(result, symbol)
			}
			if !slices.Contains(usdResult, symbol) && currency2 == assets.USDWalletAsset.String() {
				usdResult = append(usdResult, symbol)
			}
		}
	}
	return usdResult, result
}

func (s *Server) HanlderExchangeRateCompensation(usdRateMap map[string]float64, allRateMap map[string]float64, currencies []string, exchange string) map[string]float64 {
	result := make(map[string]float64)
	for _, currency := range currencies {
		for _, currency2 := range currencies {
			if currency2 == currency {
				continue
			}
			var symbol1 = strings.ToUpper(currency)
			var symbol2 = strings.ToUpper(currency2)
			var symbol = s.GetSymbolByExchange(exchange, symbol1, symbol2)
			var lowerSymbol = strings.ReplaceAll(strings.ToLower(symbol), "-", "")
			rate, exist := allRateMap[symbol]
			if exist {
				result[lowerSymbol] = rate
			} else {
				//if symbol1 is usd, inverse rate
				if symbol1 == assets.USDWalletAsset.ToStringUpper() {
					inverseRate := usdRateMap[symbol2]
					calcRate := float64(0)
					if inverseRate > 0 {
						calcRate = 1 / inverseRate
					}
					result[lowerSymbol] = calcRate
				} else if symbol2 == assets.USDWalletAsset.ToStringUpper() {
					result[lowerSymbol] = usdRateMap[symbol1]
				} else {
					usdRateSymbol1 := usdRateMap[symbol1]
					usdRateSymbol2 := usdRateMap[symbol2]
					calcRate := float64(0)
					if usdRateSymbol2 > 0 {
						calcRate = usdRateSymbol1 / usdRateSymbol2
					}
					result[lowerSymbol] = calcRate
				}
			}
		}
	}
	return result
}

func (s *Server) GetBinanceAllMultilPrice(currencies []string) (map[string]float64, map[string]float64, error) {
	//Create all symbol from input currencies
	usdSymbol, allSymbol := s.GetAllSymbolFromChainList(currencies, Binance)
	req := &dohttp.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: binancePriceURL,
		Payload: map[string]string{},
	}
	var tickerList []ticker
	if err := dohttp.HttpRequest(req, &tickerList); err != nil {
		return make(map[string]float64), make(map[string]float64), err
	}

	usdRateMap := make(map[string]float64)
	result := make(map[string]float64)
	for _, ticker := range tickerList {
		symbol := ticker.Symbol
		if slices.Contains(allSymbol, symbol) {
			result[strings.ReplaceAll(symbol, "USDT", "USD")] = ticker.Price
		}
		if slices.Contains(usdSymbol, symbol) {
			usdRateMap[strings.ReplaceAll(symbol, "USDT", "")] = ticker.Price
		}
	}
	lowerUsdRateMap := make(map[string]float64)
	for key, value := range usdRateMap {
		lowerUsdRateMap[strings.ToLower(key)] = value
	}
	result = s.HanlderExchangeRateCompensation(usdRateMap, result, currencies, Binance)
	return lowerUsdRateMap, result, nil
}

func (s *Server) GetKucoinAllMultilPrice(currencies []string) (map[string]float64, map[string]float64, error) {
	//Create all symbol from input currencies
	usdSymbol, allSymbol := s.GetAllSymbolFromChainList(currencies, Kucoin)
	req := &dohttp.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: kucoinAllTickerPriceURL,
		Payload: map[string]string{},
	}
	var tickerAll KucoinAllTickersResponse
	if err := dohttp.HttpRequest(req, &tickerAll); err != nil {
		return make(map[string]float64), make(map[string]float64), err
	}

	usdRateMap := make(map[string]float64)
	result := make(map[string]float64)
	for _, ticker := range tickerAll.Data.Ticker {
		symbol := ticker.Symbol
		if slices.Contains(allSymbol, symbol) {
			result[strings.ReplaceAll(symbol, "USDT", "USD")] = ticker.Last
		}
		if slices.Contains(usdSymbol, symbol) {
			usdRateMap[strings.ReplaceAll(strings.ReplaceAll(symbol, "USDT", ""), "-", "")] = ticker.Last
		}
	}
	lowerUsdRateMap := make(map[string]float64)
	for key, value := range usdRateMap {
		lowerUsdRateMap[strings.ToLower(key)] = value
	}
	result = s.HanlderExchangeRateCompensation(usdRateMap, result, currencies, Kucoin)
	return lowerUsdRateMap, result, nil
}

func (s *Server) GetMexcAllMultilPrice(currencies []string) (map[string]float64, map[string]float64, error) {
	//Create all symbol from input currencies
	usdSymbol, allSymbol := s.GetAllSymbolFromChainList(currencies, Mexc)
	req := &dohttp.ReqConfig{
		Method:  http.MethodGet,
		HttpUrl: mexcPriceURL,
		Payload: map[string]string{},
	}
	var tickerList []ticker
	if err := dohttp.HttpRequest(req, &tickerList); err != nil {
		return make(map[string]float64), make(map[string]float64), err
	}

	usdRateMap := make(map[string]float64)
	result := make(map[string]float64)
	for _, ticker := range tickerList {
		symbol := ticker.Symbol
		if slices.Contains(allSymbol, symbol) {
			result[strings.ReplaceAll(symbol, "USDT", "USD")] = ticker.Price
		}
		if slices.Contains(usdSymbol, symbol) {
			usdRateMap[strings.ReplaceAll(symbol, "USDT", "")] = ticker.Price
		}
	}
	lowerUsdRateMap := make(map[string]float64)
	for key, value := range usdRateMap {
		lowerUsdRateMap[strings.ToLower(key)] = value
	}
	result = s.HanlderExchangeRateCompensation(usdRateMap, result, currencies, Mexc)
	return lowerUsdRateMap, result, nil
}

// Get all rate. Return : usdratemap, allratemap, error
func (s *Server) GetAllMultilPrice(currencies []string) (map[string]float64, map[string]float64, error) {
	//get exchange price on config
	exchangeConfig := s.Conf.Exchange
	if utils.IsEmpty(exchangeConfig) {
		return nil, nil, fmt.Errorf("%s", "The exchange has not been established yet. Check the config file!")
	}
	switch exchangeConfig {
	case Binance:
		return s.GetBinanceAllMultilPrice(currencies)
	case Kucoin:
		return s.GetKucoinAllMultilPrice(currencies)
	case Mexc:
		return s.GetMexcAllMultilPrice(currencies)
	default:
		return nil, nil, fmt.Errorf("%s", "Exhchange failed!")
	}
}
