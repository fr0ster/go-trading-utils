package connection

type (
	Connection interface {
		GetAPIKey() string
		SetApiKey(key string)
		GetSecretKey() string
		SetSecretKey(key string)
		GetUseTestNet() bool
		SetUseTestNet(useTestNet bool)
		GetCommissionMaker() float64
		SetCommissionMaker(commission float64)
		GetCommissionTaker() float64
		SetCommissionTaker(commission float64)
	}
)
