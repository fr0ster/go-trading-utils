package interfaces

type (
	Item interface {
		GetItem() *Item
	}
	Tree interface {
		Lock()
		Unlock()
		GetTree() *Tree
	}
)
