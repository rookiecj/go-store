package store

import (
	"fmt"
	"github.com/rookiecj/go-store/logger"
	"reflect"
	"strings"
	"testing"
)

func Test_Store_example(t *testing.T) {

	t.Run("example", func(t *testing.T) {
		initialState := myState{}
		reducer := func(state myState, action Action) (myState, error) {
			switch reified := action.(type) {
			case *addAction:
				return myState{
					id:    state.id,
					value: state.value + reified.value,
				}, nil
			}
			return initialState, nil
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

		store.Stop()
		store.WaitForStore()
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
			//if wantRaw.reduceScheduler != Main || gotRaw.reduceScheduler != Main {
			//	t.Errorf("NewStore() = %v, want %v", got, tt.want)
			//}
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

func Test_baseStore_getState(t *testing.T) {
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

			if got := tt.b.getState(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getState() = %v, want %v", got, tt.want)
			}

			tt.b.Stop()
			tt.b.WaitForStore()
		})
	}
}

func Test_baseStore_Dispatch(t *testing.T) {

	type args struct {
		actions []Action
	}
	// debugger got confused to get the right symbol, change name
	type testCaseDispatch[S State] struct {
		name string
		b    Store[S]
		args args
		want S
	}
	tests := []testCaseDispatch[myState]{
		{
			name: "nil actions - no dispatch",
			b:    newMyStateStore(),
			args: args{
				actions: nil,
			},
			want: myInitialState,
		},
		{
			name: "add action - empty - no dispatch",
			b:    newMyStateStore(),
			args: args{
				actions: []Action{},
			},
			want: myInitialState,
		},
		{
			name: "add action - nil",
			b:    newMyStateStore(),
			args: args{
				actions: []Action{nil},
			},
			want: myInitialState,
		},

		{
			name: "add action - 123",
			b:    newMyStateStore(),
			args: args{
				actions: []Action{
					&addAction{"123"},
				},
			},
			want: myState{
				id:    0,
				value: "123",
			},
		},
		{
			name: "add action - 123 - 456",
			b:    newMyStateStore(),
			args: args{
				actions: []Action{
					&addAction{"123"},
					&addAction{"456"},
				},
			},
			want: myState{
				id:    0,
				value: "123456",
			},
		},
	}

	logger.SetLogEnable(true)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			for _, action := range tt.args.actions {
				tt.b.Dispatch(action)
			}

			tt.b.Stop()
			tt.b.WaitForStore()

			want := tt.want
			got := tt.b.getState()
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Dispatch: want %v got %v, actions %v", want, got, tt.args.actions)
			}
		})
	}
}

func Test_baseStore_ReduceSerialized(t *testing.T) {

	type args struct {
		times int
	}
	type testCaseReduceSerial[S State] struct {
		name string
		b    Store[S]
		args args
		want S
	}

	limit := 1024

	tests := []testCaseReduceSerial[myState]{
		{
			name: "add action - x times",
			b:    newMyStateStore(),
			args: args{
				times: limit,
			},
			want: myState{
				id: 0,
				value: func() (result string) {
					for idx := 0; idx < limit; idx++ {
						if idx == 0 {
							result += fmt.Sprintf("%d", idx+1)
						} else {
							result += fmt.Sprintf(",%d", idx+1)
						}
					}
					return
				}(),
			},
		},
	}

	logger.SetLogEnable(true)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			for idx := 0; idx < tt.args.times; idx++ {
				nth := idx + 1
				if nth == 1 {
					tt.b.Dispatch(&addAction{
						value: fmt.Sprintf("%d", nth),
					})
				} else {
					tt.b.Dispatch(&addAction{
						value: fmt.Sprintf(",%d", nth),
					})
				}
			}

			tt.b.Stop()
			tt.b.WaitForStore()

			want := tt.want
			got := tt.b.getState()
			wantToks := strings.Split(want.value, ",")
			gotToks := strings.Split(got.value, ",")
			if len(wantToks) != len(gotToks) {
				t.Errorf("ReduceSerialized: want %v", want)
				t.Errorf("ReduceSerialized: got %v", got)
			}
			for idx := 0; idx < len(wantToks); idx++ {
				if wantToks[idx] != gotToks[idx] {
					t.Errorf("ReduceSerialized: %d: '%s','%s'", idx, wantToks[idx], gotToks[idx])
				}
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("ReduceSerialized: want %v", want)
				t.Errorf("ReduceSerialized: got %v", got)
			}
		})
	}
}
