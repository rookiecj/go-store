package store

import (
	"github.com/rookiecj/go-store/sched"
	"reflect"
	"testing"
)

type myState struct {
	id    int
	value string
}

type addAction struct {
	value string
}

type setAction struct {
	value string
}

var (
	myInitialState = myState{
		id:    0,
		value: "",
	}
)

func (c myState) StateInterface() {}

// new store with testScheduler and myStateReducer
func newMyStateStore() Store[myState] {
	testScheduler := sched.NewMainScheduler()
	return NewStoreOn(testScheduler, myInitialState, myStateReducer)
}

func myStateReducer(state myState, action Action) (myState, error) {
	// support nil action
	if action == nil {
		return state, nil
	}
	switch reified := action.(type) {
	case *addAction:
		return myState{
			id:    state.id,
			value: state.value + reified.value,
		}, nil
	case *setAction:
		reifiedAction := action.(*setAction)
		return myState{
			id:    state.id,
			value: reifiedAction.value,
		}, nil
	}
	return state, nil
}

func getTestSubscriber[S State](t *testing.T, inner func(t *testing.T, state S, old S, action Action)) Subscriber[S] {
	testSubscriber := func(state S, old S, action Action) {
		inner(t, state, old, action)
	}
	return testSubscriber
}

func assertState(t *testing.T, got myState, want myState, action Action) {
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Dispatch() state got %v want %v, action = %v", got, want, action)
	}
}
