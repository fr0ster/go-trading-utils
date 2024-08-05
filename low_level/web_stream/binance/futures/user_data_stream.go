package futures_api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bitly/go-simplejson"
	api_common "github.com/fr0ster/go-trading-utils/low_level/common"
	futures_rest "github.com/fr0ster/go-trading-utils/low_level/rest_api/binance/futures"
	api "github.com/fr0ster/go-trading-utils/low_level/rest_api/common"
	common "github.com/fr0ster/go-trading-utils/low_level/web_stream/common"

	"github.com/sirupsen/logrus"
)

type UserDataStream struct {
	apiKey string
}

func (uds *UserDataStream) listenKey(method string, useTestNet ...bool) (listenKey string, err error) {
	baseURL := GetWsBaseUrl(useTestNet...)
	endpoint := "/fapi/v1/listenKey"
	var result map[string]interface{}

	body, err := api.CallAPI(baseURL, method, nil, endpoint, api_common.NewSign(uds.apiKey, ""))
	if err != nil {
		return
	}

	// Парсинг відповіді
	err = json.Unmarshal(body, &result)
	listenKey = result["listenKey"].(string)
	return
}

func (uds *UserDataStream) Start(callBack func(*simplejson.Json), quit chan struct{}, useTestNet ...bool) {
	wss := futures_rest.GetAPIBaseUrl(useTestNet...)
	listenKey, err := uds.listenKey(http.MethodPost, useTestNet...)
	if err != nil {
		logrus.Fatalf("Error getting listen key: %v", err)
	}
	wsURL := fmt.Sprintf("%s/%s", wss, listenKey)
	common.StartStreamer(wsURL, func(message []byte) {
		json, err := api_common.NewJSON(message)
		if err != nil {
			logrus.Fatalf("Error parsing JSON: %v, message: %s", err, message)
		}
		if callBack != nil {
			callBack(json)
		}
	}, quit)
	go func() {
		for {
			select {
			case <-quit:
				_, err := uds.listenKey(http.MethodDelete, useTestNet...)
				if err != nil {
					logrus.Fatalf("Error deleting listen key: %v", err)
				}
				close(quit)
				return
			case <-time.After(60 * time.Minute):
				_, err := uds.listenKey(http.MethodPut, useTestNet...)
				if err != nil {
					logrus.Fatalf("Error refreshing listen key: %v", err)
				}
			}
		}
	}()
}

func NewUserDataStream(apiKey, symbol string) *UserDataStream {
	return &UserDataStream{
		apiKey: apiKey,
	}
}
