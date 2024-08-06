package signature

type Sign interface {
	CreateSignature(queryString string) string
	GetAPIKey() string
}
