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
// type BookTicker struct {
// 	UpdateID     int64  `json:"u"`
// 	Symbol       string `json:"s"`
// 	BestBidPrice string `json:"b"`
// 	BestBidQty   string `json:"B"`
// 	BestAskPrice string `json:"a"`
// 	BestAskQty   string `json:"A"`
// }

// Функція для парсингу JSON
func parseBookTickerJSON(data []byte) (*types.BookTicker, error) {
	var kline types.BookTicker
	err := json.Unmarshal(data, &kline)
	if err != nil {
		return nil, err
	}
	return &kline, nil
}

func BookTickersStream(symbol string, callBack func(*types.BookTicker), quit chan struct{}, useTestNet ...bool) {
	wss := GetAPIBaseUrl(useTestNet...)
	wsURL := fmt.Sprintf("%s/%s@bookTicker", wss, strings.ToLower(symbol))
	common.StartStreamer(wsURL, func(message []byte) {
		// Парсинг JSON
		kline, err := parseBookTickerJSON([]byte(message))
		if err != nil {
			logrus.Fatalf("Error parsing JSON: %v, message: %s", err, message)
		}
		if callBack != nil {
			callBack(kline)
		}
	}, quit)
}
