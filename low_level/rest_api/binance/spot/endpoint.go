package spot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	BaseAPIMainUrl    = "https://api.binance.com"
	BaseAPITestnetUrl = "https://testnet.binance.vision"
)

func GetAPIBaseUrl(useTestNet ...bool) (endpoint string) {
	if len(useTestNet) > 0 && useTestNet[0] {
		endpoint = BaseAPITestnetUrl
	} else {
		endpoint = BaseAPIMainUrl
	}
	return
}

func ListenKey(apiKey string, useTestNet ...bool) (listenKey string, err error) {
	baseUrl := GetAPIBaseUrl(useTestNet...)
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
