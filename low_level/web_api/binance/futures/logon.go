package spot_web_api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

	web_api "github.com/fr0ster/go-trading-utils/low_level/web_api/common"
)

// Функція для логіну
func (wa *WebApi) Logon() (response *LogonResponse, limit []web_api.RateLimit, err error) {
	// Створення параметрів запиту
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	message := "apiKey=" + wa.apiKey + "&timestamp=" + strconv.FormatInt(timestamp, 10)
	signature := wa.sign.CreateSignature(message)

	params := LogonParams{
		APIKey:    wa.apiKey,
		Signature: signature,
		Timestamp: timestamp,
	}

	request := LogonRequest{
		ID:     uuid.New().String(),
		Method: "session.logon",
		Params: params,
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
