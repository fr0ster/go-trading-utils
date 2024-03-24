package order

type (
	Order interface {
		Create() error
		Cancel() error
		Check() bool
	}
)
