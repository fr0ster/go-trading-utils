package common

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bitly/go-simplejson"
	common "github.com/fr0ster/turbo-restler/web_stream"

	"github.com/sirupsen/logrus"
)

type BookTickersStream struct {
	symbol             string
	websocketKeepalive bool
	useTestNet         bool
	doneC              chan struct{}
	stopC              chan struct{}
	baseUrl            string
}

// Функція для парсингу JSON
func (bts *BookTickersStream) parseBookTickerJSON(data []byte) (*BookTicker, error) {
	var kline BookTicker
	err := json.Unmarshal(data, &kline)
	if err != nil {
		return nil, err
	}
	return &kline, nil
}

func (bts *BookTickersStream) Start(callBack func(*BookTicker)) (err error) {
	wsURL := fmt.Sprintf("%s/%s@bookTicker", bts.baseUrl, strings.ToLower(bts.symbol))
	bts.doneC, bts.stopC, err = common.StartStreamer(
		wsURL,
		func(message *simplejson.Json) {
			// Парсинг JSON
			js, _ := message.MarshalJSON()
			kline, err := bts.parseBookTickerJSON(js)
			if err != nil {
				logrus.Fatalf("Error parsing JSON: %v, message: %s", err, message)
			}
			if callBack != nil {
				callBack(kline)
			}
		},
		func(err error) {
			logrus.Errorf("Error reading from websocket: %v", err)
		})
	if err != nil {
		return
	}
	return
}

func NewBookTickersStream(symbol string, useTestNet bool, baseUrl string, websocketKeepalive ...bool) *BookTickersStream {
	var WebsocketKeepalive bool
	if len(websocketKeepalive) > 0 {
		WebsocketKeepalive = websocketKeepalive[0]
	}
	return &BookTickersStream{
		symbol:             symbol,
		websocketKeepalive: WebsocketKeepalive,
		useTestNet:         useTestNet,
		doneC:              make(chan struct{}),
		stopC:              make(chan struct{}),
		baseUrl:            baseUrl,
	}
}
