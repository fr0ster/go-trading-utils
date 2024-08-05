package futures_api

import (
	"encoding/json"
	"fmt"
	"strings"

	types "github.com/fr0ster/go-trading-utils/low_level/web_stream/binance/common"
	common "github.com/fr0ster/go-trading-utils/low_level/web_stream/common"

	"github.com/sirupsen/logrus"
)

// // Визначення структури для JSON
// type Kline struct {
// 	EventType string `json:"e"`
// 	EventTime int64  `json:"E"`
// 	Symbol    string `json:"s"`
// 	KlineData struct {
// 		StartTime       int64  `json:"t"`
// 		CloseTime       int64  `json:"T"`
// 		Symbol          string `json:"s"`
// 		Interval        string `json:"i"`
// 		FirstTradeID    int64  `json:"f"`
// 		LastTradeID     int64  `json:"L"`
// 		OpenPrice       string `json:"o"`
// 		ClosePrice      string `json:"c"`
// 		HighPrice       string `json:"h"`
// 		LowPrice        string `json:"l"`
// 		Volume          string `json:"v"`
// 		NumberOfTrades  int    `json:"n"`
// 		IsFinal         bool   `json:"x"`
// 		QuoteVolume     string `json:"q"`
// 		ActiveBuyVolume string `json:"V"`
// 		ActiveBuyQuote  string `json:"Q"`
// 		Ignore          string `json:"B"`
// 	} `json:"k"`
// }

// Функція для парсингу JSON
func parseKlineJSON(data []byte) (*types.Kline, error) {
	var kline types.Kline
	err := json.Unmarshal(data, &kline)
	if err != nil {
		return nil, err
	}
	return &kline, nil
}

func KlinesStream(symbol string, interval string, callBack func(*types.Kline), quit chan struct{}, useTestNet ...bool) {
	baseUrl := GetWsEndpoint(useTestNet...)
	wsURL := fmt.Sprintf("%s/%s@kline_%s", baseUrl, strings.ToLower(symbol), interval)
	common.StartStreamer(wsURL, func(message []byte) {
		// Парсинг JSON
		kline, err := parseKlineJSON([]byte(message))
		if err != nil {
			logrus.Fatalf("Error parsing JSON: %v, message: %s", err, message)
		}
		if callBack != nil {
			callBack(kline)
		}
	}, quit)
}
