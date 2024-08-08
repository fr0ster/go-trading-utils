package spot_web_api

import (
	"fmt"

	web_api "github.com/fr0ster/turbo-restler/web_api"
)

// Функція для перевірки статусу сесії
func (wa *WebApi) Status() (result *LogonResult, err error) {
	response, err := web_api.CallWebAPI(wa.waHost, wa.waPath, "session.status", nil, nil)
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
