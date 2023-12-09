# go-store

`go-store` is a state holder in Redux pattern.
The State is read-only, the Changes are made with reducer in uni-directional way

## The Redux pattern

`Store` holds a state.
`Action` is delivered to the store to change the state.
`Reducer` changes the state with Action.
Store delivers the change to `Subscriber`


## How To Use

it can be installed by as follows:

```sh
go get github.com/rookiecj/go-store
```


## TODO
- [ ] make sure all subscribers notified
- [ ] add Store callbacks like OnFirstSubscription
- [ ] add Dispatch Scheduler, SubscribeOn
- [ ] add README
- [ ] add doc
- [ ] add more testing