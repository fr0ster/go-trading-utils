package connection

import "encoding/json"

type (
	Connection struct {
		APIKey          string  `json:"api_key"`
		APISecret       string  `json:"api_secret"`
		UseTestNet      bool    `json:"use_test_net"`
		CommissionMaker float64 `json:"commission_maker"`
		CommissionTaker float64 `json:"commission_taker"`
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

// Implement the GetCommissionMaker method
func (cf *Connection) GetCommissionMaker() float64 {
	return cf.CommissionMaker
}

// Implement the SetCommissionMaker method
func (cf *Connection) SetCommissionMaker(commission float64) {
	cf.CommissionMaker = commission
}

// Implement the GetCommissionTaker method
func (cf *Connection) GetCommissionTaker() float64 {
	return cf.CommissionTaker
}

// Implement the SetCommissionTaker method
func (cf *Connection) SetCommissionTaker(commission float64) {
	cf.CommissionTaker = commission
}

func (cf *Connection) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		APIKey          string  `json:"api_key"`
		APISecret       string  `json:"api_secret"`
		UseTestNet      bool    `json:"use_test_net"`
		CommissionMaker float64 `json:"commission_maker"`
		CommissionTaker float64 `json:"commission_taker"`
	}{
		APIKey:          cf.APIKey,
		APISecret:       cf.APISecret,
		UseTestNet:      cf.UseTestNet,
		CommissionMaker: cf.CommissionMaker,
		CommissionTaker: cf.CommissionTaker,
	})
}

func (cf *Connection) UnmarshalJSON(data []byte) error {
	temp := &struct {
		APIKey          string  `json:"api_key"`
		APISecret       string  `json:"api_secret"`
		UseTestNet      bool    `json:"use_test_net"`
		CommissionMaker float64 `json:"commission_maker"`
		CommissionTaker float64 `json:"commission_taker"`
	}{}
	if err := json.Unmarshal(data, temp); err != nil {
		return err
	}
	cf.APIKey = temp.APIKey
	cf.APISecret = temp.APISecret
	cf.UseTestNet = temp.UseTestNet
	cf.CommissionMaker = temp.CommissionMaker
	cf.CommissionTaker = temp.CommissionTaker
	return nil
}

func NewConnection(
	apiKey string,
	apiSecret string,
	useTestNet bool,
	commissionMaker float64,
	commissionTaker float64,
) *Connection {
	return &Connection{
		APIKey:          apiKey,
		APISecret:       apiSecret,
		UseTestNet:      useTestNet,
		CommissionMaker: commissionMaker,
		CommissionTaker: commissionTaker,
	}
}
