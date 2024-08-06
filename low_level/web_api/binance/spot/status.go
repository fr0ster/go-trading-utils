package spot_web_api

import (
	"encoding/json"

	web_api "github.com/fr0ster/go-trading-utils/low_level/web_api/common"
)

// Функція для логіну
func (wa *WebApi) Status() (response *LogonResponse, limit []web_api.RateLimit, err error) {
	method := "session.status"

	body, limit, err := web_api.CallWebAPI(wa.waHost, wa.waPath, method, nil)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &response)
	return
}
