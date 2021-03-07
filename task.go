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
//
// I don't think, currently, exposing this to public is good idea.
// If it is published, I would be tempted to make `GetTask` and `PutTask` API,
// because pooling will make this library faster.
//
// But that also opens a bad chance for user to misuse the api (waiting for already-put Task, etc)
// which would make a lot more problem to explain.
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
