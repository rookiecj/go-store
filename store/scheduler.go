package store

var (
	//  Immedidate runs tasks on the caller's context
	Immediate = &immediateScheduler{}
	Main      = &mainScheduler{}
)

type Callback func()
type Scheduler interface {
	Schedule(task Task, onCompleted Callback)
}

type Task interface {
	Do()
}

type immediateScheduler struct{}
type mainScheduler struct{}

func (c *immediateScheduler) Schedule(task Task, onCompleted Callback) {
	// run task on the caller's context
	task.Do()
	if onCompleted != nil {
		onCompleted()
	}
}

func (c *mainScheduler) Schedule(task Task, onCompleted Callback) {
	// TODO run it in main context
	task.Do()
	if onCompleted != nil {
		onCompleted()
	}
}
