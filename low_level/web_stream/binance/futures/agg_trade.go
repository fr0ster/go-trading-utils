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
// type AggTrade struct {
// 	EventType    string `json:"e"`
// 	EventTime    int64  `json:"E"`
// 	TradeID      int64  `json:"a"`
// 	Symbol       string `json:"s"`
// 	Price        string `json:"p"`
// 	Quantity     string `json:"q"`
// 	FirstTradeID int64  `json:"f"`
// 	LastTradeID  int64  `json:"l"`
// 	TradeTime    int64  `json:"T"`
// 	IsBuyerMaker bool   `json:"m"`
// }

// Функція для парсингу JSON
func parseAggTradeJSON(data []byte) (*types.AggTrade, error) {
	var aggTrade types.AggTrade
	err := json.Unmarshal(data, &aggTrade)
	if err != nil {
		return nil, err
	}
	return &aggTrade, nil
}

func AggTradeStream(symbol string, callBack func(*types.AggTrade), quit chan struct{}, useTestNet ...bool) {
	baseUrl := GetWsBaseUrl(useTestNet...)
	wsURL := fmt.Sprintf("%s/%s@aggTrade", baseUrl, strings.ToLower(symbol))
	common.StartStreamer(wsURL, func(message []byte) {
		// Парсинг JSON
		aggTrade, err := parseAggTradeJSON([]byte(message))
		if err != nil {
			logrus.Fatalf("Error parsing JSON: %v, message: %s", err, message)
		}
		if callBack != nil {
			callBack(aggTrade)
		}
	}, quit)
}
