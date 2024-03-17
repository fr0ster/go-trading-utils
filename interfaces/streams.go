package interfaces

type Stream interface {
	Start() (doneC, stopC chan struct{}, err error)
}
