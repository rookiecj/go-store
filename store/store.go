package store

import (
	"sync"
	"sync/atomic"
)

// Store holds the state of the application.
type Store[S State] interface {
	// Dispatch dispatches an action to the store.
	Dispatch(action Action)
	// Subscribe adds a subscriber to the store.
	// subscribers are notified when the state changes.
	Subscribe(subscriber Subscriber[S]) Store[S]
	// SubscribeOn adds a subscriber to the store.
	// when the state changes, subscribers are notified on the scheduler.
	SubscribeOn(scheduler Scheduler, subscriber Subscriber[S]) Store[S]

	// getState returns the current state of the store.
	getState() S
	//dispatchOn(scheduler Scheduler, action Action)
	// waitForDispatch waits for all dispatched actions to be processed.
	waitForDispatch()
}

// State is value class
type State interface {
	StateInterface()
}

// Reducer reduces the state of the application, it is called in Main context
type Reducer[S State] func(state S, action Action) S

// Subscriber is notified when the state changes.
type Subscriber[S State] func(newState S, oldState S, action Action)

type baseStore[S State] struct {
	state S
	// reducer is called in Main context
	reducer     Reducer[S]
	subscribers []subscriberEntry[S]

	age          int64
	dispatchLock sync.Mutex
}

type subscriberEntry[S State] struct {
	scheduler  Scheduler
	subscriber Subscriber[S]
}

func NewStore[S State](initialState S, reducer Reducer[S]) Store[S] {
	return &baseStore[S]{
		state:   initialState,
		reducer: reducer,
	}
}

func (b *baseStore[S]) Dispatch(action Action) {
	if b == nil {
		return
	}
	// reduce state in Main context
	b.dispatchOn(Main, action)
}

// dispatchOn dispatches an action to the store on the scheduler.
func (b *baseStore[S]) dispatchOn(scheduler Scheduler, action Action) {
	if b == nil {
		return
	}
	scheduler.Schedule(NewTask(func() {

		// reduce
		oldState := b.getState()
		b.state = b.reduce(oldState, action)

		// dispatch
		dispatchScheduler.Schedule(NewTask(func() {
			b.dispatch(oldState, action, b.state)
		}), nil)
	}), nil)
}

func (b *baseStore[S]) Subscribe(subscriber Subscriber[S]) Store[S] {
	if b == nil {
		return b
	}
	return b.SubscribeOn(Main, subscriber)
}

func (b *baseStore[S]) SubscribeOn(scheduler Scheduler, subscriber Subscriber[S]) Store[S] {
	if b == nil {
		return b
	}

	if len(b.subscribers) == 0 {
		b.onFirstSubscribe()
	}

	entry := subscriberEntry[S]{
		scheduler:  scheduler,
		subscriber: subscriber}
	b.subscribers = append(b.subscribers, entry)

	b.dispatchWhenSubscribe(entry, b.state, b.state, InitAction)
	return b
}

func (b *baseStore[S]) getState() (state S) {
	if b == nil {
		return
	}
	return b.state
}

func (b *baseStore[S]) waitForDispatch() {
	b.dispatchLock.Lock()
	b.dispatchLock.Unlock()
}

// reduce should be called in Main context
func (b *baseStore[S]) reduce(state S, action Action) S {
	if b == nil {
		return state
	}
	return b.reducer(state, action)
}

// dispatch state to subscribers in their context
func (b *baseStore[S]) dispatch(oldState S, action Action, newState S) {
	if b == nil {
		return
	}

	// wait for previous dispatching
	b.dispatchLock.Lock()
	if len(b.subscribers) > 0 {
		clonedSubscribers := b.subscribers[:]
		age := atomic.AddInt64(&b.age, 1)
		wg := sync.WaitGroup{}
		// dispatch state in subscriber's context
		for _, entry := range clonedSubscribers {
			b.doDispatchSubscriberLocked(entry, &wg, age, newState, oldState, action)
		}
		wg.Wait()
	}
	b.dispatchLock.Unlock()
	return
}

func (b *baseStore[S]) dispatchWhenSubscribe(entry subscriberEntry[S], newState S, oldState S, action Action) {
	if b == nil {
		return
	}

	b.dispatchLock.Lock()
	wg := sync.WaitGroup{}
	b.doDispatchSubscriberLocked(entry, &wg, b.age, newState, oldState, action)
	wg.Wait()
	b.dispatchLock.Unlock()
}

func (b *baseStore[S]) doDispatchSubscriberLocked(entry subscriberEntry[S], wg *sync.WaitGroup, age int64, newState S, oldState S, action Action) {
	if wg != nil {
		wg.Add(1)
	}
	entry.scheduler.Schedule(NewTask(func() {
		// Dispatcher not to wait for calling a subscriber
		if wg != nil {
			wg.Done()
		}
		// and call subscriber
		entry.subscriber(newState, oldState, action)
	}), nil)
}

func (b *baseStore[S]) onFirstSubscribe() {

}
