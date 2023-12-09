package store

var (
	//  Immedidate runs tasks on the caller's context
	Immediate = &immediateScheduler{}
	Main      = &mainScheduler{}
)

type Scheduler interface {
	Schedule(task Task)
}

type Task interface {
	Do()
}

type immediateScheduler struct{}
type mainScheduler struct{}

func (c *immediateScheduler) Schedule(task Task) {
	// run task on the caller's context
	task.Do()
}

func (c *mainScheduler) Schedule(task Task) {
	// just run
	task.Do()
}
