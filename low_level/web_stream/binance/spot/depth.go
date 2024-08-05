package spot_api

import (
	"encoding/json"
	"fmt"
	"strings"

	common "github.com/fr0ster/go-trading-utils/low_level/web_stream/common"

	"github.com/sirupsen/logrus"
)

type DepthUpdate struct {
	LastUpdateID int64      `json:"lastUpdateId"`
	Bids         [][]string `json:"bids"`
	Asks         [][]string `json:"asks"`
}

// Функція для парсингу JSON
func parseDepthUpdateJSON(data []byte) (*DepthUpdate, error) {
	var orderBook DepthUpdate
	err := json.Unmarshal(data, &orderBook)
	if err != nil {
		return nil, err
	}
	return &orderBook, nil
}

func DepthStream(symbol string, levels string, rateStr string, callBack func(*DepthUpdate), quit chan struct{}, useTestNet ...bool) {
	wss := GetAPIBaseUrl(useTestNet...)
	wsURL := fmt.Sprintf("%s/%s@depth%s%s", wss, strings.ToLower(symbol), levels, rateStr)
	common.StartStreamer(wsURL, func(message []byte) {
		// Парсинг JSON
		depthUpdate, err := parseDepthUpdateJSON([]byte(message))
		if err != nil {
			logrus.Fatalf("Error parsing JSON: %v, message: %s", err, message)
		}
		if callBack != nil {
			callBack(depthUpdate)
		}
	}, quit)
}
