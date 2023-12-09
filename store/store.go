package store

import (
	"errors"
	"sync"
	"sync/atomic"
)

// Store holds the state of the application.
type Store[S State] interface {
	// GetState returns the current state of the store.
	GetState() S
	// Dispatch dispatches an action to the store.
	Dispatch(action Action)
	// Subscribe adds a subscriber to the store.
	// subscribers are notified when the state changes.
	Subscribe(subscriber Subscriber[S]) Store[S]
}

// State is value class
type State interface {
	stateInterface()
}

// Reducer reduces the state of the application.
type Reducer[S State] func(state S, action Action) S

// Subscriber is notified when the state changes.
type Subscriber[S State] func(newState S, oldState S, action Action)

type baseStore[S State] struct {
	state       S
	reducer     Reducer[S]
	subscribers []Subscriber[S]
	//lock         sync.RWMutex

	age          int64
	dispatchLock sync.Mutex
}

func NewStore[S State](initialState S, reducer Reducer[S]) Store[S] {
	return &baseStore[S]{
		state:   initialState,
		reducer: reducer,
	}
}

func (b *baseStore[State]) GetState() (state State) {
	if b == nil {
		return
	}
	return b.state
}

func (b *baseStore[State]) Dispatch(action Action) {
	if b == nil {
		return
	}

	// reduces state in Main scheduler
	oldState := b.GetState()
	b.state = b.reduce(oldState, action)

	// dispatch state to subscriber in their context
	b.dispatch(oldState, action, b.state)
}

func (b *baseStore[S]) Subscribe(subscriber Subscriber[S]) Store[S] {
	if b == nil {
		return b
	}

	if len(b.subscribers) == 0 {
		// onFirstSubscriber
	}

	b.subscribers = append(b.subscribers, subscriber)

	// schedule the task on caller's context
	b.dispatchLock.Lock()
	b.doDispatchOn(nil, b.age, Immediate, subscriber, b.state, b.state, InitAction)
	b.dispatchLock.Unlock()
	return b
}

func (b *baseStore[S]) reduce(oldState S, action Action) S {
	if b == nil {
		return oldState
	}
	// reduce on Main context
	return b.reducer(oldState, action)
}

func (b *baseStore[S]) dispatch(oldState S, action Action, newState S) error {
	if b == nil {
		return errors.New("store is nil")
	}

	// dispatch
	b.dispatchLock.Lock()
	clonedSubscribers := b.subscribers[:]
	wg := sync.WaitGroup{}
	age := atomic.AddInt64(&b.age, 1)
	for _, subscriber := range clonedSubscribers {
		b.doDispatchOn(&wg, age, Immediate, subscriber, newState, oldState, action)
	}
	wg.Wait()
	b.dispatchLock.Unlock()
	return nil
}

func (b *baseStore[S]) doDispatchOn(wg *sync.WaitGroup, age int64, scheduler Scheduler, subscriber Subscriber[S], newState S, oldState S, action Action) {
	if wg != nil {
		wg.Add(1)
	}
	scheduler.Schedule(NewTask(age, subscriber, newState, oldState, action), func() {
		if wg != nil {
			wg.Done()
		}
	})
}
