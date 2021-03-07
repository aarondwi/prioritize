package circular

import (
	"log"
	"testing"
	"time"

	"github.com/aarondwi/prioritize/common"
)

func TestCircularQueue(t *testing.T) {
	cq := NewCircularQueue(5)

	// PHASE 1
	for i := 0; i < 4; i++ {
		err := cq.PushOrError(common.QItem{ID: uint64(i)})
		if err != nil {
			t.Fatalf("Phase 1: It should be ok because still slots left, but it is not, and we got %v", err)
		}
	}

	for i := 0; i < 2; i++ {
		res, err := cq.PopOrWaitTillClose()
		if err != nil {
			t.Fatalf("It should not error, given still have items, but we got %v", err)
		}
		if res.ID != uint64(i) {
			t.Fatalf("Phase 1: It should be the same, because same order, but instead we got %d and %d", res, uint64(i))
		}
	}

	// PHASE 2
	for i := 0; i < 3; i++ {
		err := cq.PushOrError(common.QItem{ID: uint64(10 + i)})
		if err != nil {
			t.Fatalf("Phase 2: It should be ok because there are slots left, but it is not")
		}
	}

	err := cq.PushOrError(common.QItem{ID: 100})
	if err == nil {
		t.Fatalf("Phase 2: It should fail because no slots left, but it does not")
	}

	// PHASE 3
	result := []uint64{2, 3, 10, 11, 12}
	for i := 0; i < 5; i++ {
		res, err := cq.PopOrWaitTillClose()
		if err != nil {
			t.Fatalf("It should not error, given still have items, but we got %v", err)
		}
		if res.ID != result[i] {
			t.Fatalf("Phase 3: It should be ok because still has remaining items, but it is not")
		}
	}

	cq.Close()
}

func TestCircularQueuePopWait(t *testing.T) {
	cq := NewCircularQueue(100)

	c := make(chan bool, 1) // prevent goroutine leak
	go func() {
		time.Sleep(200 * time.Millisecond)
		log.Println("timeout, returning")
		c <- false
	}()

	go func() {
		item, err := cq.PopOrWaitTillClose()
		if err != nil {
			c <- false
			return
		}
		if item.ID != 100 {
			log.Printf("We received priority %d\n", item.Priority)
			c <- false
			return
		}
		c <- true
	}()

	time.Sleep(100 * time.Millisecond)
	err := cq.PushOrError(common.QItem{ID: 100})
	if err != nil {
		t.Fatalf("It should not error because slots are available, but we got %v", err)
	}

	result := <-c
	if !result {
		t.Fatal("We should receive true, because all the above are true, but we are not")
	}

	cq.Close()
}

func TestCircularQueueAfterClose(t *testing.T) {
	cq := NewCircularQueue(100)
	cq.Close()

	err := cq.PushOrError(common.QItem{})
	if err == nil || err != common.ErrQueueIsClosed {
		t.Fatalf("It should be error, cause already closed, but it is not")
	}

	_, err = cq.PopOrWaitTillClose()
	if err == nil || err != common.ErrQueueIsClosed {
		t.Fatalf("It should be error, cause already closed, but it is not")
	}
}

func BenchmarkCircularQueue(b *testing.B) {
	cq := NewCircularQueue(128)
	for i := 0; i < b.N; i++ {
		for j := 0; j < 128; j++ {
			cq.PushOrError(common.QItem{ID: uint64(j)})
		}
		for j := 0; j < 128; j++ {
			cq.PopOrWaitTillClose()
		}
	}
	cq.Close()
}

func BenchmarkCircularQueueParallel(b *testing.B) {
	cq := NewCircularQueue(1024 * 2)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for j := 0; j < 128; j++ {
				cq.PushOrError(common.QItem{ID: uint64(j)})
			}
			for j := 0; j < 128; j++ {
				cq.PopOrWaitTillClose()
			}
		}
	})
	cq.Close()
}
