package store

// Action triggers a state change.
type Action any

// AsyncAction provides a way to dispatch actions asynchronously.
// It is called with Dispatcher which can be used to dispatch actions.
type AsyncAction func(Dispatcher)

// Dispatcher dispatches an action.
type Dispatcher interface {

	// Dispatch dispatches an action.
	Dispatch(Action)
}

var (
	InitAction = &initAction{}
)

// InitAction is dispatched when to initialise the store or a subscriber subscribes
type initAction struct{}

type ResetAction[S State] struct {
	state S
}
