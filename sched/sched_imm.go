package sched

type immediateScheduler struct{}

func newImmScheduler() Scheduler {
	return &immediateScheduler{}
}

func (c *immediateScheduler) start() {}

func (c *immediateScheduler) Stop() {}

func (c *immediateScheduler) Schedule(task TaskFunc) error {
	task()
	return nil
}

func (c *immediateScheduler) WaitForIdle() {
}

func (c *immediateScheduler) WaitForScheduler() {}
