package common

import (
	"encoding/json"

	web_api "github.com/fr0ster/go-trading-utils/low_level/web_api"
)

// Функція для логіну
func (wa *WebApi) Logout() (response *LogonResponse, limit []web_api.RateLimit, err error) {
	method := "session.logout"

	body, limit, err := web_api.CallWebAPI(wa.waHost, wa.waPath, method, nil)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &response)
	return
}
