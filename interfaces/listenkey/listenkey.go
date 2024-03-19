package listenkey

type ListGen interface {
	GetListenKey() (string, error)
}
