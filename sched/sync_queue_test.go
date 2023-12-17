package sched

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

type myState struct {
	id    int
	value string
}

func TestNewSyncQueue_NewClose(t *testing.T) {

	t.Run("NewClose", func(t *testing.T) {
		sq := NewSyncQueue[myState]()
		if sq.Len() != 0 {
			t.Errorf("Producer/Consumer want 0 got %d", sq.Len())
		}
	})
}

func TestNewSyncQueue_PushPop(t *testing.T) {

	t.Run("PushPop-Sequential", func(t *testing.T) {
		sq := NewSyncQueue[myState]()

		limit := 100000
		for idx := 0; idx < limit; idx++ {
			sq.Push(myState{
				id:    idx,
				value: fmt.Sprintf("%d", idx),
			})
		}

		if sq.Len() != limit {
			t.Errorf("PushPop want %d got %d", limit, sq.Len())
		}
		for idx := 0; idx < limit; idx++ {
			item, err := sq.Pop()
			if err != nil {
				t.Errorf("PushPop got error %d, %v", idx, err)
				break
			}
			if item.id != idx {
				t.Errorf("PushPop want %v got %v", idx, item)
				break
			}
		}
		if sq.Len() != 0 {
			t.Errorf("PushPop Len want %v got %v", 0, sq.Len())
		}
	})
}

func TestNewSyncQueue_PushPeek(t *testing.T) {

	t.Run("PushPeek-1", func(t *testing.T) {
		sq := NewSyncQueue[myState]()

		want := myState{
			id:    1234,
			value: "1234",
		}
		sq.Push(want)
		if sq.Len() != 1 {
			t.Errorf("PushPeek size want %d got %d", 1, sq.Len())
		}

		got, err := sq.Peek()
		if err != nil {
			t.Errorf("PushPeek err %v", err)
			return
		}
		if sq.Len() != 1 {
			t.Errorf("PushPeek size want %d got %d", 1, sq.Len())
		}

		if !reflect.DeepEqual(want, got) {
			t.Errorf("PushPeek want %v got %v", want, got)
		}

	})
}

func TestNewSyncQueue_PopPush(t *testing.T) {

	t.Run("PopPush", func(t *testing.T) {
		sq := NewSyncQueue[myState]()

		limit := 100000

		cwg := sync.WaitGroup{}

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			wg.Done()

			cwg.Add(1)
			for idx := 0; idx < limit; idx++ {
				item, err := sq.Pop()
				if err != nil {
					t.Errorf("PopPush got error %d, %v", idx, err)
					break
				}
				if item.id != idx {
					t.Errorf("PopPush want %v got %v", idx, item)
					break
				}
			}
			cwg.Done()
		}()
		wg.Wait()

		time.Sleep(100 * time.Millisecond)

		for idx := 0; idx < limit; idx++ {
			sq.Push(myState{
				id:    idx,
				value: fmt.Sprintf("%d", idx),
			})
		}

		cwg.Wait()

		if sq.Len() != 0 {
			t.Errorf("PushPop Len want %v got %v", 0, sq.Len())
		}
	})
}

func TestNewSyncQueue_MultipleProducersSingleConsumer(t *testing.T) {

	t.Run("push/pop multiple producers, single consumer", func(t *testing.T) {

		sq := NewSyncQueue[myState]()

		limit := 1000000
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			wg.Add(limit)
			for idx := 0; idx < limit; idx++ {
				go func() {
					sq.Push(myState{
						id:    idx,
						value: fmt.Sprintf("%d", idx),
					})
					wg.Done()
				}()
			}
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			for idx := 0; idx < limit; idx++ {
				sq.Peek()
				sq.Pop()
			}
			wg.Done()
		}()

		wg.Wait()

		if sq.Len() != 0 {
			t.Errorf("Producer/Consumer want %d got %d", 0, sq.Len())
		}
	})
}

func TestNewSyncQueue_SingleProducerMultipleConsumers(t *testing.T) {

	t.Run("push/pop single producer/multiple consumers", func(t *testing.T) {

		sq := NewSyncQueue[myState]()

		limit := 1000000
		wg := sync.WaitGroup{} // ptr
		wg.Add(1)
		go func() {
			for idx := 0; idx < limit; idx++ {
				sq.Push(myState{
					id:    idx,
					value: fmt.Sprintf("%d", idx),
				})
			}
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			wg.Add(limit)
			for idx := 0; idx < limit; idx++ {
				go func() {
					sq.Peek()
					sq.Pop()
					wg.Done()
				}()
			}
			wg.Done()
		}()

		wg.Wait()

		if sq.Len() != 0 {
			t.Errorf("Producer/Consumer want 0 got %d", sq.Len())
		}
	})
}

func TestNewSyncQueue_MultipleProducersMultipleConsumers(t *testing.T) {

	t.Run("push/pop multiple producers/multiple consumers", func(t *testing.T) {

		sq := NewSyncQueue[myState]()

		limit := 1000000
		wg := sync.WaitGroup{} // ptr
		wg.Add(1)
		go func() {
			wg.Add(limit)
			for idx := 0; idx < limit; idx++ {
				go func() {
					sq.Push(myState{
						id:    idx,
						value: fmt.Sprintf("%d", idx),
					})
					wg.Done()
				}()
			}
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			wg.Add(limit)
			for idx := 0; idx < limit; idx++ {
				go func() {
					sq.Peek()
					sq.Pop()
					wg.Done()
				}()
			}
			wg.Done()
		}()

		wg.Wait()

		if sq.Len() != 0 {
			t.Errorf("Producer/Consumer want %d got %d", 0, sq.Len())
		}
	})
}
