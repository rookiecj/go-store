package sched

import "sync"

type backgroundScheduler struct {
	taskCount int
	lock      *sync.Mutex
	signal    *sync.Cond
	doneWG    sync.WaitGroup
}

func newBackgroundScheduler() Scheduler {
	lock := &sync.Mutex{}
	return &backgroundScheduler{
		lock:   lock,
		signal: sync.NewCond(lock),
	}
}

func (c *backgroundScheduler) Start() {
	//c.lock = &sync.Mutex{}
	//c.signal = sync.NewCond(c.lock)
}

func (c *backgroundScheduler) Stop() {

}

func (c *backgroundScheduler) WaitForScheduler() {
	c.lock.Lock()
	for c.taskCount > 0 {
		c.signal.Wait()
	}
	c.lock.Unlock()
}

func (c *backgroundScheduler) Schedule(task TaskFunc) {
	c.lock.Lock()
	c.taskCount++
	c.lock.Unlock()

	go func() {
		task()

		c.lock.Lock()
		c.taskCount--
		c.signal.Signal()
		c.lock.Unlock()
	}()
}
