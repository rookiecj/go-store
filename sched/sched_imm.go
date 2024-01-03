package sched

type immediateScheduler struct{}

func newImmScheduler() Scheduler {
	return &immediateScheduler{}
}

func (c *immediateScheduler) Start() {}

func (c *immediateScheduler) Stop() {}

func (c *immediateScheduler) Schedule(task TaskFunc) {
	task()
}

func (c *immediateScheduler) WaitForScheduler() {}
