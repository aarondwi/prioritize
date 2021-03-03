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
		res := cq.PopOrWait()
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
		res := cq.PopOrWait().ID
		if res != result[i] {
			t.Fatalf("Phase 3: It should be ok because still has remaining items, but it is not")
		}
	}
}

func TestCircularQueuePopWait(t *testing.T) {
	q := NewCircularQueue(100)

	c := make(chan bool, 1) // prevent goroutine leak
	go func() {
		time.Sleep(200 * time.Millisecond)
		log.Println("timeout, returning")
		c <- false
	}()

	go func() {
		item := q.PopOrWait()
		if item.ID != 100 {
			log.Printf("We received priority %d\n", item.Priority)
			c <- false
		}
		c <- true
	}()

	time.Sleep(100 * time.Millisecond)
	err := q.PushOrError(common.QItem{ID: 100})
	if err != nil {
		t.Fatalf("It should not error because slots are available, but we got %v", err)
	}

	result := <-c
	if !result {
		t.Fatal("We should receive true, because all the above are true, but we are not")
	}
}

func BenchmarkCircularQueue(b *testing.B) {
	cq := NewCircularQueue(100)
	for i := 0; i < b.N; i++ {
		for j := 0; j < 10; j++ {
			err := cq.PushOrError(common.QItem{ID: uint64(j)})
			if err != nil {
				b.Fatalf("On Push, should always be ok, but it is not, on iteration %d -> %d", i, j)
			}
		}
		for j := 0; j < 10; j++ {
			cq.PopOrWait()
		}
	}
}
