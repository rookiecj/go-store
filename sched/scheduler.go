package sched

import (
	"github.com/rookiecj/go-store/logger"
	"sync"
)

type TaskFunc func()

type Scheduler interface {
	Start()
	Stop()
	StopWait()
	Schedule(task TaskFunc)
}

type immediateScheduler struct{}

type mainScheduler struct {
	//taskChan  chan func()
	taskCount int
	taskQ     SyncQueue[TaskFunc]
}

type backgroundScheduler struct {
	taskCount int
	lock      *sync.Mutex
	signal    *sync.Cond
}

var (
	// Immediate runs a task immediately, no schedule
	Immediate = &immediateScheduler{}
	// Main runs a task on the same context.
	Main Scheduler = &mainScheduler{}
	// Background context
	Background Scheduler = newBackgroundScheduler()
)

func (c *immediateScheduler) Start() {}

func (c *immediateScheduler) Stop() {}

func (c *immediateScheduler) Schedule(task TaskFunc) {
	task()
}

func (c *immediateScheduler) StopWait() {}

func (c *mainScheduler) Start() {
	c.taskQ = NewSyncQueue[TaskFunc]()

	// Main context
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		wg.Done()
		logger.Infof("mainScheduler:")
		c.taskCount++
		for c.taskCount > 0 {
			if task, err := c.taskQ.Pop(); err == nil {
				task()
			}
		}
		logger.Infof("mainScheduler: exit")
	}()
	wg.Wait()
}

func (c *mainScheduler) Stop() {
	go func() {
		c.taskQ.Push(func() {
			c.taskCount--
		})
	}()
}

func (c *mainScheduler) StopWait() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	c.taskQ.Push(func() {
		c.taskCount--
		wg.Done()
	})
	wg.Wait()
}

func (c *mainScheduler) Schedule(task TaskFunc) {
	c.taskQ.Push(task)
}

//
// background scheduler
//

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

func (c *backgroundScheduler) StopWait() {
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
