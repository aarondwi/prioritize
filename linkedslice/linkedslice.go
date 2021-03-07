package linkedslice

import (
	"log"
	"sync"

	"github.com/aarondwi/prioritize/common"
)

// LinkedSlice is a queue in which never full,
// but also don't care about priority, instead it is FIFO
//
// This can also be used as base of other priority queuing,
// in which they split into multiple internal queue,
// and need each queue to be unbounded
//
// There are 2 pointer needed here.
//
// 1. head maintains the base of the linked list, and pop always takes from head
//
// 2. pushPointer is a pointer pointing to which node new insert should go
//
// As items are popped, head gonna go forward, and the previous one will be put back to pool.
type LinkedSlice struct {
	mu          *sync.Mutex
	notEmpty    *sync.Cond
	head        *internalSlice
	pushPointer *internalSlice
	running     bool
}

// NewLinkedSlice creates our LinkedSlice struct
func NewLinkedSlice() *LinkedSlice {
	mu := &sync.Mutex{}
	notEmpty := sync.NewCond(mu)

	return &LinkedSlice{
		mu:          mu,
		notEmpty:    notEmpty,
		head:        nil,
		pushPointer: nil,
		running:     true,
	}
}

func (ls *LinkedSlice) checkHeadExist() {
	if ls.head == nil {
		ls.head = internalSlicePool.Get().(*internalSlice)
		ls.pushPointer = ls.head
	}
}

// PushOrError insert item into the queue.
// But as this implementation is unbounded, error should always be nil.
// Any error found results in panic, cause it means either
// broken implementation, or some environment issue happens (e.g. OOM).
func (ls *LinkedSlice) PushOrError(item common.QItem) error {
	ls.mu.Lock()

	// double check, ensuring see the changes after lock call
	if !ls.running {
		ls.mu.Unlock()
		return common.ErrQueueIsClosed
	}

	ls.checkHeadExist()
	if !ls.pushPointer.canPush() { //meaning full already
		newSlice := internalSlicePool.Get().(*internalSlice)
		ls.pushPointer.next = newSlice
		ls.pushPointer = newSlice
	}
	err := ls.pushPointer.push(item.ID)
	if err != nil {
		log.Println(err)
		panic("Some implementation/environment goes wrong, cause it should not return any error now")
	}
	ls.notEmpty.Signal()
	ls.mu.Unlock()
	return nil
}

// PopOrWaitTillClose returns 1 item from the queue, or wait if none exists
func (ls *LinkedSlice) PopOrWaitTillClose() (common.QItem, error) {
	ls.mu.Lock()
	// double check, ensuring see the changes after lock call
	if !ls.running {
		ls.mu.Unlock()
		return common.MinQItem, common.ErrQueueIsClosed
	}

	ls.checkHeadExist()
	// because we handle slotsUsedUp check below
	// we don't need to check inside this wait-loop
	for ls.head.isEmpty() {
		ls.notEmpty.Wait()
	}
	result, _ := ls.head.pop()
	if ls.head.slotsUsedUp() {
		usedLS := ls.head
		ls.head = ls.head.next
		putInternalSlice(usedLS)
	}
	ls.mu.Unlock()
	return common.QItem{ID: result}, nil
}

// Close LinkedSlice, preventing it from accepting new request
func (ls *LinkedSlice) Close() {
	ls.mu.Lock()
	ls.running = false
	ls.notEmpty.Broadcast()
	ls.mu.Unlock()
}
