package store

import (
	"log"
	"testing"
)

func Test_baseStore_Subscribe(t *testing.T) {

	type args[S State] struct {
		action      Action
		actions     int // actions to dispatch
		subscribers int // subscribers to add
	}
	type testCase[S State] struct {
		name   string
		b      Store[S]
		args   args[S]
		called int // 호출횟수 count
		want   int // call times
	}
	tests := []testCase[myState]{
		{
			name: "action 0 - subscriber 1 - callback when subscribe",
			b:    newMyStateStore(),
			args: args[myState]{
				action:      nil,
				subscribers: 1,
			},
			want: 1,
		},
		{
			name: "action 0 - subscriber 1",
			b:    newMyStateStore(),
			args: args[myState]{
				action:      nil,
				subscribers: 1,
			},
			want: 1,
		},
		{
			name: "action 1 - subscriber 1",
			b:    newMyStateStore(),
			args: args[myState]{
				action:      &addAction{},
				actions:     1,
				subscribers: 1,
			},
			want: 2,
		},
		{
			name: "action 1 - subscriber 2",
			b:    newMyStateStore(),
			args: args[myState]{
				action:      &addAction{},
				actions:     1,
				subscribers: 2,
			},
			want: 4,
		},
		{
			name: "action 2 - subscriber 2",
			b:    newMyStateStore(),
			args: args[myState]{
				action:      &addAction{},
				actions:     2,
				subscribers: 2,
			},
			want: 2 + 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			for idx := 0; idx < tt.args.subscribers; idx++ {
				log.Println("Subscribe", idx)
				tt.b.Subscribe(func(state myState, old myState, action Action) {
					tt.called++
				})
			}
			log.Println("Dispatch action", tt.args.action)
			if tt.args.action != nil {
				for idx := 0; idx < tt.args.actions; idx++ {
					tt.b.Dispatch(tt.args.action)
				}
			}

			tt.b.waitForDispatch()

			if tt.want != tt.called {
				t.Errorf("Subscribe: want %d, got %d, state %v, action %v", tt.want, tt.called, tt.b.getState(), tt.args.action)
			}
		})
	}
}
