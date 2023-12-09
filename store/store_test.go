package store

import (
	"fmt"
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
	myInitialState = myState{}
)

func (c myState) stateInterface()     {}
func (c *addAction) actionInterface() {}
func (c *setAction) actionInterface() {}

func newMyStateStore() Store[myState] {
	return newMyStateStoreWithReducer(myStateReducer)
}
func newMyStateStoreWithReducer(reducer Reducer[myState]) Store[myState] {
	return NewStore(myInitialState, reducer)
}

func myStateReducer(state myState, action Action) myState {
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

func Test_Store_example(t *testing.T) {

	t.Run("example", func(t *testing.T) {
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
			fmt.Println(newState)
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
	})
}

func Test_NewStore(t *testing.T) {
	type args[S State] struct {
		state   S
		reducer Reducer[S]
	}
	type testCase[S State] struct {
		name string
		args args[S]
		want Store[S]
	}
	tests := []testCase[myState]{
		{
			name: "new store",
			args: args[myState]{
				state:   myInitialState,
				reducer: myStateReducer,
			},
			want: NewStore(myInitialState, myStateReducer),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := tt.want
			got := NewStore(tt.args.state, tt.args.reducer)
			if got == nil {
				t.Errorf("NewStore() = %v, want %v", got, tt.want)
			}
			wantRaw := want.(*baseStore[myState])
			gotRaw := got.(*baseStore[myState])
			if wantRaw == nil || gotRaw == nil {
				t.Errorf("NewStore() = %v, want %v", got, tt.want)
			}

			if !reflect.DeepEqual(wantRaw.state, gotRaw.state) {
				t.Errorf("NewStore() = %v, want %v", got, tt.want)
			}

			// 함수 비교는 false
			//if !reflect.DeepEqual(wantRaw.reducer, gotRaw.reducer) {
			//	t.Errorf("NewStore() = %v, want %v", got, tt.want)
			//}

			if !reflect.DeepEqual(wantRaw.subscribers, gotRaw.subscribers) {
				t.Errorf("NewStore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_baseStore_GetState(t *testing.T) {
	type testCase[S State] struct {
		name string
		b    Store[S]
		want State
	}
	tests := []testCase[myState]{
		{
			name: "state - initial",
			b:    newMyStateStore(),
			want: myInitialState,
		},
		{
			name: "state - state",
			b: NewStore(myState{
				id:    1,
				value: "1",
			}, myStateReducer),
			want: myState{
				id:    1,
				value: "1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.GetState(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_baseStore_Dispatch(t *testing.T) {

	type args struct {
		action Action
	}
	type testCase[S State] struct {
		name string
		b    Store[S]
		args args
		want S
	}
	tests := []testCase[myState]{
		{
			name: "nil action",
			b:    newMyStateStore(),
			args: args{
				action: nil,
			},
			want: myInitialState,
		},
		{
			name: "add action - empty",
			b:    newMyStateStore(),
			args: args{
				action: &addAction{},
			},
			want: myInitialState,
		},
		{
			name: "add action - 123",
			b:    newMyStateStore(),
			args: args{
				action: &addAction{
					value: "123",
				},
			},
			want: myState{
				id:    0,
				value: "123",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.b.Dispatch(tt.args.action)
			want := tt.want
			got := tt.b.GetState()
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Dispatch: want %v got %v, action %v", want, got, tt.args.action)
			}
		})
	}
}
