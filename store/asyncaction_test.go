package store

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

type testAsyncAction struct {
	work int
	run  func(work int, dispatcher Dispatcher)
}

func (c *testAsyncAction) ActionInterface() {}
func (c *testAsyncAction) Run(dispatcher Dispatcher) {
	// do async work
	c.run(c.work, dispatcher)
}

func Test_AsyncAction_Run(t *testing.T) {

	type args struct {
		action AsyncAction
	}
	type testCaseAsyncActionRun[S State] struct {
		name string
		b    Store[S]
		args args
		want int
	}

	limit := 100 // workload

	tests := []testCaseAsyncActionRun[myState]{
		{
			name: "add testAsyncAction - delay 100ms",
			b:    newMyStateStore(),
			args: args{
				action: &testAsyncAction{
					work: limit,
					run: func(work int, dispatcher Dispatcher) {

						// do async work
						delay := rand.Intn(work)
						time.Sleep(time.Duration(delay) * time.Millisecond)

						// dispatch result
						dispatcher.Dispatch(&setAction{
							value: fmt.Sprintf("%d", delay),
						})
					},
				},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testScheduler.Start()

			called := 0
			tt.b.Subscribe(func(newState myState, oldState myState, action Action) {
				switch action.(type) {
				case *setAction:
					called++
				}
			})

			//start := time.Now()
			tt.b.Dispatch(tt.args.action)

			// give time to async action
			time.Sleep(time.Duration(limit) * time.Millisecond)

			testScheduler.Stop()
			tt.b.waitForDispatch()

			//diff := time.Now().Sub(start).Milliseconds()
			//if diff >= tt.want {
			//t.Errorf("AsyncAction_Run: want %d, got %d", tt.want, diff)
			//}

			if called != tt.want {
				t.Errorf("AsyncAction_Run: done action call want %d times but %d", tt.want, called)
			}
		})
	}
}
