package store

var (
	// Caller runs a task on that caller's context immediately
	Caller = &callerScheduler{}
	// Main context
	Main = &mainScheduler{}
	// Background context
	Background = &backgroundScheduler{}
)

type Callback func()
type Scheduler interface {
	Schedule(task Task, onCompleted Callback)
}

type Task interface {
	// Do runs a Task
	Do()
	// Result is only available after Do
	Result() any
}

type callerScheduler struct{}
type mainScheduler struct{}
type backgroundScheduler struct{}

func (c *callerScheduler) Schedule(task Task, onCompleted Callback) {
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

func (c *backgroundScheduler) Schedule(task Task, onCompleted Callback) {
	// TODO run it in background context
	task.Do()
	if onCompleted != nil {
		onCompleted()
	}
}
