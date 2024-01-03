package store

import (
	"reflect"
	"testing"
)

type setAndSkipAction struct {
	value string
	skip  bool
}

func (c *setAndSkipAction) ActionInterface() {}

func Test_baseStore_AddReducer(t *testing.T) {
	type args[S State] struct {
		reducers []Reducer[S]
		actions  []Action
	}
	type addReducerTestCase[S State] struct {
		name string
		b    Store[S]
		args args[S]
		want string
	}

	myStateFirstSetReducer := func(state myState, action Action) (myState, error) {
		// support nil action
		if action == nil {
			return state, nil
		}
		switch action.(type) {
		case *setAndSkipAction:
			reifiedAction := action.(*setAndSkipAction)
			var err error
			if reifiedAction.skip {
				err = ErrSkipReducing
			}
			return myState{
				id:    state.id,
				value: "first: " + reifiedAction.value,
			}, err
		}
		return state, nil
	}

	myStateSecondSetReducer := func(state myState, action Action) (myState, error) {
		// support nil action
		if action == nil {
			return state, nil
		}
		switch action.(type) {
		case *setAction:
			reifiedAction := action.(*setAction)
			return myState{
				id:    state.id,
				value: "second: " + reifiedAction.value,
			}, nil
		}
		return state, nil
	}

	tests := []addReducerTestCase[myState]{
		{
			name: "add zero",
			b:    newMyStateStore(), // with default reducer
			args: args[myState]{
				reducers: []Reducer[myState]{},
				actions: []Action{
					&setAction{
						value: "0",
					},
				},
			},
			want: "0",
		},
		{
			name: "add two more - no skip",
			b:    newMyStateStore(), // with default reducer
			args: args[myState]{
				reducers: []Reducer[myState]{
					myStateFirstSetReducer,
					myStateSecondSetReducer,
				},
				actions: []Action{
					&setAction{
						value: "1",
					},
				},
			},
			want: "second: 1",
		},
		{
			name: "add two more - skip at first",
			b:    newMyStateStore(), // with default reducer
			args: args[myState]{
				reducers: []Reducer[myState]{
					myStateFirstSetReducer,
					myStateSecondSetReducer,
				},
				actions: []Action{
					&setAction{
						value: "1",
					},
					&setAndSkipAction{
						value: "2",
						skip:  true,
					},
				},
			},
			want: "first: 2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, reducer := range tt.args.reducers {
				tt.b.AddReducer(reducer)
			}

			for _, action := range tt.args.actions {
				tt.b.Dispatch(action)
			}

			tt.b.waitForDispatch()

			got := tt.b.getState()
			if !reflect.DeepEqual(tt.want, got.value) {
				t.Errorf("AddReducer(): got %v, want %v", got, tt.want)
			}
		})
	}
}
