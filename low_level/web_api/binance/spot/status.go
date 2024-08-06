package spot_web_api

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	web_api "github.com/fr0ster/go-trading-utils/low_level/web_api/common"
)

// Функція для логіну
func (wa *WebApi) Status() (response []byte, limit []web_api.RateLimit, err error) {

	request := StatusRequest{
		ID:     uuid.New().String(),
		Method: "session.status",
	}

	// Серіалізація запиту в JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		err = fmt.Errorf("error marshaling request: %v", err)
		return
	}

	return web_api.CallWebAPI(wa.waHost, wa.waPath, requestBody)
}
