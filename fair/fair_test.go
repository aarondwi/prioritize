package fair

import (
	"log"
	"runtime"
	"testing"
	"time"

	"github.com/aarondwi/prioritize/common"
)

func TestFairQueue(t *testing.T) {
	fq, err := NewFairQueue(2048, 16)
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

	err = fq.PushOrError(common.QItem{ID: 3, Priority: 5})
	if err != nil {
		t.Fatalf("It should not return error, cause not full yet, but we got %v", err)
	}

	err = fq.PushOrError(common.QItem{ID: 4, Priority: 13})
	if err != nil {
		t.Fatalf("It should not return error, cause not full yet, but we got %v", err)
	}

	result, err := fq.PopOrWaitTillClose()
	if err != nil {
		t.Fatalf("It should not error, cause not closed yet, but we got %v", err)
	}
	if result.ID != 1 || result.Priority != 8 {
		t.Fatalf("First item should be returned first, but instead we got %v", result)
	}

	result, err = fq.PopOrWaitTillClose()
	if err != nil {
		t.Fatalf("It should not error, cause not closed yet, but we got %v", err)
	}
	if result.ID != 3 || result.Priority != 5 {
		t.Fatalf("Left-hand side of first item should be returned first, but instead we got %v", result)
	}

	result, err = fq.PopOrWaitTillClose()
	if err != nil {
		t.Fatalf("It should not error, cause not closed yet, but we got %v", err)
	}
	if result.ID != 2 || result.Priority != 13 {
		t.Fatalf("After left-hand side, should roll back to higher one, but instead we got %v", result)
	}

	result, err = fq.PopOrWaitTillClose()
	if err != nil {
		t.Fatalf("It should not error, cause not closed yet, but we got %v", err)
	}
	if result.ID != 4 || result.Priority != 13 {
		t.Fatalf("After left-hand side, should roll back to higher one, but instead we got %v", result)
	}
	fq.Close()
}

func TestFairQueueValidation(t *testing.T) {
	_, err := NewFairQueue(-2048, 1)
	if err == nil || err != common.ErrParamShouldBePositive {
		t.Fatal("It should error, cause sizeLimit can't be negative, but it is not")
	}

	_, err = NewFairQueue(2048, -16)
	if err == nil || err != common.ErrParamShouldBePositive {
		t.Fatal("It should error, cause numOfPriority can't be negative, but it is not")
	}

	fq, err := NewFairQueue(2048, 16)
	if err != nil {
		t.Fatalf("It should not error, instead we got %v", err)
	}

	err = fq.PushOrError(common.QItem{Priority: -1})
	if err == nil || err != common.ErrPriorityOutOfRange {
		t.Fatal("It should error, cause cannot accept negative priority, but it is not")
	}

	err = fq.PushOrError(common.QItem{Priority: 16})
	if err == nil || err != common.ErrPriorityOutOfRange {
		t.Fatal("It should error, cause can only accept priority [0, numOfPriority), but it is not")
	}

	if fq.size != 0 {
		t.Fatalf("No item is added yet, but the size is %d", fq.size)
	}

	for i := 0; i < 2048; i++ {
		err = fq.PushOrError(
			common.QItem{ID: uint64(i), Priority: i % 16})
		if err != nil {
			t.Fatalf("It should not error, because slots left, but instead, at iteration %d, size %d, sizeLimit %d, we got %v", i, fq.size, fq.sizeLimit, err)
		}
	}

	err = fq.PushOrError(common.QItem{ID: 2048, Priority: 1})
	if err == nil {
		t.Fatalf("It should error, because no slots left, but it is not")
	}

	fq.Close()
}

func TestFairQueuePopWait(t *testing.T) {
	fq, err := NewFairQueue(100, 16)

	c := make(chan bool, 1)
	go func() {
		time.Sleep(200 * time.Millisecond)
		log.Println("timeout, returning")
		c <- false
	}()

	go func() {
		item, err := fq.PopOrWaitTillClose()
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
	err = fq.PushOrError(common.QItem{Priority: 10})
	if err != nil {
		t.Fatalf("It should not error because slots are available, but we got %v", err)
	}

	result := <-c
	if !result {
		t.Fatal("We should receive true, because all the above are true, but we are not")
	}
	fq.Close()
}

func TestFairQueueAfterClose(t *testing.T) {
	fq, _ := NewFairQueue(2000, 8)
	fq.Close()

	err := fq.PushOrError(common.QItem{})
	if err == nil || err != common.ErrQueueIsClosed {
		t.Fatalf("It should be error, cause already closed, but it is not")
	}

	_, err = fq.PopOrWaitTillClose()
	if err == nil || err != common.ErrQueueIsClosed {
		t.Fatalf("It should be error, cause already closed, but it is not")
	}
}

func BenchmarkFairQueue(b *testing.B) {
	fq, _ := NewFairQueue(1024, 8)
	for i := 0; i < b.N; i++ {
		fq.PushOrError(
			common.QItem{ID: uint64(i), Priority: i % 8})
		fq.PopOrWaitTillClose()
	}
	fq.Close()
}

func BenchmarkFairQueueInLoop(b *testing.B) {
	fq, _ := NewFairQueue(1024, 8)
	for i := 0; i < b.N; i++ {
		for j := 0; j < 128; j++ {
			fq.PushOrError(
				common.QItem{ID: uint64(j), Priority: j % 8})
		}
		for j := 0; j < 128; j++ {
			fq.PopOrWaitTillClose()
		}
	}
	fq.Close()
}

func BenchmarkFairQueueParallelOneCoreOnly(b *testing.B) {
	fq, _ := NewFairQueue(1024, 8)
	runtime.GOMAXPROCS(1)
	b.RunParallel(func(pb *testing.PB) {
		j := 0
		for pb.Next() {
			j++
			fq.PushOrError(
				common.QItem{ID: uint64(j), Priority: j % 8})
			fq.PopOrWaitTillClose()
		}
	})
	fq.Close()
}

func BenchmarkFairQueueInLoopParallelOneCoreOnly(b *testing.B) {
	fq, _ := NewFairQueue(1024, 8)
	runtime.GOMAXPROCS(1)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for j := 0; j < 128; j++ {
				fq.PushOrError(
					common.QItem{ID: uint64(j), Priority: j % 8})
			}
			for j := 0; j < 128; j++ {
				fq.PopOrWaitTillClose()
			}
		}
	})
	fq.Close()
}

func BenchmarkFairQueueParallel(b *testing.B) {
	fq, _ := NewFairQueue(1024, 8)
	b.RunParallel(func(pb *testing.PB) {
		j := 0
		for pb.Next() {
			j++
			fq.PushOrError(
				common.QItem{ID: uint64(j), Priority: j % 8})
			fq.PopOrWaitTillClose()
		}
	})
	fq.Close()
}

func BenchmarkFairQueueInLoopParallel(b *testing.B) {
	fq, _ := NewFairQueue(1024, 8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for j := 0; j < 128; j++ {
				fq.PushOrError(
					common.QItem{ID: uint64(j), Priority: j % 8})
			}
			for j := 0; j < 128; j++ {
				fq.PopOrWaitTillClose()
			}
		}
	})
	fq.Close()
}
