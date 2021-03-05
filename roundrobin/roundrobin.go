package roundrobin

import (
	"errors"
	"sync"
	"sync/atomic"

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
	// so checking numberOfTasksInEachQueue just need 1 cache miss (putting into cpu cache)
	numberOfTasksInEachQueue []int

	// we also create separate queues for each priority
	// so it is simple to push/pop the item
	priorityQueues map[int]*linkedslice.LinkedSlice

	// simple metadata
	limitPriority             int
	size                      int
	sizeLimit                 int
	currentPriorityToRetrieve int
	runningFlag               int32
}

// ErrParamShouldBePositive is returned when either sizeLimit or priority parameter is negative
var ErrParamShouldBePositive = errors.New("sizeLimit and priority given should be positive")

// NewRoundRobinPriorityQueue creates our RR PQ.
//
// It caps at sizeLimit, and allows priorirty [0,numOfPriority)
func NewRoundRobinPriorityQueue(sizeLimit, numOfPriority int) (*RoundRobinPriorityQueue, error) {
	mu := &sync.Mutex{}
	notEmpty := sync.NewCond(mu)

	if sizeLimit <= 0 || numOfPriority <= 0 {
		return nil, ErrParamShouldBePositive
	}

	numberOfTasksInEachQueue := make([]int, numOfPriority)
	priorityQueues := make(map[int]*linkedslice.LinkedSlice)

	return &RoundRobinPriorityQueue{
		mu:                        mu,
		notEmpty:                  notEmpty,
		numberOfTasksInEachQueue:  numberOfTasksInEachQueue,
		priorityQueues:            priorityQueues,
		limitPriority:             numOfPriority,
		size:                      0,
		sizeLimit:                 sizeLimit,
		currentPriorityToRetrieve: -1,
		runningFlag:               1,
	}, nil
}

// ErrPriorityOutOfRange is returned if priority given is outside of range/
//
// If we accept it, to maintain the guarantee, needs to maintain too much queue,
// and hard to scan over.
var ErrPriorityOutOfRange = errors.New("Roundrobin Priority Queue is full, rejecting new qitem")

// PushOrError put the item into the rrpq, and returns error if no slot available
func (rr *RoundRobinPriorityQueue) PushOrError(item common.QItem) error {
	if running := atomic.LoadInt32(&rr.runningFlag); running == 0 {
		return common.ErrQueueIsClosed
	}
	if item.Priority < 0 || item.Priority >= rr.limitPriority {
		return ErrPriorityOutOfRange
	}

	rr.mu.Lock()

	// double check, ensuring see the changes after lock call
	if running := atomic.LoadInt32(&rr.runningFlag); running == 0 {
		return common.ErrQueueIsClosed
	}
	if item.Priority < 0 || item.Priority >= rr.limitPriority {
		return ErrPriorityOutOfRange
	}

	if rr.size == rr.sizeLimit {
		rr.mu.Unlock()
		return common.ErrQueueIsFull
	}
	if _, ok := rr.priorityQueues[item.Priority]; !ok {
		rr.priorityQueues[item.Priority] = linkedslice.NewLinkedSlice()
	}
	rr.priorityQueues[item.Priority].PushOrError(item)

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
	if running := atomic.LoadInt32(&rr.runningFlag); running == 0 {
		return common.MinQItem, common.ErrQueueIsClosed
	}

	rr.mu.Lock()

	// double check, ensuring see the changes after lock call
	if running := atomic.LoadInt32(&rr.runningFlag); running == 0 {
		rr.mu.Unlock()
		return common.MinQItem, common.ErrQueueIsClosed
	}

	for rr.size == 0 {
		rr.notEmpty.Wait()
		// double check, ensuring see the changes after wait call
		if running := atomic.LoadInt32(&rr.runningFlag); running == 0 {
			rr.mu.Unlock()
			return common.MinQItem, common.ErrQueueIsClosed
		}
	}

	// if we wait blindly, it gonna stuck
	// but we are tracking it manually, ensuring it will never wait
	qitem, err := rr.priorityQueues[rr.currentPriorityToRetrieve].PopOrWaitTillClose()
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
		if newPos == -1 {
			for i := rr.limitPriority - 1; i > rr.currentPriorityToRetrieve; i-- {
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
	atomic.CompareAndSwapInt32(&rr.runningFlag, 1, 0)
	rr.notEmpty.Broadcast()
	for i := 0; i < rr.limitPriority; i++ {
		if _, ok := rr.priorityQueues[i]; ok {
			rr.priorityQueues[i].Close()
		}
	}
}
