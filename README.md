# go-store

`go-store` is a state holder in Redux pattern.
The State is read-only, the Changes are made with reducer in uni-directional way

## The Redux pattern

`Store` holds a state.
`Action` is delivered to the store to change the state.
`Reducer` changes the state with Action.
Store delivers the change to `Subscriber`


## How To Use

It can be installed by:
```sh
go get github.com/rookiecj/go-store
```

how to use as follows: 
```go

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

func (c myState) StateInterface()     {}

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

```

## TODO
- [X] make sure all subscribers notified
- [X] add Store callbacks like onFirstSubscribe
- [X] add SubscribeOn
- [X] add README
- [X] add doc
- [X] support Main/Background/Dispatch Scheduler(Experimental)
- [X] remove explicit scheduler start/stop
- [X] add AsyncAction for async work
- [X] add AddReducer
- [ ] make getState public for subscribers not to save the state locally 
- [ ] add State history
- [ ] make age precisely
- [ ] add more testing
