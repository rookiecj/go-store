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

// InitAction is dispatched when to initialise the store or a subscriber subscribes
type InitAction struct{}

// 상태에 변화를 주지않는 action
type UnitAction struct{}

// 상태를 Init 상태로 되돌리는 action
type ResetAction[S State] struct {
	state S
}
