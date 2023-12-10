package store

var (
	// Immediate runs a task on caller or dispatcher's context
	Immediate = &immediateScheduler{}
	// Main context
	Main = &mainScheduler{}
	// Background context
	Background = &backgroundScheduler{}
)

type Callback func()
type Scheduler interface {
	Schedule(task Task, onCompleted Callback)
}

type immediateScheduler struct{}
type mainScheduler struct{}
type backgroundScheduler struct{}

func (c *immediateScheduler) Schedule(task Task, onCompleted Callback) {
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
