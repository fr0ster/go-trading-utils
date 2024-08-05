package futures

import (
	"encoding/json"
	"net/http"

	common "github.com/fr0ster/go-trading-utils/low_level/common"
	api "github.com/fr0ster/go-trading-utils/low_level/rest_api/common"
)

const (
	BaseAPIMainUrl    = "https://fapi.binance.com"
	BaseAPITestnetUrl = "https://testnet.binancefuture.com"
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
	baseURL := GetAPIBaseUrl(useTestNet...)
	endpoint := "/fapi/v1/listenKey"
	var result map[string]interface{}
	// // Створення запиту
	// req, err := http.NewRequest("POST", url, nil)
	// if err != nil {
	// 	fmt.Println("Error creating request:", err)
	// 	return
	// }

	// // Додавання заголовків
	// req.Header.Set("X-MBX-APIKEY", apiKey)

	// // Виконання запиту
	// client := &http.Client{}
	// resp, err := client.Do(req)
	// if err != nil {
	// 	fmt.Println("Error making request:", err)
	// 	return
	// }
	// defer resp.Body.Close()

	// // Читання відповіді
	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	fmt.Println("Error reading response:", err)
	// 	return
	// }

	// // Перевірка статусу відповіді
	// if resp.StatusCode != http.StatusOK {
	// 	fmt.Printf("Error: status code %d\n", resp.StatusCode)
	// 	fmt.Println(string(body))
	// 	return
	// }

	body, err := api.CallAPI(baseURL, http.MethodPost, nil, endpoint, common.NewSign(apiKey, ""))
	if err != nil {
		return
	}

	// Парсинг відповіді
	err = json.Unmarshal(body, &result)
	if err != nil {
		return
	}
	listenKey = result["listenKey"].(string)
	return
}
