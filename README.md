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

type myState struct {
    id    int
    value string
}

type addAction struct {
    value string
}

initialState := myState{}
reducer := func(state myState, action Action) myState {
    switch action.(type) {
    case *addAction:
        reifiedAction := action.(*addAction)
        return myState{
            id:    state.id,
            value: state.value + reifiedAction.value,
        }
    }
    return initialState
}

store := NewStore[myState](initialState, reducer)

store.Subscribe(func(newState myState, oldState myState, action Action) {
    fmt.Println("subscriber1", newState)
})

store.Subscribe(func(newState myState, oldState myState, action Action) {
    fmt.Println("subscriber2", newState)
})

store.Dispatch(&addAction{
    value: "1",
})
store.Dispatch(&addAction{
    value: "2",
})
store.Dispatch(&addAction{
    value: "3",
})
```

## TODO
- [X] make sure all subscribers notified
- [X] add Store callbacks like OnFirstSubscription
- [X] add Dispatch Scheduler, SubscribeOn
- [ ] add README
- [X] add doc
- [ ] add more testing