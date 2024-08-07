package common

import (
	"encoding/json"
	"fmt"
	"strings"

	common "github.com/fr0ster/turbo-restler/web_stream"

	"github.com/sirupsen/logrus"
)

type AggTradeStream struct {
	symbol             string
	websocketKeepalive bool
	useTestNet         bool
	doneC              chan struct{}
	stopC              chan struct{}
	baseUrl            string
}

// Функція для парсингу JSON
func (ats *AggTradeStream) parseAggTradeJSON(data []byte) (*AggTrade, error) {
	var aggTrade AggTrade
	err := json.Unmarshal(data, &aggTrade)
	if err != nil {
		return nil, err
	}
	return &aggTrade, nil
}

func (ats *AggTradeStream) Start(callBack func(*AggTrade)) (err error) {
	wsURL := fmt.Sprintf("%s/%s@aggTrade", ats.baseUrl, strings.ToLower(ats.symbol))
	ats.doneC, ats.stopC, err = common.StartStreamer(
		wsURL,
		func(message []byte) {
			// Парсинг JSON
			aggTrade, err := ats.parseAggTradeJSON([]byte(message))
			if err != nil {
				logrus.Fatalf("Error parsing JSON: %v, message: %s", err, message)
			}
			if callBack != nil {
				callBack(aggTrade)
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

func NewAggTradeStream(symbol string, useTestNet bool, baseUrl string, websocketKeepalive ...bool) *AggTradeStream {
	var WebsocketKeepalive bool
	if len(websocketKeepalive) > 0 {
		WebsocketKeepalive = websocketKeepalive[0]
	}
	return &AggTradeStream{
		symbol:             symbol,
		websocketKeepalive: WebsocketKeepalive,
		useTestNet:         useTestNet,
		doneC:              make(chan struct{}),
		stopC:              make(chan struct{}),
		baseUrl:            baseUrl,
	}
}
