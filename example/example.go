package main

import (
	"fmt"
	"time"

	"github.com/rookiecj/go-store/store"
)

type myState struct {
	id    int
	value string
}

type addAction struct {
	value string
}

func (c myState) StateInterface() {}

func main() {

	initialState := myState{}
	reducer := func(state myState, action store.Action) (myState, error) {
		switch action.(type) {
		case *addAction:
			reifiedAction := action.(*addAction)
			return myState{
				id:    state.id,
				value: state.value + reifiedAction.value,
			}, nil
		}
		return state, nil
	}

	stateStore := store.NewStore[myState](initialState, reducer)

	stateStore.Subscribe(func(newState myState, oldState myState, action store.Action) {
		fmt.Println("subscriber1", newState)
	})

	stateStore.Subscribe(func(newState myState, oldState myState, action store.Action) {
		fmt.Println("subscriber2", newState)
	})

	stateStore.Dispatch(&addAction{
		value: "1",
	})
	stateStore.Dispatch(&addAction{
		value: "2",
	})
	stateStore.Dispatch(&addAction{
		value: "3",
	})

	// store.waitForDispatch()
	time.Sleep(100 * time.Millisecond)
}
