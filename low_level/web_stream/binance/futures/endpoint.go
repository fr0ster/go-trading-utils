package futures_api

// Endpoints
const (
	BaseWsMainUrl          = "wss://fstream.binance.com/ws"
	BaseWsTestnetUrl       = "wss://stream.binancefuture.com/ws"
	BaseCombinedMainURL    = "wss://fstream.binance.com/stream?streams="
	BaseCombinedTestnetURL = "wss://stream.binancefuture.com/stream?streams="
)

func GetWsEndpoint(useTestNet ...bool) (endpoint string) {
	if len(useTestNet) > 0 && useTestNet[0] {
		endpoint = BaseWsTestnetUrl
	} else {
		endpoint = BaseWsMainUrl
	}
	return
}
