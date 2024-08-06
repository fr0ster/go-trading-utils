package spot_web_api

import (
	"encoding/json"
	"strconv"
	"time"

	web_api "github.com/fr0ster/go-trading-utils/low_level/web_api/common"
)

// Функція для логіну
func (wa *WebApi) Logon() (response *LogonResponse, limit []web_api.RateLimit, err error) {
	method := "session.logon"
	// Створення параметрів запиту
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	message := "apiKey=" + wa.apiKey + "&timestamp=" + strconv.FormatInt(timestamp, 10)
	signature := wa.sign.CreateSignature(message)

	params := LogonParams{
		APIKey:    wa.apiKey,
		Signature: signature,
		Timestamp: timestamp,
	}

	body, limit, err := web_api.CallWebAPI(wa.waHost, wa.waPath, method, params)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &response)
	return
}
