package store

var (
	// Main context
	Main = &mainScheduler{}
	// Background context
	Background = &backgroundScheduler{}
	// Dispatcher context
	dispatchScheduler = &dispatcherScheduler{}
)

type Callback func()
type Scheduler interface {
	Schedule(task Task, onCompleted Callback)
}

type mainScheduler struct{}
type backgroundScheduler struct{}
type dispatcherScheduler struct{}

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

func (c *dispatcherScheduler) Schedule(task Task, onCompleted Callback) {
	// TODO run it in dispatcher context
	task.Do()
	if onCompleted != nil {
		onCompleted()
	}
}
