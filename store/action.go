package store

// Action triggers a state change.
type Action interface {
	ActionInterface()
}

var (
	// InitAction is dispatched when to initialise the store.
	InitAction = &initAction{}
)

type initAction struct {
}

type ResetAction[S State] struct {
	state S
}

func (c *initAction) ActionInterface() {}

func (*ResetAction[S]) ActionInterface() {}
