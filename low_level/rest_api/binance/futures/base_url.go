package futures

import (
	common "github.com/fr0ster/go-trading-utils/low_level/common"
	api "github.com/fr0ster/go-trading-utils/low_level/rest_api/common"
	"github.com/sirupsen/logrus"
)

const (
	BaseAPIMainUrl    = "https://fapi.binance.com"
	BaseAPITestnetUrl = "https://testnet.binancefuture.com"
)

func GetAPIBaseUrl(useTestNet ...bool) (endpoint string) {
	if len(useTestNet) > 0 && useTestNet[0] {
		endpoint = BaseAPITestnetUrl
	} else {
		endpoint = BaseAPIMainUrl
	}
	return
}

func ListenKey(apiKey string, method string, useTestNet ...bool) (listenKey string, err error) {
	baseURL := GetAPIBaseUrl(useTestNet...)
	endpoint := "/fapi/v1/listenKey"

	body, err := api.CallAPI(baseURL, method, nil, endpoint, common.NewSign(apiKey, ""))
	if err != nil {
		return
	}

	json, err := common.NewJSON(body)
	if err != nil {
		logrus.Fatalf("Error parsing JSON: %v, message: %s", err, json)
	}
	listenKey = json.Get("listenKey").MustString()
	return
}
