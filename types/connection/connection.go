package connection

import "encoding/json"

type (
	Connection struct {
		APIKey     string `json:"api_key"`
		APISecret  string `json:"api_secret"`
		UseTestNet bool   `json:"use_test_net"`
	}
)

func (cf *Connection) GetAPIKey() string {
	return cf.APIKey
}

func (cf *Connection) SetApiKey(key string) {
	cf.APIKey = key
}

func (cf *Connection) GetSecretKey() string {
	return cf.APISecret
}

func (cf *Connection) SetSecretKey(key string) {
	cf.APISecret = key
}

func (cf *Connection) GetUseTestNet() bool {
	return cf.UseTestNet
}

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
