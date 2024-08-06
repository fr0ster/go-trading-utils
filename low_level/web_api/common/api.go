package common

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/websocket"
)

type (
	Response struct {
		ID         string      `json:"id"`
		Status     int         `json:"status"`
		Error      ErrorDetail `json:"error"`
		Result     interface{} `json:"result"`
		RateLimits []RateLimit `json:"rateLimits"`
	}

	ErrorDetail struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}

	RateLimit struct {
		RateLimitType string `json:"rateLimitType"`
		Interval      string `json:"interval"`
		IntervalNum   int    `json:"intervalNum"`
		Limit         int    `json:"limit"`
		Count         int    `json:"count"`
	}
)

func ParseResponse(data []byte) (*Response, error) {
	var response Response
	err := json.Unmarshal(data, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}
	return &response, nil
}

func ParseLimit(data []byte) ([]RateLimit, error) {
	var response []RateLimit
	err := json.Unmarshal(data, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}
	return response, nil
}

// Функція для розміщення ордера через WebSocket
func CallWebAPI(host, path string, requestBody []byte) (response []byte, limits []RateLimit, err error) {
	// Підключення до WebSocket
	u := url.URL{Scheme: "wss", Host: host, Path: path}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		err = fmt.Errorf("error connecting to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	// Відправка запиту на розміщення ордера
	err = conn.WriteMessage(websocket.TextMessage, requestBody)
	if err != nil {
		err = fmt.Errorf("error sending message: %v", err)
		return
	}

	// Читання відповіді
	_, body, err := conn.ReadMessage()
	msg, err := ParseResponse(body)
	if err != nil {
		err = fmt.Errorf("error parsing response: %v", err)
		return
	}
	if msg.Status != 200 {
		err = fmt.Errorf("error response: %v", msg.Error)
	}
	jMap, err := simplejson.NewJson(body)
	if err != nil {
		return
	}
	response, err = jMap.Get("result").Encode()
	if err != nil {
		return
	}
	limit, err := jMap.Get("rateLimits").Encode()
	if err != nil {
		return
	}
	limits, err = ParseLimit(limit)
	return
}
