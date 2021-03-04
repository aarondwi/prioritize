package prioritize

import (
	"context"
	"sync"
)

// TaskFunc is our interface, to be implemented by user
type TaskFunc func(context.Context, interface{}) (interface{}, error)

// Task is the main object that prioritize schedules.
// It is is basically a `promise` implementation.
type Task struct {
	ctx      context.Context
	priority int
	fn       TaskFunc
	arg      interface{}
	wg       *sync.WaitGroup
	result   interface{}
	err      error
}

// newTask creates a prioritize.Task object with the given parameter
func newTask(
	ctx context.Context,
	priority int,
	fn TaskFunc,
	arg interface{}) *Task {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	return &Task{
		ctx:      ctx,
		priority: priority,
		fn:       fn,
		arg:      arg,
		wg:       wg,
		result:   nil,
		err:      nil,
	}
}

func (t *Task) set(result interface{}, err error) {
	t.result = result
	t.err = err
	t.wg.Done()
}

// Result waits until the Task object completes
func (t *Task) Result() (interface{}, error) {
	t.wg.Wait()
	if t.err != nil {
		return nil, t.err
	}
	return t.result, nil
}
