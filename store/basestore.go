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
	age               int64
	dispatchLock      *sync.Mutex
	stopWg            *sync.WaitGroup
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
// scheduler should be started/stopped properly before/after using Store
func NewStoreOn[S State](scheduler sched.Scheduler, initialState S, reducer Reducer[S]) Store[S] {
	return &baseStore[S]{
		state:             initialState,
		reducers:          []Reducer[S]{reducer},
		dispatchScheduler: scheduler,
		age:               0,
		dispatchLock:      &sync.Mutex{},
		stopWg:            &sync.WaitGroup{},
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

	// reduce state in dispatcher context
	b.dispatchOn(b.dispatchScheduler, action)
}

// dispatchOn dispatches an action to the store on the scheduler.
func (b *baseStore[S]) dispatchOn(scheduler sched.Scheduler, action Action) {
	if b == nil {
		return
	}
	switch reified := action.(type) {
	case AsyncAction:
		scheduler.Schedule(func() {
			reified(b)
		})
	default:
		scheduler.Schedule(func() {
			// reduce
			oldState := b.getState()
			//logger.Debugf("Store: reduce: action:%v\n", action)
			b.state = b.reduce(oldState, action)
			// dispatch
			//logger.Debugf("Store: dispatch: action:%v, state: %v\n", action, b.state)
			b.dispatch(oldState, action, b.state)
		})
	}
}

func (b *baseStore[S]) Subscribe(subscriber Subscriber[S]) Disposer {
	if b == nil {
		return nil
	}

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
	b.dispatchWhenSubscribe(&entry, b.state, b.state, &InitAction{})

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

func (b *baseStore[S]) Stop() {
	if b.dispatchScheduler != sched.Main {
		b.dispatchScheduler.Stop()
	}
}

func (b *baseStore[S]) WaitForStore() {
	if b == nil {
		return
	}

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
			logger.Errf("error reducing: %s\n", err)
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

		// for a subscriber with its own scheduler
		wg := &sync.WaitGroup{}
		// dispatch state in subscriber's context
		for _, entry := range clonedSubscribers {
			b.doDispatchSubscriberLocked(entry, wg, age, newState, oldState, action)
		}
		// wait for subscribers scheduler to done
		wg.Wait()
	}
	b.dispatchLock.Unlock()
}

func (b *baseStore[S]) dispatchWhenSubscribe(entry *subscriberEntry[S], newState S, oldState S, action Action) {
	if b == nil {
		return
	}

	b.dispatchLock.Lock()
	b.dispatchScheduler.Schedule(func() {
		wg := sync.WaitGroup{}
		b.doDispatchSubscriberLocked(entry, &wg, b.age, newState, oldState, action)
		wg.Wait()
	})
	b.dispatchLock.Unlock()

}

func (b *baseStore[S]) doDispatchSubscriberLocked(entry *subscriberEntry[S], wg *sync.WaitGroup, age int64, newState S, oldState S, action Action) {

	// we are in the dispatcher context, so we can call subscriber directly
	if entry.scheduler == b.dispatchScheduler {
		//logger.Debugf("Store: doDispatchSubscriberLocked: schedule in same scheduler with action %v\n", action)
		entry.subscriber(newState, oldState, action)
		return
	}

	// if scheduler has it own scheduler, the dispatcher should wait for it to done
	if entry.scheduler != b.dispatchScheduler {
		if wg != nil {
			wg.Add(1)
		}

		//logger.Debugf("Store: doDispatchSubscriberLocked: schedule subscriber with action %v\n", action)
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
