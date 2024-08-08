package futures_web_api

import (
	"fmt"
	"net/url"

	web_api "github.com/fr0ster/turbo-restler/web_api"
)

type (
	LogonParams struct {
		APIKey    string `json:"apiKey"`
		Signature string `json:"signature"`
		Timestamp int64  `json:"timestamp"`
	}
	LogonResult struct {
		APIKey           string `json:"apiKey"`
		AuthorizedSince  int64  `json:"authorizedSince"`
		ConnectedSince   int64  `json:"connectedSince"`
		ReturnRateLimits bool   `json:"returnRateLimits"`
		ServerTime       int64  `json:"serverTime"`
	}
)

// Функція для логіну
func (wa *WebApi) Logon() (result *LogonResult, err error) {
	// Створення параметрів запиту
	params := url.Values{}
	params.Set("apiKey", wa.apiKey)

	response, err := web_api.CallWebAPI(wa.waHost, wa.waPath, "session.logon", params, nil)
	if err != nil {
		return
	}

	if response.Status != 200 {
		err = fmt.Errorf("error request: %v", response.Error)
		return
	}

	result = response.Result.(*LogonResult)
	return
}
