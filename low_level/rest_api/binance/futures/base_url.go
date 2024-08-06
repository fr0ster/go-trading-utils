package futures

// import (
// common "github.com/fr0ster/go-trading-utils/low_level/common"
// signature "github.com/fr0ster/go-trading-utils/low_level/common/signature"
// api "github.com/fr0ster/go-trading-utils/low_level/rest_api/common"
// "github.com/sirupsen/logrus"
// )

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
