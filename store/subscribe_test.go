package store

import (
	"fmt"
	"github.com/rookiecj/go-store/logger"
	"github.com/rookiecj/go-store/sched"
	"log"
	"sync/atomic"
	"testing"
)

func Test_baseStore_Subscribe(t *testing.T) {

	type args[S State] struct {
		action      Action
		actions     int // actions to dispatch
		subscribers int // subscribers to add
	}
	type testCaseSubscribe[S State] struct {
		name   string
		b      Store[S]
		args   args[S]
		called int // temporal variable for a testcase
		want   int // call times
	}
	tests := []testCaseSubscribe[myState]{
		{
			name: "action 0 - subscriber 1 - callback when subscribe",
			b:    newMyStateStore(),
			args: args[myState]{
				action:      nil,
				actions:     0,
				subscribers: 1,
			},
			called: 0,
			want:   1,
		},
		{
			name: "action 0 - subscriber 1",
			b:    newMyStateStore(),
			args: args[myState]{
				action:      nil,
				actions:     0,
				subscribers: 1,
			},
			called: 0,
			want:   1,
		},
		{
			name: "action 1 - subscriber 1",
			b:    newMyStateStore(),
			args: args[myState]{
				action:      &addAction{"1"},
				actions:     1,
				subscribers: 1,
			},
			called: 0,
			want:   2,
		},
		{
			name: "action 1 - subscriber 2",
			b:    newMyStateStore(),
			args: args[myState]{
				action:      &addAction{"12"},
				actions:     1,
				subscribers: 2,
			},
			called: 0,
			want:   4,
		},
		{
			name: "action 2 - subscriber 2",
			b:    newMyStateStore(),
			args: args[myState]{
				action:      &addAction{"22"},
				actions:     2,
				subscribers: 2,
			},
			called: 0,
			want:   2 + 4,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			log.Println("Subscriber: subscribers:", tt.args.subscribers)
			for idx := 0; idx < tt.args.subscribers; idx++ {
				idxdup := idx
				tt.b.Subscribe(func(state myState, old myState, action Action) {
					tt.called++
					log.Printf("Subscriber %d: got called: %d state:%v\n", idxdup, tt.called, state)
				})
			}

			log.Printf("Dispatch: actions: %d, action=%v", tt.args.actions, tt.args.action)
			for idx := 0; idx < tt.args.actions; idx++ {
				log.Printf("Dispatch: idx:%d action", idx)
				tt.b.Dispatch(tt.args.action)
			}

			tt.b.Stop()
			tt.b.WaitForStore()

			if tt.want != tt.called {
				t.Errorf("Subscribe: want %d, got %d, state %v, action %v", tt.want, tt.called, tt.b.getState(), tt.args.action)
			}
		})
	}
}

func Test_baseStore_SubscribeOn(t *testing.T) {

	type args[S State] struct {
		// subscriber scheduler
		scheduler   sched.Scheduler
		action      Action
		actions     int64 // actions to dispatch
		subscribers int64 // subscribers to add
	}
	type testCaseSubscribeOn[S State] struct {
		name   string
		b      Store[S]
		args   args[S]
		called int64 // temporal variable for a testcase
		want   int64 // call times
	}

	var actionLimit int64 = 16
	var subscriberlimit int64 = 64
	subscriberScheduler := sched.Background

	tests := []testCaseSubscribeOn[myState]{
		{
			name: "no scheduler - action 0 - subscriber 1 - callback when subscribe",
			b:    newMyStateStore(),
			args: args[myState]{
				scheduler:   nil,
				action:      nil,
				actions:     0,
				subscribers: 1,
			},
			want:   1,
			called: 0,
		},
		{
			name: "background - action 0 - subscriber 1 - callback when subscribe",
			b:    newMyStateStore(),
			args: args[myState]{
				scheduler:   subscriberScheduler,
				action:      nil,
				actions:     0,
				subscribers: 1,
			},
			called: 0,
			want:   1,
		},
		{
			name: "background - action 0 - subscriber 1",
			b:    newMyStateStore(),
			args: args[myState]{
				scheduler:   subscriberScheduler,
				action:      nil,
				actions:     0,
				subscribers: 1,
			},
			called: 0,
			want:   1,
		},
		{
			name: "background - action 1 - subscriber 1",
			b:    newMyStateStore(),
			args: args[myState]{
				scheduler:   subscriberScheduler,
				action:      &addAction{"1"},
				actions:     1,
				subscribers: 1,
			},
			called: 0,
			want:   2,
		},
		{
			name: "background - action 1 - subscriber 2",
			b:    newMyStateStore(),
			args: args[myState]{
				scheduler:   subscriberScheduler,
				action:      &addAction{"12"},
				actions:     1,
				subscribers: 2,
			},
			called: 0,
			want:   4,
		},
		{
			name: "background - action 2 - subscriber 2",
			b:    newMyStateStore(),
			args: args[myState]{
				scheduler:   subscriberScheduler,
				action:      &addAction{"22"},
				actions:     2,
				subscribers: 2,
			},
			called: 0,
			want:   2 + 4,
		},

		{
			name: "background - action 2 - subscriber many",
			b:    newMyStateStore(),
			args: args[myState]{
				scheduler:   subscriberScheduler,
				action:      &addAction{"2x"},
				actions:     2,
				subscribers: subscriberlimit,
			},
			called: 0,
			want:   subscriberlimit + 2*subscriberlimit,
		},

		{
			name: "background - action many - subscriber 2",
			b:    newMyStateStore(),
			args: args[myState]{
				scheduler:   subscriberScheduler,
				action:      &addAction{"X2"},
				actions:     actionLimit,
				subscribers: 2,
			},
			called: 0,
			want:   2 + actionLimit*2,
		},
		{
			name: "background - action many - subscriber many",
			b:    newMyStateStore(),
			args: args[myState]{
				scheduler:   subscriberScheduler,
				action:      nil,
				actions:     actionLimit,
				subscribers: subscriberlimit,
			},
			called: 0,
			want:   subscriberlimit + actionLimit*subscriberlimit,
		},
	}

	logger.SetLogEnable(true)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			log.Println("Subscriber: subscribers:", tt.args.subscribers)
			for idx := int64(0); idx < tt.args.subscribers; idx++ {
				idxdup := idx

				tt.b.SubscribeOn(tt.args.scheduler, func(state myState, old myState, action Action) {
					atomic.AddInt64(&tt.called, 1)
					log.Printf("Subscriber %d: got called: %d state:%v\n", idxdup, tt.called, state)
				})
			}

			log.Printf("Dispatch: actions: %d\n", tt.args.actions)
			for idx := int64(0); idx < tt.args.actions; idx++ {
				idx := idx
				log.Printf("Dispatch: %d action: %v\n", idx, tt.args.action)
				if tt.args.action != nil {
					tt.b.Dispatch(tt.args.action)
				} else {
					tt.b.Dispatch(&addAction{
						value: fmt.Sprintf("%d", idx%10),
					})
				}
			}

			// subscriber scheduler
			subscriberScheduler.Stop()
			subscriberScheduler.WaitForScheduler()

			// store scheduler
			tt.b.Stop()
			tt.b.WaitForStore()

			if tt.want != tt.called {
				t.Errorf("SubscribeOn: want %d, got %d, state %v, action %v", tt.want, tt.called, tt.b.getState(), tt.args.action)
			}
		})
	}
}

func Test_baseStore_SubscriberDispatchSerialized(t *testing.T) {

	type args[S State] struct {
		scheduler   sched.Scheduler
		action      Action
		actions     int64 // actions to dispatch
		subscribers int64 // subscribers to add
	}
	type testCaseSubscribeOnMany[S State] struct {
		name      string
		b         Store[S]
		args      args[S]
		called    int64 // temporal variable for a testcase
		collected string
		want      int64 // call times
	}

	var actionLimit int64 = 64
	var subscriberlimit int64 = 1

	tests := []testCaseSubscribeOnMany[myState]{
		{
			name: "background - action many - subscriber many",
			b:    newMyStateStore(),
			args: args[myState]{
				scheduler:   sched.Background,
				action:      nil,
				actions:     actionLimit,
				subscribers: subscriberlimit,
			},
			called:    0,
			collected: "",
			want:      subscriberlimit + actionLimit*subscriberlimit,
		},
	}

	//logger.SetLogEnable(true)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			log.Println("Subscriber: subscribers:", tt.args.subscribers)
			for idx := int64(0); idx < tt.args.subscribers; idx++ {
				//idxdup := idx
				tt.b.SubscribeOn(tt.args.scheduler, func(state myState, old myState, action Action) {
					atomic.AddInt64(&tt.called, 1)
					log.Printf("Subscriber %d: got called: %d state:%v\n", idx, tt.called, state)
					tt.collected = tt.collected + state.value
				})
			}

			log.Printf("Dispatch: actions: %d, action=%v", tt.args.actions, tt.args.action)
			for idx := int64(0); idx < tt.args.actions; idx++ {
				idx := idx
				log.Printf("Dispatch: idx:%d action", idx)
				tt.b.Dispatch(&setAction{
					value: fmt.Sprintf("%d", idx%10),
				})
			}

			tt.b.Stop()
			tt.b.WaitForStore()

			if tt.want != tt.called {
				t.Errorf("SubscribeOn: want %d, got %d, state %v, action %v", tt.want, tt.called, tt.b.getState(), tt.args.action)
			}

			wantCollected := func() (result string) {
				for idx := int64(0); idx < actionLimit; idx++ {
					result = result + fmt.Sprintf("%d", idx%10)
				}
				return
			}()
			if wantCollected != tt.collected {
				t.Errorf("SubscribeOn: want %s, got %s", wantCollected, tt.collected)
			}
		})
	}
}

func Test_baseStore_Subscribe_Dispose(t *testing.T) {

	type args[S State] struct {
		action      Action
		actions     int // actions to dispatch
		subscribers int // subscribers to add
		dispose     int
	}
	type testCaseSubscribe[S State] struct {
		name   string
		b      Store[S]
		args   args[S]
		called int // temporal variable for a testcase
		want   int // subscriber count
	}
	tests := []testCaseSubscribe[myState]{
		{
			name: "action 0 - subscriber 1 - dispose 1",
			b:    newMyStateStore(),
			args: args[myState]{
				action:      nil,
				actions:     0,
				subscribers: 1,
				dispose:     1,
			},
			called: 0,
			want:   0,
		},

		{
			name: "action 1 - subscriber 1 - dispose 1",
			b:    newMyStateStore(),
			args: args[myState]{
				action:      &addAction{"1"},
				actions:     1,
				subscribers: 1,
				dispose:     1,
			},
			called: 0,
			want:   0,
		},
		{
			name: "action 1 - subscriber 2 - dispose 1",
			b:    newMyStateStore(),
			args: args[myState]{
				action:      &addAction{"12"},
				actions:     1,
				subscribers: 2,
				dispose:     1,
			},
			called: 0,
			want:   1,
		},
		{
			name: "action 2 - subscriber 2 - dispose 2",
			b:    newMyStateStore(),
			args: args[myState]{
				action:      &addAction{"22"},
				actions:     2,
				subscribers: 2,
				dispose:     2,
			},
			called: 0,
			want:   0,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			log.Println("Subscriber: subscribers:", tt.args.subscribers)
			var disposeBag []Disposer
			for idx := 0; idx < tt.args.subscribers; idx++ {
				idxdup := idx
				disposer := tt.b.Subscribe(func(state myState, old myState, action Action) {
					tt.called++
					log.Printf("Subscriber %d: got called: %d state:%v\n", idxdup, tt.called, state)
				})
				disposeBag = append(disposeBag, disposer)
			}

			log.Printf("Dispatch: actions: %d, action=%v", tt.args.actions, tt.args.action)
			for idx := 0; idx < tt.args.actions; idx++ {
				log.Printf("Dispatch: idx:%d action", idx)
				tt.b.Dispatch(tt.args.action)
			}

			for idx, disposer := range disposeBag {
				if idx < tt.args.dispose {
					log.Printf("Dispose: idx:%d", idx)
					disposer.Dispose()
				}
			}

			tt.b.Stop()
			tt.b.WaitForStore()

			got := len(tt.b.(*baseStore[myState]).subscribers)
			if tt.want != got {
				t.Errorf("Dispose: subscribers want %d, got %d", tt.want, got)
			}
		})
	}
}
