package futures_api

import (
	"encoding/json"
	"fmt"
	"strings"

	types "github.com/fr0ster/go-trading-utils/logic_level/web_stream/binance/common"
	common "github.com/fr0ster/turbo-restler/web_stream"

	"github.com/sirupsen/logrus"
)

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
	baseUrl := GetWsBaseUrl(useTestNet...)
	wsURL := fmt.Sprintf("%s/%s@kline_%s", baseUrl, strings.ToLower(symbol), interval)
	common.StartStreamer(
		wsURL,
		func(message []byte) {
			// Парсинг JSON
			kline, err := parseKlineJSON([]byte(message))
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
