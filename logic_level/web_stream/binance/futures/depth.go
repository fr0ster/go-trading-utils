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

type DepthStream struct {
	symbol             string
	websocketKeepalive bool
	useTestNet         bool
	doneC              chan struct{}
	stopC              chan struct{}
}

// Функція для парсингу JSON
func (ds *DepthStream) parseFuturesDepthUpdateJSON(data []byte) (*DepthUpdate, error) {
	var depthUpdate DepthUpdate
	err := json.Unmarshal(data, &depthUpdate)
	if err != nil {
		return nil, err
	}
	return &depthUpdate, nil
}

func (ds *DepthStream) Start(symbol string, levels string, rateStr string, callBack func(*DepthUpdate)) (err error) {
	baseUrl := GetWsBaseUrl(ds.useTestNet)
	wsURL := fmt.Sprintf("%s/%s@depth%s%s", baseUrl, strings.ToLower(symbol), levels, rateStr)
	ds.doneC, ds.stopC, err = common.StartStreamer(
		wsURL,
		func(message []byte) {
			// Парсинг JSON
			depthUpdate, err := ds.parseFuturesDepthUpdateJSON([]byte(message))
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
	if err != nil {
		return
	}
	return
}

func NewDepthStream(symbol string, useTestNet bool, websocketKeepalive ...bool) *DepthStream {
	var WebsocketKeepalive bool
	if len(websocketKeepalive) > 0 {
		WebsocketKeepalive = websocketKeepalive[0]
	}
	return &DepthStream{
		symbol:             symbol,
		websocketKeepalive: WebsocketKeepalive,
		useTestNet:         useTestNet,
		doneC:              make(chan struct{}),
		stopC:              make(chan struct{}),
	}
}
