package store

import (
	"errors"

	"github.com/rookiecj/go-store/sched"
)

var (
	// ErrSkipReducing is returned by a reducer to stop reducing further.
	ErrSkipReducing = errors.New("skip reducing")
)

// Store holds the state of the application.
type Store[S State] interface {

	// AddReducer adds a reducer to the store.
	AddReducer(reducer Reducer[S]) Store[S]

	// Dispatch dispatches an action to the store.
	Dispatch(action Action)

	// Subscribe adds a subscriber to the store.
	// subscribers are notified when the state changes.
	Subscribe(subscriber Subscriber[S]) Store[S]

	// SubscribeOn adds a subscriber to the store.
	// when the state changes, subscribers are notified on the scheduler.
	SubscribeOn(scheduler sched.Scheduler, subscriber Subscriber[S]) Store[S]

	// getState returns the current state of the store.
	getState() S

	// waitForDispatch waits for the dispatcher to stop
	waitForDispatch()
}

// State is value class
type State interface {
	StateInterface()
}

// Reducer reduces the state of the application, it is called in Main context
// error can return ErrSkipReducing to stop reducing further
type Reducer[S State] func(state S, action Action) (S, error)

// Subscriber is notified when the state changes.
type Subscriber[S State] func(newState S, oldState S, action Action)
