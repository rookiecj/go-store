package store

import (
	"fmt"
	"github.com/rookiecj/go-store/logger"
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
		actions []AsyncAction
	}
	type testCaseAsyncActionRun[S State] struct {
		name      string
		b         Store[S]
		args      args
		want      int
		wantState string
	}

	//limit := 100 // workload

	tests := []testCaseAsyncActionRun[myState]{
		{
			name: "add testAsyncAction - delay 100ms",
			b:    newMyStateStore(),
			args: args{
				actions: []AsyncAction{
					&testAsyncAction{
						work: 100, // 100 millis
						run: func(work int, dispatcher Dispatcher) {

							// do async work
							go func() {
								delay := rand.Intn(work)
								time.Sleep(time.Duration(delay) * time.Millisecond)

								// dispatch result
								dispatcher.Dispatch(&setAction{
									value: fmt.Sprintf("%d", work),
								})
							}()
						},
					},
				},
			},
			want:      1,
			wantState: "100",
		},
		{
			name: "two async action - got later state",
			b:    newMyStateStore(),
			args: args{
				actions: []AsyncAction{
					&testAsyncAction{
						work: 100, // 100 millis
						run: func(work int, dispatcher Dispatcher) {

							// do async work
							go func() {
								logger.Infof("first async job %d", work)
								time.Sleep(time.Duration(work) * time.Millisecond)

								// dispatch result
								dispatcher.Dispatch(&setAction{
									value: fmt.Sprintf("first job %d", work),
								})
							}()
						},
					},
					&testAsyncAction{
						work: 50, // 50 millis
						run: func(work int, dispatcher Dispatcher) {

							// do async work
							go func() {
								logger.Infof("second async job %d", work)
								time.Sleep(time.Duration(work) * time.Millisecond)

								// dispatch result
								dispatcher.Dispatch(&setAction{
									value: fmt.Sprintf("second job %d", work),
								})
							}()
						},
					},
				},
			},
			want:      2,
			wantState: "first job 100",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			setActionCalled := 0
			tt.b.Subscribe(func(newState myState, oldState myState, action Action) {
				switch action.(type) {
				case *setAction:
					setActionCalled++
				}
			})

			//start := time.Now()
			delay := 0
			for _, action := range tt.args.actions {
				if delay < action.(*testAsyncAction).work {
					delay = action.(*testAsyncAction).work
				}
				tt.b.Dispatch(action)
			}

			// give time to async action
			time.Sleep(time.Duration(delay) * time.Millisecond)

			tt.b.waitForDispatch()

			//diff := time.Now().Sub(start).Milliseconds()
			//if diff >= tt.want {
			//t.Errorf("AsyncAction_Run: want %d, got %d", tt.want, diff)
			//}

			if setActionCalled != tt.want {
				t.Errorf("AsyncAction_Run: done action call want %d times but %d", tt.want, setActionCalled)
			}

			gotState := tt.b.getState()
			if tt.wantState != gotState.value {
				t.Errorf("AsyncAction_Run: state want %s got %s", tt.wantState, gotState.value)
			}
		})
	}
}
