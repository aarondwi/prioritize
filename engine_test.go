package prioritize

import (
	"context"
	"testing"

	"github.com/aarondwi/prioritize/heap"
)

func TestPrioritizeEngine(t *testing.T) {
	engine, err := New(heap.NewHeapPriorityQueue(100), 5)
	if err != nil {
		t.Fatalf("It should not error, because all are correct parameters, instead we got %v", err)
	}

	val := 1
	fn := func(ctx context.Context, arg interface{}) (interface{}, error) {
		return val + 1, nil
	}

	task, err := engine.Submit(
		context.Background(), 1, fn, nil)

	result, err := task.Result()
	if err != nil {
		t.Fatalf("It should be nil, because we return so, but it is not")
	}
	if result.(int) != 2 {
		t.Fatalf("Expected 2, received %d", result.(int))
	}

	engine.Close()
}

func TestPriorityEngineCtxFinished(t *testing.T) {
	engine, err := New(heap.NewHeapPriorityQueue(100), 5)
	if err != nil {
		t.Fatalf("It should not error, because all are correct parameters, instead we got %v", err)
	}

	val := 1
	fn := func(ctx context.Context, arg interface{}) (interface{}, error) {
		return val + 1, nil
	}

	ctxCancelled, cancelFunc := context.WithCancel(
		context.Background())
	cancelFunc()
	task, err := engine.Submit(ctxCancelled, 1, fn, nil)

	_, err = task.Result()
	if err == nil || err != ErrCtxAlreadyCancelled {
		t.Fatalf("It should not be nil, because context already cancelled, instead we got %v", err)
	}

	engine.Close()
}

func TestSubmitCallAfterClose(t *testing.T) {
	engine, err := New(heap.NewHeapPriorityQueue(100), 5)
	if err != nil {
		t.Fatalf("It should not error, because all are correct parameters, instead we got %v", err)
	}

	val := 1
	fn := func(ctx context.Context, arg interface{}) (interface{}, error) {
		return val + 1, nil
	}
	engine.Close()

	_, err = engine.Submit(context.Background(), 1, fn, nil)

	if err == nil || err != ErrAlreadyClosed {
		t.Fatalf("It should not be nil, because context already cancelled, instead we got %v", err)
	}
}
