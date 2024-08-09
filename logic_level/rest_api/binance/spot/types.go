package spot_rest_api

import (
	"sync"

	rest_api "github.com/fr0ster/turbo-restler/rest_api"
	signature "github.com/fr0ster/turbo-restler/utils/signature"
)

type (
	RestApi struct {
		apiKey     string
		apiSecret  string
		symbol     string
		apiBaseUrl rest_api.ApiBaseUrl
		mutex      *sync.Mutex
		sign       signature.Sign
	}
)
