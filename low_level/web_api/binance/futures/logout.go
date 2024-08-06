package spot_web_api

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	web_api "github.com/fr0ster/go-trading-utils/low_level/web_api/common"
)

// Функція для логіну
func (wa *WebApi) Logout() (response *LogonResponse, limit []web_api.RateLimit, err error) {
	request := LogoutRequest{
		ID:     uuid.New().String(),
		Method: "session.logout",
	}

	// Серіалізація запиту в JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		err = fmt.Errorf("error marshaling request: %v", err)
		return
	}

	body, limit, err := web_api.CallWebAPI(wa.waHost, wa.waPath, requestBody)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &response)
	return
}
