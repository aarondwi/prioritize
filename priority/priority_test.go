package priority

import (
	"log"
	"runtime"
	"testing"
	"time"

	"github.com/aarondwi/prioritize/common"
)

func TestPriorityQueue(t *testing.T) {
	fq, err := NewPriorityQueue(2048, 16)
	if err != nil {
		t.Fatalf("It should not error, cause both are positive, but we got %v", err)
	}

	err = fq.PushOrError(common.QItem{ID: 1, Priority: 8})
	if err != nil {
		t.Fatalf("It should not return error, cause not full yet, but we got %v", err)
	}

	err = fq.PushOrError(common.QItem{ID: 2, Priority: 13})
	if err != nil {
		t.Fatalf("It should not return error, cause not full yet, but we got %v", err)
	}

	err = fq.PushOrError(common.QItem{ID: 3, Priority: 13})
	if err != nil {
		t.Fatalf("It should not return error, cause not full yet, but we got %v", err)
	}

	result, err := fq.PopOrWaitTillClose()
	if err != nil {
		t.Fatalf("It should not error, cause not closed yet, but we got %v", err)
	}
	if result.ID != 2 || result.Priority != 13 {
		t.Fatalf("First item with highest priority should be returned first, but instead we got %v", result)
	}

	result, err = fq.PopOrWaitTillClose()
	if err != nil {
		t.Fatalf("It should not error, cause not closed yet, but we got %v", err)
	}
	if result.ID != 3 || result.Priority != 13 {
		t.Fatalf("Item with higher priority should be returned first, but instead we got %v", result)
	}

	result, err = fq.PopOrWaitTillClose()
	if err != nil {
		t.Fatalf("It should not error, cause not closed yet, but we got %v", err)
	}
	if result.ID != 1 || result.Priority != 8 {
		t.Fatalf("Should not error cause we still have 1 item, but instead we got %v", result)
	}
	fq.Close()
}

func TestPriorityQueueValidation(t *testing.T) {
	_, err := NewPriorityQueue(-2048, 1)
	if err == nil || err != common.ErrParamShouldBePositive {
		t.Fatal("It should error, cause sizeLimit can't be negative, but it is not")
	}

	_, err = NewPriorityQueue(2048, -16)
	if err == nil || err != common.ErrParamShouldBePositive {
		t.Fatal("It should error, cause numOfPriority can't be negative, but it is not")
	}

	pq, err := NewPriorityQueue(2048, 16)
	if err != nil {
		t.Fatalf("It should not error, instead we got %v", err)
	}

	err = pq.PushOrError(common.QItem{Priority: -1})
	if err == nil || err != common.ErrPriorityOutOfRange {
		t.Fatal("It should error, cause cannot accept negative priority, but it is not")
	}

	err = pq.PushOrError(common.QItem{Priority: 16})
	if err == nil || err != common.ErrPriorityOutOfRange {
		t.Fatal("It should error, cause can only accept priority [0, numOfPriority), but it is not")
	}

	if pq.size != 0 {
		t.Fatalf("No item is added yet, but the size is %d", pq.size)
	}

	for i := 0; i < 2048; i++ {
		err = pq.PushOrError(
			common.QItem{ID: uint64(i), Priority: i % 16})
		if err != nil {
			t.Fatalf("It should not error, because slots left, but instead, at iteration %d, size %d, sizeLimit %d, we got %v", i, pq.size, pq.sizeLimit, err)
		}
	}

	err = pq.PushOrError(common.QItem{ID: 2048, Priority: 1})
	if err == nil {
		t.Fatalf("It should error, because no slots left, but it is not")
	}

	pq.Close()
}

func TestPriorityQueuePopWait(t *testing.T) {
	pq, err := NewPriorityQueue(100, 16)

	c := make(chan bool, 1)
	go func() {
		time.Sleep(200 * time.Millisecond)
		log.Println("timeout, returning")
		c <- false
	}()

	go func() {
		item, err := pq.PopOrWaitTillClose()
		if err != nil {
			c <- false
			return
		}
		if item.Priority != 10 {
			log.Printf("We received priority %d\n", item.Priority)
			c <- false
			return
		}
		c <- true
	}()

	time.Sleep(100 * time.Millisecond)
	err = pq.PushOrError(common.QItem{Priority: 10})
	if err != nil {
		t.Fatalf("It should not error because slots are available, but we got %v", err)
	}

	result := <-c
	if !result {
		t.Fatal("We should receive true, because all the above are true, but we are not")
	}
	pq.Close()
}

func TestPriorityQueueAfterClose(t *testing.T) {
	pq, _ := NewPriorityQueue(2000, 8)
	pq.Close()

	err := pq.PushOrError(common.QItem{})
	if err == nil || err != common.ErrQueueIsClosed {
		t.Fatalf("It should be error, cause already closed, but it is not")
	}

	_, err = pq.PopOrWaitTillClose()
	if err == nil || err != common.ErrQueueIsClosed {
		t.Fatalf("It should be error, cause already closed, but it is not")
	}
}

func BenchmarkPriorityQueue(b *testing.B) {
	pq, _ := NewPriorityQueue(1024, 8)
	for i := 0; i < b.N; i++ {
		pq.PushOrError(
			common.QItem{ID: uint64(i), Priority: i % 8})
		pq.PopOrWaitTillClose()
	}
	pq.Close()
}

func BenchmarkPriorityQueueInLoop(b *testing.B) {
	pq, _ := NewPriorityQueue(1024, 8)
	for i := 0; i < b.N; i++ {
		for j := 0; j < 128; j++ {
			pq.PushOrError(
				common.QItem{ID: uint64(j), Priority: j % 8})
		}
		for j := 0; j < 128; j++ {
			pq.PopOrWaitTillClose()
		}
	}
	pq.Close()
}

func BenchmarkPriorityQueueParallelOneCoreOnly(b *testing.B) {
	pq, _ := NewPriorityQueue(1024, 8)
	runtime.GOMAXPROCS(1)
	b.RunParallel(func(pb *testing.PB) {
		j := 0
		for pb.Next() {
			j++
			pq.PushOrError(
				common.QItem{ID: uint64(j), Priority: j % 8})
			pq.PopOrWaitTillClose()
		}
	})
	pq.Close()
}

func BenchmarkPriorityQueueInLoopParallelOneCoreOnly(b *testing.B) {
	pq, _ := NewPriorityQueue(1024, 8)
	runtime.GOMAXPROCS(1)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for j := 0; j < 128; j++ {
				pq.PushOrError(
					common.QItem{ID: uint64(j), Priority: j % 8})
			}
			for j := 0; j < 128; j++ {
				pq.PopOrWaitTillClose()
			}
		}
	})
	pq.Close()
}

func BenchmarkPriorityQueueParallel(b *testing.B) {
	pq, _ := NewPriorityQueue(1024, 8)
	b.RunParallel(func(pb *testing.PB) {
		j := 0
		for pb.Next() {
			j++
			pq.PushOrError(
				common.QItem{ID: uint64(j), Priority: j % 8})
			pq.PopOrWaitTillClose()
		}
	})
	pq.Close()
}

func BenchmarkPriorityQueueInLoopParallel(b *testing.B) {
	pq, _ := NewPriorityQueue(1024, 8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for j := 0; j < 128; j++ {
				pq.PushOrError(
					common.QItem{ID: uint64(j), Priority: j % 8})
			}
			for j := 0; j < 128; j++ {
				pq.PopOrWaitTillClose()
			}
		}
	})
	pq.Close()
}
