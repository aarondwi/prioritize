package priority

import (
	"sync"

	"github.com/aarondwi/prioritize/common"
	"github.com/aarondwi/prioritize/linkedslice"
)

// PriorityQueue is a queue in which
// always try to return higher priority first
//
// It is not designed using heap,
// because as we bound the number of priority allowed,
// this implementation can reduce check-and-swap nature of `heap`-based implementation
// and also getting a much clearer code-base
type PriorityQueue struct {
	// synchronization primitive
	mu       *sync.Mutex
	notEmpty *sync.Cond

	// we separate number tracking from the priorityQueues
	// so checking numberOfTasksInEachQueue just need 1 cache miss (putting into cpu L1 cache)
	numberOfTasksInEachQueue []int

	// we also create separate queues for each priority
	// so it is simple to push/pop the item
	queues []*linkedslice.LinkedSlice

	// simple metadata
	limitPriority int
	size          int
	sizeLimit     int
	running       bool
}

func NewPriorityQueue(sizeLimit, numOfPriority int) (*PriorityQueue, error) {
	if sizeLimit <= 0 || numOfPriority <= 0 {
		return nil, common.ErrParamShouldBePositive
	}

	mu := &sync.Mutex{}
	notEmpty := sync.NewCond(mu)

	numberOfTasksInEachQueue := make([]int, numOfPriority)
	queues := make([]*linkedslice.LinkedSlice, numOfPriority)

	return &PriorityQueue{
		mu:                       mu,
		notEmpty:                 notEmpty,
		numberOfTasksInEachQueue: numberOfTasksInEachQueue,
		queues:                   queues,
		limitPriority:            numOfPriority,
		size:                     0,
		sizeLimit:                sizeLimit,
		running:                  true,
	}, nil
}

// PushOrError put the item into the pq, and returns error if no slot available
func (pq *PriorityQueue) PushOrError(item common.QItem) error {
	if item.Priority < 0 || item.Priority >= pq.limitPriority {
		return common.ErrPriorityOutOfRange
	}

	pq.mu.Lock()
	if !pq.running {
		pq.mu.Unlock()
		return common.ErrQueueIsClosed
	}
	if pq.size == pq.sizeLimit {
		pq.mu.Unlock()
		return common.ErrQueueIsFull
	}

	if pq.queues[item.Priority] == nil {
		pq.queues[item.Priority] = linkedslice.NewLinkedSlice()
	}
	err := pq.queues[item.Priority].PushOrError(item)
	// meaning already closed, cause linkedslices is unbounded
	if err != nil {
		pq.mu.Unlock()
		return err
	}
	pq.numberOfTasksInEachQueue[item.Priority]++
	pq.size++

	pq.notEmpty.Signal()
	pq.mu.Unlock()
	return nil
}

// PopOrWaitTillClose returns 1 QItem from pq, or waits if none exists
func (pq *PriorityQueue) PopOrWaitTillClose() (common.QItem, error) {
	pq.mu.Lock()
	if !pq.running {
		pq.mu.Unlock()
		return common.MinQItem, common.ErrQueueIsClosed
	}

	for pq.size == 0 {
		pq.notEmpty.Wait()
		// double check, ensuring see the changes after wait call
		if !pq.running {
			pq.mu.Unlock()
			return common.MinQItem, common.ErrQueueIsClosed
		}
	}

	// we will undoubtedly get at least one item
	priorityToRetrieve := -1
	for i := pq.limitPriority - 1; i >= 0; i-- {
		if pq.numberOfTasksInEachQueue[i] > 0 {
			priorityToRetrieve = i
			break
		}
	}

	// if we wait blindly, it gonna stuck
	// but we are tracking it manually, ensuring it will never wait
	qitem, err := pq.queues[priorityToRetrieve].PopOrWaitTillClose()
	if err != nil {
		// the only error possible here is closed already
		// so we just continue it
		pq.mu.Unlock()
		return common.MinQItem, err
	}
	result := common.QItem{
		ID:       qitem.ID,
		Priority: priorityToRetrieve,
	}
	pq.numberOfTasksInEachQueue[priorityToRetrieve]--
	pq.size--

	pq.mu.Unlock()
	return result, nil
}

// Close PriorityQueue, preventing it from accepting new request
func (pq *PriorityQueue) Close() {
	pq.mu.Lock()
	pq.running = false
	for i := 0; i < pq.limitPriority; i++ {
		if pq.queues[i] != nil {
			pq.queues[i].Close()
		}
	}
	pq.notEmpty.Broadcast()
	pq.mu.Unlock()
}
