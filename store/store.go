package store

import (
	"github.com/rookiecj/go-store/logger"
	"github.com/rookiecj/go-store/sched"
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
	SubscribeOn(scheduler sched.Scheduler, subscriber Subscriber[S]) Store[S]

	// getState returns the current state of the store.
	getState() S

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
	subscribers []*subscriberEntry[S]

	// reduce and dispatch context
	dispatchScheduler sched.Scheduler
	age               int64
	dispatchLock      *sync.Mutex
}

type subscriberEntry[S State] struct {
	scheduler  sched.Scheduler
	subscriber Subscriber[S]
}

func NewStore[S State](initialState S, reducer Reducer[S]) Store[S] {
	return NewStoreOn(sched.Immediate, initialState, reducer)
}

func NewStoreOn[S State](scheduler sched.Scheduler, initialState S, reducer Reducer[S]) Store[S] {
	return &baseStore[S]{
		state:             initialState,
		reducer:           reducer,
		dispatchScheduler: scheduler,
		age:               0,
		dispatchLock:      &sync.Mutex{},
	}
}

func (b *baseStore[S]) Dispatch(action Action) {
	if b == nil {
		return
	}
	// reduce state in Main context
	b.dispatchOn(b.dispatchScheduler, action)
}

// dispatchOn dispatches an action to the store on the scheduler.
func (b *baseStore[S]) dispatchOn(scheduler sched.Scheduler, action Action) {
	if b == nil {
		return
	}
	scheduler.Schedule(func() {
		logger.Infof("reduce: action:%v\n", action)
		// reduce
		oldState := b.getState()
		b.state = b.reduce(oldState, action)
		// dispatch
		logger.Infof("dispatch: state %v with action: %v", b.state, action)
		b.dispatch(oldState, action, b.state)
	})
}

func (b *baseStore[S]) Subscribe(subscriber Subscriber[S]) Store[S] {
	if b == nil {
		return b
	}
	return b.SubscribeOn(b.dispatchScheduler, subscriber)
}

func (b *baseStore[S]) SubscribeOn(scheduler sched.Scheduler, subscriber Subscriber[S]) Store[S] {
	if b == nil {
		return b
	}

	if len(b.subscribers) == 0 {
		b.onFirstSubscribe()
	}

	if scheduler == nil {
		scheduler = b.dispatchScheduler
	}

	entry := subscriberEntry[S]{
		scheduler:  scheduler,
		subscriber: subscriber}

	// dispatch before adding to subscribers
	b.dispatchWhenSubscribe(&entry, b.state, b.state, InitAction)

	b.subscribers = append(b.subscribers, &entry)

	return b
}

func (b *baseStore[S]) getState() (state S) {
	if b == nil {
		return
	}
	return b.state
}

func (b *baseStore[S]) waitForDispatch() {
	b.dispatchScheduler.StopWait()
}

// reduce should be called in the same(Main) context
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
		wg := &sync.WaitGroup{}
		// dispatch state in subscriber's context
		for _, entry := range clonedSubscribers {
			b.doDispatchSubscriberLocked(entry, wg, age, newState, oldState, action)
		}
		wg.Wait()
	}
	b.dispatchLock.Unlock()
	return
}

func (b *baseStore[S]) dispatchWhenSubscribe(entry *subscriberEntry[S], newState S, oldState S, action Action) {
	if b == nil {
		return
	}

	b.dispatchLock.Lock()
	wg := sync.WaitGroup{}
	b.doDispatchSubscriberLocked(entry, &wg, b.age, newState, oldState, action)
	wg.Wait()
	b.dispatchLock.Unlock()
}

func (b *baseStore[S]) doDispatchSubscriberLocked(entry *subscriberEntry[S], wg *sync.WaitGroup, age int64, newState S, oldState S, action Action) {

	// not schedule but run a task here, not to make deadlock on dispatcher
	if entry.scheduler == b.dispatchScheduler {
		entry.subscriber(newState, oldState, action)
	} else {
		if wg != nil {
			wg.Add(1)
		}
		entry.scheduler.Schedule(func() {
			// waits up Dispatcher
			if wg != nil {
				wg.Done()
			}

			// and call subscriber
			entry.subscriber(newState, oldState, action)

		})
	}
}

func (b *baseStore[S]) onFirstSubscribe() {

}
