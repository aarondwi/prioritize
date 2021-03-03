package circular

import (
	"sync"

	"github.com/aarondwi/prioritize/common"
)

// CircularQueue implements bounded-size uint64 queue,
// and doesnt care about priority given to it
type CircularQueue struct {
	mu          *sync.Mutex
	notEmpty    *sync.Cond
	arr         []uint64
	maxSize     int
	currentSize int
	head        int
	tail        int
}

// NewCircularQueue creates a CircularQueue with size n
func NewCircularQueue(n int) *CircularQueue {
	mu := &sync.Mutex{}
	notEmpty := sync.NewCond(mu)

	return &CircularQueue{
		mu:          mu,
		notEmpty:    notEmpty,
		arr:         make([]uint64, n),
		maxSize:     n,
		currentSize: 0,
		head:        0,
		tail:        0,
	}
}

// PushOrError item.ID into circularQueue, or fail if no slots available
func (c *CircularQueue) PushOrError(item common.QItem) error {
	c.mu.Lock()
	if c.isFull() {
		c.mu.Unlock()
		return common.ErrQueueIsFull
	}
	c.arr[c.head] = item.ID
	c.head = c.getNextIndex(c.head)
	c.currentSize++
	c.notEmpty.Signal()
	c.mu.Unlock()
	return nil
}

// PopOrWait returns 1 item from queue, or wait if none exists
func (c *CircularQueue) PopOrWait() common.QItem {
	c.mu.Lock()
	for c.isEmpty() {
		c.notEmpty.Wait()
	}
	result := common.QItem{ID: c.arr[c.tail]}
	c.tail = c.getNextIndex(c.tail)
	c.currentSize--
	c.mu.Unlock()
	return result
}

func (c *CircularQueue) getNextIndex(index int) int {
	if index == c.maxSize-1 {
		return 0
	}
	return index + 1
}

// isFull checks whether circularQueue has no remaining slots
func (c *CircularQueue) isFull() bool {
	return c.currentSize == c.maxSize
}

// isEmpty checks whether circularQueue has remaining slots
func (c *CircularQueue) isEmpty() bool {
	return c.currentSize == 0
}
