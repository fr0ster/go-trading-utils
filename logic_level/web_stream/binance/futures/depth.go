package futures_api

import (
	"encoding/json"
	"fmt"
	"strings"

	common "github.com/fr0ster/turbo-restler/web_stream"

	"github.com/sirupsen/logrus"
)

// Визначення структури для JSON
type DepthUpdate struct {
	EventType     string     `json:"e"`
	EventTime     int64      `json:"E"`
	TransactTime  int64      `json:"T"`
	Symbol        string     `json:"s"`
	FirstUpdateID int64      `json:"U"`
	LastUpdateID  int64      `json:"u"`
	PrevUpdateID  int64      `json:"pu"`
	Bids          [][]string `json:"b"`
	Asks          [][]string `json:"a"`
}

// Функція для парсингу JSON
func parseFuturesDepthUpdateJSON(data []byte) (*DepthUpdate, error) {
	var depthUpdate DepthUpdate
	err := json.Unmarshal(data, &depthUpdate)
	if err != nil {
		return nil, err
	}
	return &depthUpdate, nil
}

func DepthStream(symbol string, levels string, rateStr string, callBack func(*DepthUpdate), quit chan struct{}, useTestNet ...bool) {
	baseUrl := GetWsBaseUrl(useTestNet...)
	wsURL := fmt.Sprintf("%s/%s@depth%s%s", baseUrl, strings.ToLower(symbol), levels, rateStr)
	common.StartStreamer(
		wsURL,
		func(message []byte) {
			// Парсинг JSON
			depthUpdate, err := parseFuturesDepthUpdateJSON([]byte(message))
			if err != nil {
				logrus.Fatalf("Error parsing JSON: %v, message: %s", err, message)
			}
			if callBack != nil {
				callBack(depthUpdate)
			}
		},
		func(err error) {
			logrus.Fatalf("Error reading from websocket: %v", err)
		})
}
