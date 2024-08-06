package spot

// import (
// common "github.com/fr0ster/turbo-restler/common"
// signature "github.com/fr0ster/turbo-restler/common/signature"
// api "github.com/fr0ster/turbo-restler/rest_api/common"
// "github.com/sirupsen/logrus"
// )

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
