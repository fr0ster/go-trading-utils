package spot_api

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bitly/go-simplejson"
	common "github.com/fr0ster/turbo-restler/web_stream"

	"github.com/sirupsen/logrus"
)

type DepthUpdate struct {
	LastUpdateID int64      `json:"lastUpdateId"`
	Bids         [][]string `json:"bids"`
	Asks         [][]string `json:"asks"`
}

type DepthStream struct {
	symbol             string
	websocketKeepalive bool
	useTestNet         bool
	doneC              chan struct{}
	stopC              chan struct{}
}

// Функція для парсингу JSON
func (ds *DepthStream) parseDepthUpdateJSON(data []byte) (*DepthUpdate, error) {
	var orderBook DepthUpdate
	err := json.Unmarshal(data, &orderBook)
	if err != nil {
		return nil, err
	}
	return &orderBook, nil
}

func (ds *DepthStream) Start(levels string, rateStr string, callBack func(*DepthUpdate)) {
	wss := GetWsBaseUrl(ds.useTestNet)
	wsURL := fmt.Sprintf("%s/%s@depth%s%s", wss, strings.ToLower(ds.symbol), levels, rateStr)
	common.StartStreamer(
		wsURL,
		func(message *simplejson.Json) {
			// Парсинг JSON
			js, _ := message.MarshalJSON()
			depthUpdate, err := ds.parseDepthUpdateJSON(js)
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
