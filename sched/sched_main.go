package sched

import (
	"errors"
	"github.com/rookiecj/go-store/logger"
	"sync"
)

type mainScheduler struct {
	//taskChan  chan func()
	taskCount int
	taskQ     SyncQueue[TaskFunc]
	doneWG    sync.WaitGroup
}

func NewMainScheduler() Scheduler {
	return &mainScheduler{
		doneWG: sync.WaitGroup{},
	}
}

func (c *mainScheduler) Start() {
	if c == nil {
		return
	}

	c.taskCount = 0
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
		logger.Infof("mainScheduler: exit remains %d", c.taskQ.Len())
		c.doneWG.Done()
	}()
	wg.Wait()
}

func (c *mainScheduler) Stop() {
	if c == nil {
		return
	}
	c.doneWG.Add(1)
	go func() {
		c.taskQ.Push(func() {
			c.taskCount--
		})
	}()
}

func (c *mainScheduler) WaitForScheduler() {
	if c == nil {
		return
	}
	c.doneWG.Wait()
}

func (c *mainScheduler) Schedule(task TaskFunc) {
	if c == nil {
		return
	}
	if c.taskQ == nil {
		panic(errors.New("did you start scheduler?"))
	}
	c.taskQ.Push(task)
}
