package sched

import (
	"errors"
	"sync"
)

var ErrNoItem = errors.New("no items")

type SyncQueue[T any] interface {
	Push(item T)
	Pop() (T, error)
	Peek() (T, error)
	Len() int
}

type syncQueue[T any] struct {
	lock   *sync.Mutex
	signal *sync.Cond

	items []T
}

func NewSyncQueue[T any]() SyncQueue[T] {
	lock := &sync.Mutex{}
	q := syncQueue[T]{
		lock:   lock,               // ptr
		signal: sync.NewCond(lock), // ptr
	}
	return &q
}

func (c *syncQueue[T]) Push(item T) {
	if c == nil {
		return
	}
	c.lock.Lock()
	c.items = append(c.items, item)
	c.signal.Signal()
	c.lock.Unlock()
}

func (s *syncQueue[T]) Pop() (item T, err error) {
	if s == nil {
		err = errors.New("ref is nil")
		return
	}
	s.lock.Lock()
	for len(s.items) == 0 {
		s.signal.Wait()
	}
	item = s.items[0]
	s.items = s.items[1:]
	s.lock.Unlock()
	return item, nil
}

func (s *syncQueue[T]) Peek() (item T, err error) {
	if s == nil {
		err = errors.New("ref is nil")
		return
	}

	s.lock.Lock()
	if len(s.items) == 0 {
		err = ErrNoItem
		s.lock.Unlock()
		return
	}
	item = s.items[0]
	s.lock.Unlock()
	return item, nil
}

func (s *syncQueue[T]) Len() int {
	if s == nil {
		return 0
	}
	s.lock.Lock()
	size := len(s.items)
	s.lock.Unlock()
	return size
}
