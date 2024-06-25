package connection

import "encoding/json"

type (
	Connection struct {
		APIKey     string `json:"api_key"`
		APISecret  string `json:"api_secret"`
		UseTestNet bool   `json:"use_test_net"`
	}
)

// Implement the GetAPIKey method
func (cf *Connection) GetAPIKey() string {
	return cf.APIKey
}

// Implement the SetApiKey method
func (cf *Connection) SetApiKey(key string) {
	cf.APIKey = key
}

// Implement the GetSecretKey method
func (cf *Connection) GetSecretKey() string {
	return cf.APISecret
}

// Implement the SetSecretKey method
func (cf *Connection) SetSecretKey(key string) {
	cf.APISecret = key
}

// Implement the GetUseTestNet method
func (cf *Connection) GetUseTestNet() bool {
	return cf.UseTestNet
}

// Implement the SetUseTestNet method
func (cf *Connection) SetUseTestNet(useTestNet bool) {
	cf.UseTestNet = useTestNet
}

func (cf *Connection) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		APIKey     string `json:"api_key"`
		APISecret  string `json:"api_secret"`
		UseTestNet bool   `json:"use_test_net"`
	}{
		APIKey:     cf.APIKey,
		APISecret:  cf.APISecret,
		UseTestNet: cf.UseTestNet,
	})
}

func (cf *Connection) UnmarshalJSON(data []byte) error {
	temp := &struct {
		APIKey     string `json:"api_key"`
		APISecret  string `json:"api_secret"`
		UseTestNet bool   `json:"use_test_net"`
	}{}
	if err := json.Unmarshal(data, temp); err != nil {
		return err
	}
	cf.APIKey = temp.APIKey
	cf.APISecret = temp.APISecret
	cf.UseTestNet = temp.UseTestNet
	return nil
}

func NewConnection(
	apiKey string,
	apiSecret string,
	useTestNet bool,
) *Connection {
	return &Connection{
		APIKey:     apiKey,
		APISecret:  apiSecret,
		UseTestNet: useTestNet,
	}
}
