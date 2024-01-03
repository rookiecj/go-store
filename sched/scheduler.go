package sched

type TaskFunc func()

// Scheduler schedules tasks
type Scheduler interface {
	// Start starts the scheduler
	Start()
	// Schedule schedules a task
	Schedule(task TaskFunc)
	// Stop stops the scheduler
	Stop()
	// WaitForScheduler waits for the scheduler to stop
	WaitForScheduler()
}

var (
	// Immediate runs tasks immediately, no schedule
	Immediate = newImmScheduler()
	// Main runs tasks on the same context in order which is arrived. It is shared across all stores.
	Main = NewMainScheduler()
	// Background context, run tasks in any order
	Background = newBackgroundScheduler()
)
