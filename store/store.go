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

	SubscribeOn(scheduler Scheduler, subscriber Subscriber[S]) Store[S]
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
	subscribers []subscriberEntry[S]
	//lock         sync.RWMutex

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
	return b.SubscribeOn(Caller, subscriber)
}

func (b *baseStore[S]) SubscribeOn(scheduler Scheduler, subscriber Subscriber[S]) Store[S] {
	if b == nil {
		return b
	}

	if len(b.subscribers) == 0 {
		// onFirstSubscriber
		b.onBeginSubscribe()
	}

	b.subscribers = append(b.subscribers, subscriberEntry[S]{
		scheduler:  scheduler,
		subscriber: subscriber})

	// schedule the task on caller's context
	b.dispatchLock.Lock()
	b.doDispatchOn(scheduler, nil, b.age, subscriber, b.state, b.state, InitAction)
	b.dispatchLock.Unlock()
	return b
}

func (b *baseStore[S]) reduce(state S, action Action) S {
	if b == nil {
		return state
	}

	// reduce state on Main context
	//return b.reducer(oldState, action)
	return b.doReduceOn(Main, b.age, b.reducer, state, action)
}

func (b *baseStore[S]) doReduceOn(scheduler Scheduler, age int64, reducer Reducer[S], state S, action Action) (newState S) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	reduceTask := NewReduceTask(age, reducer, state, action)
	scheduler.Schedule(reduceTask, func() {
		wg.Done()
	})
	wg.Wait()
	newState = reduceTask.Result().(S)
	return
}

func (b *baseStore[S]) dispatch(oldState S, action Action, newState S) error {
	if b == nil {
		return errors.New("store is nil")
	}

	// dispatch state in subscriber's context
	b.dispatchLock.Lock()
	if len(b.subscribers) > 0 {
		clonedSubscribers := b.subscribers[:]
		age := atomic.AddInt64(&b.age, 1)
		wg := sync.WaitGroup{}
		for _, entry := range clonedSubscribers {
			b.doDispatchOn(entry.scheduler, &wg, age, entry.subscriber, newState, oldState, action)
		}
		wg.Wait()
	}
	b.dispatchLock.Unlock()
	return nil
}

func (b *baseStore[S]) doDispatchOn(scheduler Scheduler, wg *sync.WaitGroup, age int64, subscriber Subscriber[S], newState S, oldState S, action Action) {
	if wg != nil {
		wg.Add(1)
	}
	scheduler.Schedule(NewDispatchTask(age, subscriber, newState, oldState, action), func() {
		if wg != nil {
			wg.Done()
		}
	})
}

func (b *baseStore[State]) onBeginSubscribe() {

}
