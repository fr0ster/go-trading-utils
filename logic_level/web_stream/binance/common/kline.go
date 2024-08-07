package common

import (
	"encoding/json"
	"fmt"
	"strings"

	common "github.com/fr0ster/turbo-restler/web_stream"

	"github.com/sirupsen/logrus"
)

type KlinesStream struct {
	symbol             string
	websocketKeepalive bool
	useTestNet         bool
	doneC              chan struct{}
	stopC              chan struct{}
	baseUrl            string
}

// Функція для парсингу JSON
func (ks *KlinesStream) parseKlineJSON(data []byte) (*Kline, error) {
	var kline Kline
	err := json.Unmarshal(data, &kline)
	if err != nil {
		return nil, err
	}
	return &kline, nil
}

func (ks *KlinesStream) Start(symbol string, interval string, callBack func(*Kline)) (err error) {
	wsURL := fmt.Sprintf("%s/%s@kline_%s", ks.baseUrl, strings.ToLower(symbol), interval)
	ks.doneC, ks.stopC, err = common.StartStreamer(
		wsURL,
		func(message []byte) {
			// Парсинг JSON
			kline, err := ks.parseKlineJSON([]byte(message))
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
	if err != nil {
		return
	}
	return
}

func NewKlineStream(symbol string, useTestNet bool, baseUrl string, websocketKeepalive ...bool) *KlinesStream {
	var WebsocketKeepalive bool
	if len(websocketKeepalive) > 0 {
		WebsocketKeepalive = websocketKeepalive[0]
	}
	return &KlinesStream{
		symbol:             symbol,
		websocketKeepalive: WebsocketKeepalive,
		useTestNet:         useTestNet,
		doneC:              make(chan struct{}),
		stopC:              make(chan struct{}),
		baseUrl:            baseUrl,
	}
}
