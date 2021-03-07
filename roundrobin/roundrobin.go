package roundrobin

import (
	"sync"

	"github.com/aarondwi/prioritize/common"
	"github.com/aarondwi/prioritize/linkedslice"
)

// RoundRobinPriorityQueue is a priority queue in which
// each priority gets a chance to return value,
// starting from first item put going downwards,
// and then rolled back from highest.
//
// This behavior allows some starvation prevention for lower priorities,
// assuming that highest priority tasks have much lower number of tasks,
// else it gonna be pretty much not useful, just like random/normal queue.
//
// Internally, we are using unbounded linkedslice.
// Because we may have limits, but each priority can takes up to that limit, and using linkedslice can reduce unwanted memory usage.
type RoundRobinPriorityQueue struct {
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
	limitPriority             int
	size                      int
	sizeLimit                 int
	currentPriorityToRetrieve int
	running                   bool
}

// NewRoundRobinPriorityQueue creates our RR PQ.
//
// It caps at sizeLimit, and allows priorirty [0,numOfPriority)
func NewRoundRobinPriorityQueue(sizeLimit, numOfPriority int) (*RoundRobinPriorityQueue, error) {
	if sizeLimit <= 0 || numOfPriority <= 0 {
		return nil, common.ErrParamShouldBePositive
	}

	mu := &sync.Mutex{}
	notEmpty := sync.NewCond(mu)

	numberOfTasksInEachQueue := make([]int, numOfPriority)
	queues := make([]*linkedslice.LinkedSlice, numOfPriority)

	return &RoundRobinPriorityQueue{
		mu:                        mu,
		notEmpty:                  notEmpty,
		numberOfTasksInEachQueue:  numberOfTasksInEachQueue,
		queues:                    queues,
		limitPriority:             numOfPriority,
		size:                      0,
		sizeLimit:                 sizeLimit,
		currentPriorityToRetrieve: -1,
		running:                   true,
	}, nil
}

// PushOrError put the item into the rrpq, and returns error if no slot available
func (rr *RoundRobinPriorityQueue) PushOrError(item common.QItem) error {
	if item.Priority < 0 || item.Priority >= rr.limitPriority {
		return common.ErrPriorityOutOfRange
	}

	rr.mu.Lock()
	if !rr.running {
		rr.mu.Unlock()
		return common.ErrQueueIsClosed
	}

	if rr.size == rr.sizeLimit {
		rr.mu.Unlock()
		return common.ErrQueueIsFull
	}
	if rr.queues[item.Priority] == nil {
		rr.queues[item.Priority] = linkedslice.NewLinkedSlice()
	}
	err := rr.queues[item.Priority].PushOrError(item)
	if err != nil {
		// meaning already closed, cause linkedslices is unbounded
		rr.mu.Unlock()
		return err
	}

	// The only item in the queue, set this to position
	if rr.size == 0 {
		rr.currentPriorityToRetrieve = item.Priority
	}

	// update the tracker too
	rr.numberOfTasksInEachQueue[item.Priority]++
	rr.size++
	rr.notEmpty.Signal()

	rr.mu.Unlock()

	return nil
}

// PopOrWaitTillClose returns 1 QItem from RRPQ, or waits if none exists
func (rr *RoundRobinPriorityQueue) PopOrWaitTillClose() (common.QItem, error) {
	rr.mu.Lock()
	if !rr.running {
		rr.mu.Unlock()
		return common.MinQItem, common.ErrQueueIsClosed
	}

	for rr.size == 0 {
		rr.notEmpty.Wait()
		// double check, ensuring see the changes after wait call
		if !rr.running {
			rr.mu.Unlock()
			return common.MinQItem, common.ErrQueueIsClosed
		}
	}

	// if we wait blindly, it gonna stuck
	// but we are tracking it manually, ensuring it will never wait
	qitem, err := rr.queues[rr.currentPriorityToRetrieve].PopOrWaitTillClose()
	if err != nil {
		// the only error possible here is closed already
		// so we just continue it
		rr.mu.Unlock()
		return common.MinQItem, err
	}
	result := common.QItem{
		ID:       qitem.ID,
		Priority: rr.currentPriorityToRetrieve,
	}
	rr.numberOfTasksInEachQueue[rr.currentPriorityToRetrieve]--
	rr.size--

	if rr.size == 0 {
		//fast path, no need to check rr.numberOfTasksInEachQueue
		rr.currentPriorityToRetrieve = -1
	} else {
		// Check new rr.currentPosToRetrieve position, cause we still have item somewhere
		newPos := -1
		for i := rr.currentPriorityToRetrieve - 1; i >= 0; i-- {
			if rr.numberOfTasksInEachQueue[i] > 0 {
				newPos = i
				break
			}
		}
		// not yet found, meaning remaining items reside on higher index
		// currentPriorityToRetrieve should be the last index to be checked
		if newPos == -1 {
			for i := rr.limitPriority - 1; i >= rr.currentPriorityToRetrieve; i-- {
				if rr.numberOfTasksInEachQueue[i] > 0 {
					newPos = i
					break
				}
			}
		}
		rr.currentPriorityToRetrieve = newPos
	}

	rr.mu.Unlock()
	return result, nil
}

// Close RoundRobinPriorityQueue, preventing it from accepting new request
func (rr *RoundRobinPriorityQueue) Close() {
	rr.mu.Lock()
	rr.running = false
	for i := 0; i < rr.limitPriority; i++ {
		if rr.queues[i] != nil {
			rr.queues[i].Close()
		}
	}
	rr.notEmpty.Broadcast()
	rr.mu.Unlock()
}
