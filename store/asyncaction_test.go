package store

import (
	"fmt"
	"github.com/rookiecj/go-store/logger"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"
)

type testAsyncArgs struct {
	work   int
	action AsyncAction
}

func Test_AsyncAction_Run(t *testing.T) {

	type args struct {
		asyncArgs []testAsyncArgs
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
			name: "add testAsyncArgs - delay 100ms",
			b:    newMyStateStore(),
			args: args{
				asyncArgs: []testAsyncArgs{
					{
						work: 100,
						action: func(dispatcher Dispatcher) {
							// do async work
							// and dispatch event
							work := 100
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
				asyncArgs: []testAsyncArgs{
					{
						work: 100,
						action: func(dispatcher Dispatcher) {
							// do async work
							// and dispatch event
							work := 100
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

					{
						work: 50,
						action: func(dispatcher Dispatcher) {
							// do async work
							// and dispatch event
							work := 50
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

			var setActionCalled int64

			tt.b.Subscribe(func(newState myState, oldState myState, action Action) {
				switch action.(type) {
				case *setAction:
					atomic.AddInt64(&setActionCalled, 1)
				}
			})

			//start := time.Now()
			delay := 0
			for _, asyncArg := range tt.args.asyncArgs {
				if delay < asyncArg.work {
					delay = asyncArg.work
				}
				tt.b.Dispatch(asyncArg.action)
			}

			// give enough(*2) time to async action
			time.Sleep(time.Duration(delay) * 2 * time.Millisecond)

			tt.b.Stop()
			tt.b.WaitForStore()

			//diff := time.Now().Sub(start).Milliseconds()
			//if diff >= tt.want {
			//t.Errorf("AsyncAction_Run: want %d, got %d", tt.want, diff)
			//}

			if setActionCalled != int64(tt.want) {
				t.Errorf("AsyncAction_Run: done action call want %d times but %d", tt.want, setActionCalled)
			}

			gotState := tt.b.getState()
			if tt.wantState != gotState.value {
				t.Errorf("AsyncAction_Run: state want %s got %s", tt.wantState, gotState.value)
			}
		})
	}
}
