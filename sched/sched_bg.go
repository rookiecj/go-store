package sched

import (
	"github.com/rookiecj/go-store/logger"
	"sync"
)

type backgroundScheduler struct {
	taskCount int
	lock      *sync.Mutex
	signal    *sync.Cond
	doneWG    sync.WaitGroup
}

func newBackgroundScheduler() Scheduler {
	lock := &sync.Mutex{}
	scheduler := &backgroundScheduler{
		lock:   lock,
		signal: sync.NewCond(lock),
	}
	return scheduler
}

func (c *backgroundScheduler) start() {
	//c.lock = &sync.Mutex{}
	//c.signal = sync.NewCond(c.lock)
}

func (c *backgroundScheduler) Stop() {

}

func (c *backgroundScheduler) WaitForIdle() {
	c.lock.Lock()
	logger.Debugf("BG: WaitForIdle: taskCount=%d \n", c.taskCount)
	for c.taskCount > 0 {
		c.signal.Wait()
	}
	c.lock.Unlock()
}

func (c *backgroundScheduler) WaitForScheduler() {
	c.lock.Lock()
	logger.Debugf("BG: WaitForScheduler: taskCount=%d\n", c.taskCount)
	for c.taskCount > 0 {
		c.signal.Wait()
	}
	c.lock.Unlock()
}

func (c *backgroundScheduler) Schedule(task TaskFunc) error {
	logger.Debugf("BG: Schedule:\n")
	c.lock.Lock()
	c.taskCount++
	c.lock.Unlock()

	go func() {
		logger.Debugf("BG: Schedule: run task\n")
		task()

		c.lock.Lock()
		c.taskCount--
		c.signal.Signal()
		c.lock.Unlock()
	}()
	return nil
}
