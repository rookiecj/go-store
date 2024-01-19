package sched

type TaskFunc func()

// Scheduler schedules tasks
type Scheduler interface {
	// Start starts the scheduler
	start()

	// Schedule schedules a task
	Schedule(task TaskFunc) error

	// idle -> close model

	// WaitForIdle waits for idle
	WaitForIdle()
	// Close scheduler
	//Close()

	// stop -> wait model

	// Stop request scheduler to stop
	Stop()
	// WaitForScheduler waits for stopping scheduler
	WaitForScheduler()
}

var (
	// Immediate runs tasks immediately, no schedule
	Immediate = newImmScheduler()
	Main      = NewMainScheduler()
	// Background context, run tasks in any order
	Background = newBackgroundScheduler()
)
