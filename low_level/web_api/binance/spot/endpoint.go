package spot_web_api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	BaseWsMainUrl    = "wss://ws-api.binance.com:443/ws-api/v3"
	BaseWsTestnetUrl = "wss://testnet.binance.vision/ws-api/v3"
)

func GetWsBaseUrl(useTestNet ...bool) (endpoint string) {
	if len(useTestNet) > 0 && useTestNet[0] {
		endpoint = BaseWsTestnetUrl
	} else {
		endpoint = BaseWsMainUrl
	}
	return
}

func ListenKey(apiKey string, useTestNet ...bool) (listenKey string, err error) {
	baseUrl := GetWsBaseUrl(useTestNet...)
	url := fmt.Sprintf("%s/api/v3/userDataStream", baseUrl)
	var result map[string]interface{}
	// Створення запиту
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Додавання заголовків
	req.Header.Set("X-MBX-APIKEY", apiKey)

	// Виконання запиту
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	// Читання відповіді
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	// Перевірка статусу відповіді
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: status code %d\n", resp.StatusCode)
		fmt.Println(string(body))
		return
	}

	// Парсинг відповіді
	err = json.Unmarshal(body, &result)
	listenKey = result["listenKey"].(string)
	return
}
