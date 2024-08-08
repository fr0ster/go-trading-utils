package spot_web_api

import (
	"fmt"

	web_api "github.com/fr0ster/turbo-restler/web_api"
)

// Функція для логіну
func (wa *WebApi) Logout() (result *LogonResult, err error) {
	method := "session.logout"

	response, err := web_api.CallWebAPI(wa.waHost, wa.waPath, method, nil, nil)
	if err != nil {
		return
	}

	if response.Status != 200 {
		err = fmt.Errorf("error request: %v", response.Error)
		return
	}

	result = response.Result.(*LogonResult)
	return
}
