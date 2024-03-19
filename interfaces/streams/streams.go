package streams

type Stream interface {
	Start() (doneC, stopC chan struct{}, err error)
	GetStreamEvent() chan bool
}
