package sched

import (
	"errors"
	"github.com/rookiecj/go-store/logger"
	"sync"
	"sync/atomic"
)

var ErrNotStarted = errors.New("scheduler not started")

type mainScheduler struct {
	taskCount  atomic.Int64
	taskQ      SyncQueue[TaskFunc]
	tasks      []TaskFunc
	idleLock   *sync.Mutex
	idleSignal *sync.Cond
	doneWG     sync.WaitGroup
}

func NewMainScheduler() Scheduler {
	idleLock := &sync.Mutex{}
	scheduler := &mainScheduler{
		taskCount:  atomic.Int64{},
		idleLock:   idleLock,
		idleSignal: sync.NewCond(idleLock),
		doneWG:     sync.WaitGroup{},
	}
	scheduler.start()
	return scheduler
}

func (c *mainScheduler) start() {
	if c == nil {
		return
	}

	// not to exit loop
	c.taskCount.Store(1)
	c.taskQ = NewSyncQueue[TaskFunc]()

	// Main context
	wg := sync.WaitGroup{}
	wg.Add(1)

	// for WaitForScheduler
	c.doneWG.Add(1)

	go func() {
		wg.Done()
		logger.Infof("mainScheduler: run\n")
		//c.taskCount.Add(1)
		for c.taskCount.Load() > 0 {
			//logger.Debugf("mainScheduler: task pop: %d\n", c.taskQ.Len())
			if task, err := c.taskQ.Pop(); err == nil {
				//logger.Debugf("mainScheduler: task run remains:%d\n", c.taskQ.Len())
				task()
			}

			c.idleLock.Lock()
			c.idleSignal.Signal()
			c.idleLock.Unlock()
		}
		logger.Debugf("mainScheduler: loop exit: remains:%d\n", c.taskQ.Len())
		if remains := c.taskQ.Len(); remains != 0 {
			logger.LogForcedf("mainScheduler: exit remains %d\n", remains)
			for c.taskQ.Len() > 0 {
				c.taskQ.Pop()
			}
		}

		logger.Debugf("mainScheduler: done\n")
		// done WaitForScheduler
		c.doneWG.Done()
	}()
	wg.Wait()
}

func (c *mainScheduler) Stop() {
	if c == nil {
		return
	}
	logger.Debugf("mainScheduler: Stop\n")

	c.stop()
}

func (c *mainScheduler) stop() {
	c.Schedule(func() {
		logger.Debugf("mainScheduler: stop task\n")
		c.taskCount.Add(-1)
	})
}

func (c *mainScheduler) WaitForScheduler() {
	if c == nil {
		return
	}
	logger.Debugf("mainScheduler: WaitForScheduler")

	c.doneWG.Wait()
}

func (c *mainScheduler) WaitForIdle() {
	if c == nil {
		return
	}
	logger.Debugf("mainScheduler: WaitForIdle")

	c.idleLock.Lock()
	for c.taskQ.Len() > 0 {
		c.idleSignal.Wait()
	}
	c.idleLock.Unlock()
}

func (c *mainScheduler) Schedule(task TaskFunc) error {
	if c == nil {
		return nil
	}
	//logger.Debugf("mainScheduler: Schedule")
	if c.taskQ == nil {
		return ErrNotStarted
	}

	c.taskQ.Push(task)
	return nil
}
