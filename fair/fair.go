package fair

import (
	"sync"

	"github.com/aarondwi/prioritize/common"
	"github.com/aarondwi/prioritize/linkedslice"
)

// FairQueue is a queue in which
// each priority gets a chance to return value,
// starting from first item put going downwards,
// and then rolled back from highest.
//
// This behavior allows some starvation prevention for lower priorities,
// assuming that highest priority tasks have much lower number of tasks,
// else it gonna be pretty much not useful, just like random/normal queue.
//
// Internally, we are using unbounded linkedslice.
// But because we need size limits, we track it here
type FairQueue struct {
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

// NewFairQueue creates our fair queue.
//
// It caps at sizeLimit, and allows priorirty [0,numOfPriority)
func NewFairQueue(sizeLimit, numOfPriority int) (*FairQueue, error) {
	if sizeLimit <= 0 || numOfPriority <= 0 {
		return nil, common.ErrParamShouldBePositive
	}

	mu := &sync.Mutex{}
	notEmpty := sync.NewCond(mu)

	numberOfTasksInEachQueue := make([]int, numOfPriority)
	queues := make([]*linkedslice.LinkedSlice, numOfPriority)

	return &FairQueue{
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
func (fq *FairQueue) PushOrError(item common.QItem) error {
	if item.Priority < 0 || item.Priority >= fq.limitPriority {
		return common.ErrPriorityOutOfRange
	}

	fq.mu.Lock()
	if !fq.running {
		fq.mu.Unlock()
		return common.ErrQueueIsClosed
	}

	if fq.size == fq.sizeLimit {
		fq.mu.Unlock()
		return common.ErrQueueIsFull
	}
	if fq.queues[item.Priority] == nil {
		fq.queues[item.Priority] = linkedslice.NewLinkedSlice()
	}
	err := fq.queues[item.Priority].PushOrError(item)
	if err != nil {
		// meaning already closed, cause linkedslices is unbounded
		fq.mu.Unlock()
		return err
	}

	// The only item in the queue, set this to position
	if fq.size == 0 {
		fq.currentPriorityToRetrieve = item.Priority
	}

	// update the tracker too
	fq.numberOfTasksInEachQueue[item.Priority]++
	fq.size++
	fq.notEmpty.Signal()

	fq.mu.Unlock()

	return nil
}

// PopOrWaitTillClose returns 1 QItem from RRPQ, or waits if none exists
func (fq *FairQueue) PopOrWaitTillClose() (common.QItem, error) {
	fq.mu.Lock()
	if !fq.running {
		fq.mu.Unlock()
		return common.MinQItem, common.ErrQueueIsClosed
	}

	for fq.size == 0 {
		fq.notEmpty.Wait()
		// double check, ensuring see the changes after wait call
		if !fq.running {
			fq.mu.Unlock()
			return common.MinQItem, common.ErrQueueIsClosed
		}
	}

	// if we wait blindly, it gonna stuck
	// but we are tracking it manually, ensuring it will never wait
	qitem, err := fq.queues[fq.currentPriorityToRetrieve].PopOrWaitTillClose()
	if err != nil {
		// the only error possible here is closed already
		// so we just continue it
		fq.mu.Unlock()
		return common.MinQItem, err
	}
	result := common.QItem{
		ID:       qitem.ID,
		Priority: fq.currentPriorityToRetrieve,
	}
	fq.numberOfTasksInEachQueue[fq.currentPriorityToRetrieve]--
	fq.size--

	if fq.size == 0 {
		//fast path, no need to check rr.numberOfTasksInEachQueue
		fq.currentPriorityToRetrieve = -1
	} else {
		// Check new rr.currentPosToRetrieve position, cause we still have item somewhere
		newPos := -1
		for i := fq.currentPriorityToRetrieve - 1; i >= 0; i-- {
			if fq.numberOfTasksInEachQueue[i] > 0 {
				newPos = i
				break
			}
		}
		// not yet found, meaning remaining items reside on higher index
		// currentPriorityToRetrieve should be the last index to be checked
		if newPos == -1 {
			for i := fq.limitPriority - 1; i >= fq.currentPriorityToRetrieve; i-- {
				if fq.numberOfTasksInEachQueue[i] > 0 {
					newPos = i
					break
				}
			}
		}
		fq.currentPriorityToRetrieve = newPos
	}

	fq.mu.Unlock()
	return result, nil
}

// Close FairQueue, preventing it from accepting new request
func (fq *FairQueue) Close() {
	fq.mu.Lock()
	fq.running = false
	for i := 0; i < fq.limitPriority; i++ {
		if fq.queues[i] != nil {
			fq.queues[i].Close()
		}
	}
	fq.notEmpty.Broadcast()
	fq.mu.Unlock()
}
