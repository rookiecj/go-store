package store

import (
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
	testScheduler  = Main
	myInitialState = myState{
		id:    0,
		value: "",
	}
)

func (c myState) StateInterface()     {}
func (c *addAction) ActionInterface() {}
func (c *setAction) ActionInterface() {}

func newMyStateStore() Store[myState] {
	return newMyStateStoreWithReducer(myStateReducer)
}

func newMyStateStoreWithReducer(reducer Reducer[myState]) Store[myState] {
	// test on Immediate scheduler
	return NewStoreOn(testScheduler, myInitialState, reducer)
	//return NewStore(myInitialState, reducer)
}

func myStateReducer(state myState, action Action) myState {
	// support nil action
	if action == nil {
		return state
	}
	switch action.(type) {
	case *addAction:
		reifiedAction := action.(*addAction)
		return myState{
			id:    state.id,
			value: state.value + reifiedAction.value,
		}
	case *setAction:
		reifiedAction := action.(*setAction)
		return myState{
			id:    state.id,
			value: reifiedAction.value,
		}
	}
	return state
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