package futures_api

// Endpoints
const (
	BaseWsMainUrl          = "wss://fstream.binance.com/ws"
	BaseWsTestnetUrl       = "wss://fstream.binancefuture.com/ws"
	BaseCombinedMainURL    = "wss://fstream.binance.com/stream?streams="
	BaseCombinedTestnetURL = "wss://fstream.binancefuture.com/stream?streams="
)

func GetWsBaseUrl(useTestNet ...bool) (endpoint string) {
	if len(useTestNet) > 0 && useTestNet[0] {
		endpoint = BaseWsTestnetUrl
	} else {
		endpoint = BaseWsMainUrl
	}
	return
}
