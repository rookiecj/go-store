package sched

import (
	"testing"
)

func Test_mainScheduler_Schedule(t *testing.T) {
	type args struct {
		concurrent int
		task       TaskFunc
	}

	limit := 1_000_000

	sharedVariableWithNoLock := 0

	tests := []struct {
		name string
		s    Scheduler
		args args
		want int
	}{
		{
			name: "no tasks",
			s:    NewMainScheduler(),
			args: args{
				concurrent: 0,
				task: func() {
					sharedVariableWithNoLock++
				},
			},
			want: 0,
		},

		{
			name: "1 task",
			s:    NewMainScheduler(),
			args: args{
				concurrent: 1,
				task: func() {
					sharedVariableWithNoLock++
				},
			},
			want: 1,
		},

		{
			name: "concurrent tasks - schedule them on same context",
			s:    NewMainScheduler(),
			args: args{
				concurrent: limit,
				task: func() {
					sharedVariableWithNoLock++
				},
			},
			want: limit,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.s

			sharedVariableWithNoLock = 0
			c.Start()

			for idx := 0; idx < tt.args.concurrent; idx++ {
				c.Schedule(tt.args.task)
			}

			c.Stop()
			c.WaitForScheduler()

			got := sharedVariableWithNoLock
			if tt.want != got {
				t.Errorf("Schedule want %v got %v", tt.want, got)
			}
		})
	}
}
