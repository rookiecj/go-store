package store

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/rookiecj/go-store/logger"
	"github.com/rookiecj/go-store/sched"
)

type baseStore[S State] struct {
	state S
	// reducer is called in dispatcher context
	reducers    []Reducer[S]
	subscribers []*subscriberEntry[S]

	// reduce and dispatch context
	dispatchScheduler sched.Scheduler
	dispatchStarted   bool
	age               int64
	dispatchLock      *sync.Mutex
}

type subscriberEntry[S State] struct {
	scheduler  sched.Scheduler
	subscriber Subscriber[S]
}

type baseDisposer struct {
	dispose func()
}

func (b *baseDisposer) Dispose() {
	if b == nil {
		return
	}
	b.dispose()
}

// NewStore creates a store with a reducer and an initial state.
func NewStore[S State](initialState S, reducer Reducer[S]) Store[S] {
	return NewStoreOn(sched.NewMainScheduler(), initialState, reducer)
}

// NewStoreOn Scheduler should ensure actions to be reduced in order
func NewStoreOn[S State](scheduler sched.Scheduler, initialState S, reducer Reducer[S]) Store[S] {
	return &baseStore[S]{
		state:             initialState,
		reducers:          []Reducer[S]{reducer},
		dispatchScheduler: scheduler,
		age:               0,
		dispatchLock:      &sync.Mutex{},
	}
}

func (b *baseStore[S]) AddReducer(reducer Reducer[S]) Store[S] {
	if b == nil {
		return b
	}

	b.reducers = append(b.reducers, reducer)

	return b
}

func (b *baseStore[S]) Dispatch(action Action) {
	if b == nil {
		return
	}

	b.ensureScheduler()

	// reduce state in dispatcher context
	b.dispatchOn(b.dispatchScheduler, action)
}

func (b *baseStore[S]) ensureScheduler() {
	if b == nil {
		return
	}
	b.dispatchLock.Lock()
	if !b.dispatchStarted {
		b.dispatchScheduler.Start()
		b.dispatchStarted = true
	}
	b.dispatchLock.Unlock()
}

// dispatchOn dispatches an action to the store on the scheduler.
func (b *baseStore[S]) dispatchOn(scheduler sched.Scheduler, action Action) {
	if b == nil {
		return
	}
	switch action.(type) {
	case AsyncAction:
		scheduler.Schedule(func() {
			asyncAction := action.(AsyncAction)
			asyncAction(b)
		})
	default:
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
}

func (b *baseStore[S]) Subscribe(subscriber Subscriber[S]) Disposer {
	if b == nil {
		return nil
	}

	b.ensureScheduler()

	return b.SubscribeOn(b.dispatchScheduler, subscriber)
}

func (b *baseStore[S]) SubscribeOn(scheduler sched.Scheduler, subscriber Subscriber[S]) Disposer {
	if b == nil {
		return nil
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

	return &baseDisposer{
		dispose: func() {
			b.dispatchLock.Lock()
			for idx := 0; idx < len(b.subscribers); idx++ {
				if &entry == b.subscribers[idx] {
					b.subscribers = append(b.subscribers[:idx], b.subscribers[idx+1:]...)
					break
				}
			}
			b.dispatchLock.Unlock()
			// need notify?
		},
	}
}

func (b *baseStore[S]) getState() (state S) {
	if b == nil {
		return
	}
	return b.state
}

func (b *baseStore[S]) waitForDispatch() {
	b.dispatchLock.Lock()
	if b.dispatchStarted {
		b.dispatchScheduler.Stop()
	}
	b.dispatchLock.Unlock()

	b.dispatchScheduler.WaitForScheduler()
}

// reduce should be called in the same(Main) context
func (b *baseStore[S]) reduce(state S, action Action) S {
	if b == nil {
		return state
	}
	reducers := b.reducers[:]
	newState := b.state
	var err error
	for _, reducer := range reducers {
		newState, err = reducer(newState, action)
		if err != nil {
			if errors.Is(err, ErrSkipReducing) {
				break
			}
			logger.Errf("error reducing: %s", err)
		}
	}
	return newState
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

	// if the subscriber is called in the same context, it will be called immediately
	// not to make deadlock on dispatcher
	if entry.scheduler == b.dispatchScheduler {
		entry.subscriber(newState, oldState, action)
	} else {
		if wg != nil {
			wg.Add(1)
		}

		entry.scheduler.Schedule(func() {

			// call subscriber
			entry.subscriber(newState, oldState, action)

			// 'Done' called after calling a subscriber to ensure all subscribers are one same state
			// wake up Dispatcher
			if wg != nil {
				wg.Done()
			}
		})
	}
}

func (b *baseStore[S]) onFirstSubscribe() {

}
