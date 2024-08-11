package spot_rest_api

import (
	"sync"

	"github.com/bitly/go-simplejson"
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
		parameters *simplejson.Json
		sign       signature.Sign
	}
)
