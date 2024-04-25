package streams

type Stream interface {
	Start() (doneC, stopC chan struct{}, err error)
	GetEventChannel() chan bool
}
