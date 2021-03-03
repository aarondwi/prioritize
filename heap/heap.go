package heap

import (
	"sync"

	"github.com/aarondwi/prioritize/common"
)

// HeapPriorityQueue is the simplest priority queue.
// It is designed using heap as internal data structure,
// and it does not have any starvation-handling.
type HeapPriorityQueue struct {
	mu        *sync.Mutex
	notEmpty  *sync.Cond
	arr       []common.QItem
	size      int
	sizeLimit int
}

// NewHeapPriorityQueue setups our priorityqueue with the config.
//
// It caps at sizeLimit.
func NewHeapPriorityQueue(sizeLimit int) *HeapPriorityQueue {
	mu := &sync.Mutex{}
	notEmpty := sync.NewCond(mu)

	arr := make([]common.QItem, sizeLimit)
	for i := 0; i < sizeLimit; i++ {
		arr[i] = common.MinQItem
	}
	return &HeapPriorityQueue{
		mu:        mu,
		notEmpty:  notEmpty,
		arr:       arr,
		size:      0,
		sizeLimit: sizeLimit,
	}
}

func (hpq *HeapPriorityQueue) leaf(index int) bool {
	return (index >= (hpq.size/2) && index <= hpq.size)
}

func (hpq *HeapPriorityQueue) parent(index int) int {
	return (index - 1) / 2
}

func (hpq *HeapPriorityQueue) leftchild(index int) int {
	return 2*index + 1
}

func (hpq *HeapPriorityQueue) rightchild(index int) int {
	return 2*index + 2
}

func (hpq *HeapPriorityQueue) swap(first, second int) {
	temp := hpq.arr[first]
	hpq.arr[first] = hpq.arr[second]
	hpq.arr[second] = temp
}

func (hpq *HeapPriorityQueue) greater(first, second int) bool {
	return hpq.arr[first].Priority > hpq.arr[second].Priority
}

func (hpq *HeapPriorityQueue) up(index int) {
	for hpq.greater(index, hpq.parent(index)) {
		hpq.swap(index, hpq.parent(index))
		index = hpq.parent(index)
	}
}

func (hpq *HeapPriorityQueue) down(current int) {
	if hpq.leaf(current) {
		return
	}
	largest := current
	leftChildIndex := hpq.leftchild(current)
	rightChildIndex := hpq.rightchild(current)
	if leftChildIndex < hpq.size && hpq.greater(leftChildIndex, largest) {
		largest = leftChildIndex
	}
	if rightChildIndex < hpq.size && hpq.greater(rightChildIndex, largest) {
		largest = rightChildIndex
	}
	if largest != current {
		hpq.swap(current, largest)
		hpq.down(largest)
	}
}

// PushOrError pushes an item into the priorityqueue, or returning error if full
func (hpq *HeapPriorityQueue) PushOrError(item common.QItem) error {
	hpq.mu.Lock()
	if hpq.size == hpq.sizeLimit {
		hpq.mu.Unlock()
		return common.ErrQueueIsFull
	}
	hpq.arr[hpq.size] = item
	hpq.up(hpq.size)
	hpq.size++
	hpq.notEmpty.Signal()
	hpq.mu.Unlock()
	return nil
}

// PopOrWait remove + returns one item from the priorityqueue, or wait until a task is available
func (hpq *HeapPriorityQueue) PopOrWait() common.QItem {
	hpq.mu.Lock()
	for hpq.size == 0 {
		hpq.notEmpty.Wait()
	}
	top := hpq.arr[0]
	hpq.arr[0] = hpq.arr[hpq.size-1]
	hpq.arr[hpq.size-1] = common.MinQItem
	hpq.size--
	hpq.down(0)
	hpq.mu.Unlock()
	return top
}
