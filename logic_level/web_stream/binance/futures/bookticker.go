package futures_api

import (
	"encoding/json"
	"fmt"
	"strings"

	types "github.com/fr0ster/go-trading-utils/logic_level/web_stream/binance/common"
	common "github.com/fr0ster/go-trading-utils/low_level/web_stream"

	"github.com/sirupsen/logrus"
)

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
	wss := GetWsBaseUrl(useTestNet...)
	wsURL := fmt.Sprintf("%s/%s@bookTicker", wss, strings.ToLower(symbol))
	common.StartStreamer(
		wsURL,
		func(message []byte) {
			// Парсинг JSON
			kline, err := parseBookTickerJSON([]byte(message))
			if err != nil {
				logrus.Fatalf("Error parsing JSON: %v, message: %s", err, message)
			}
			if callBack != nil {
				callBack(kline)
			}
		},
		func(err error) {
			logrus.Fatalf("Error reading from websocket: %v", err)
		})
}
