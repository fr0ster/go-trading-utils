package spot_api

// Endpoints
const (
	BaseWsMainURL          = "wss://stream.binance.com:9443/ws"
	BaseWsTestnetURL       = "wss://testnet.binance.vision/ws"
	BaseCombinedMainURL    = "wss://stream.binance.com:9443/stream?streams="
	BaseCombinedTestnetURL = "wss://testnet.binance.vision/stream?streams="
)

func GetWsEndpoint(useTestNet ...bool) (endpoint string) {
	if len(useTestNet) > 0 && useTestNet[0] {
		endpoint = BaseWsTestnetURL
	} else {
		endpoint = BaseWsMainURL
	}
	return
}
