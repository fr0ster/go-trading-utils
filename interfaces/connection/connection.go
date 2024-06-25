package connection

type (
	Connection interface {
		GetAPIKey() string
		SetApiKey(key string)
		GetSecretKey() string
		SetSecretKey(key string)
		GetUseTestNet() bool
		SetUseTestNet(useTestNet bool)
	}
)
