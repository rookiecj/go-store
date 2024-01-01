package store

// Action triggers a state change.
type Action interface {
	ActionInterface()
}

// AsyncAction provides a way to dispatch actions asynchronously.
type AsyncAction interface {
	Action

	// Run is called with Dispatcher which can be used to dispatch actions.
	Run(Dispatcher)
}

// Dispatcher dispatches an action.
type Dispatcher interface {

	// Dispatch dispatches an action.
	Dispatch(Action)
}

var (
	// InitAction is dispatched when to initialise the store or a subscriber subscribes
	InitAction = &initAction{}
)

type initAction struct {
}

type ResetAction[S State] struct {
	state S
}

func (c *initAction) ActionInterface() {}

func (*ResetAction[S]) ActionInterface() {}
