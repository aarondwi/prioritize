package prioritize

import (
	"context"
	"errors"
	"sync"

	"github.com/aarondwi/prioritize/common"
)

// Engine is our prioritizing engine.
// It has 3 parts: queue, worker, and mapping.
//
// Worker is designed as a goroutine pool,
// in which each will take an item from queue, get the task from the mapping,
// and then do the work
type Engine struct {
	sync.Mutex
	lastID    uint64
	q         common.QInterface
	mapping   map[uint64]*Task
	closeChan chan bool
}

// ErrNumOfWorkerIsNegativeOrZero is returned when `numOfWorker` parameter is <= 0
var ErrNumOfWorkerIsNegativeOrZero = errors.New("number of workers should be positive")

// ErrCtxAlreadyCancelled is returned when task.ctx taken by worker is already done
var ErrCtxAlreadyCancelled = errors.New("Context is already cancelled when it is gonna be taken")

// ErrAlreadyClosed is returned when `Submit()` is called after `Close()`
var ErrAlreadyClosed = errors.New("This engine is already closed")

// New creates our new prioritization engine.
func New(q common.QInterface, numOfWorker int) (*Engine, error) {
	if numOfWorker <= 0 {
		return nil, ErrNumOfWorkerIsNegativeOrZero
	}
	e := &Engine{
		q:         q,
		mapping:   make(map[uint64]*Task),
		closeChan: make(chan bool),
	}
	for i := 0; i < numOfWorker; i++ {
		go e.workLoop()
	}
	return e, nil
}

func (e *Engine) workLoop() {
	for {
		select {
		case <-e.closeChan:
			return
		default:
			// we need these to return by themselves.
			// because probably we already waiting on `PopOrWaitTillClose`
			// when closeChan is closed
			item, err := e.q.PopOrWaitTillClose()
			if err != nil {
				return
			}

			e.Lock()
			task, ok := e.mapping[item.ID]
			if !ok {
				panic("Broken implementation: ID not found in the mapping!")
			}
			delete(e.mapping, item.ID)
			e.Unlock()

			select {
			case <-task.ctx.Done():
				// fast path
				// already timeout/done, skip with error
				task.set(nil, ErrCtxAlreadyCancelled)
				break
			default:
				result, err := task.fn(task.ctx, task.arg)
				task.set(result, err)
				break
			}
		}
	}
}

// Submit creates task to be done in the worker goroutine
//
// The callee can call `.Result()` call to wait for result and error returned by fn
func (e *Engine) Submit(
	ctx context.Context,
	priority int,
	fn TaskFunc,
	arg interface{}) (*Task, error) {

	select {
	case <-e.closeChan:
		return nil, ErrAlreadyClosed
	default:
		e.Lock()

		// increment first
		// if crash/error, at most we lost 1 ID (out of 2^64, which basically is nothing)
		e.lastID++

		// Create mapping first.
		// Because we don't want race condition to happen between
		// fetching from queue and looking for the task to be run
		task := newTask(ctx, priority, fn, arg)
		e.mapping[e.lastID] = task

		err := e.q.PushOrError(common.QItem{ID: e.lastID, Priority: priority})
		if err != nil {
			delete(e.mapping, e.lastID)
			e.Unlock()
			return nil, err
		}
		e.Unlock()
		return task, nil
	}
}

// Close the instance, and all background goroutine worker
//
// Subsequent request will be rejected.
func (e *Engine) Close() {
	close(e.closeChan)
	e.q.Close()
}
