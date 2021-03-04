package roundrobin

import (
	"log"
	"testing"
	"time"

	"github.com/aarondwi/prioritize/common"
)

func TestRoundRobinPriorityQueue(t *testing.T) {
	rr, err := NewRoundRobinPriorityQueue(2048, 16)
	if err != nil {
		t.Fatalf("It should not error, cause both are positive, but we got %v", err)
	}

	err = rr.PushOrError(common.QItem{ID: 1, Priority: 8})
	if err != nil {
		t.Fatalf("It should not return error, cause not full yet, but we got %v", err)
	}

	err = rr.PushOrError(common.QItem{ID: 2, Priority: 13})
	if err != nil {
		t.Fatalf("It should not return error, cause not full yet, but we got %v", err)
	}

	err = rr.PushOrError(common.QItem{ID: 3, Priority: 5})
	if err != nil {
		t.Fatalf("It should not return error, cause not full yet, but we got %v", err)
	}

	result := rr.PopOrWait()
	if result.ID != 1 || result.Priority != 8 {
		t.Fatalf("First item should be returned first, but instead we got %v", result)
	}

	result = rr.PopOrWait()
	if result.ID != 3 || result.Priority != 5 {
		t.Fatalf("Left-hand side of first item should be returned first, but instead we got %v", result)
	}

	result = rr.PopOrWait()
	if result.ID != 2 || result.Priority != 13 {
		t.Fatalf("After left-hand side, should roll back to higher one, but instead we got %v", result)
	}
}

func TestRoundRobinPriorityQueueValidation(t *testing.T) {
	_, err := NewRoundRobinPriorityQueue(-2048, 1)
	if err == nil {
		t.Fatal("It should error, cause sizeLimit can't be negative, but it is not")
	}

	_, err = NewRoundRobinPriorityQueue(2048, -16)
	if err == nil {
		t.Fatal("It should error, cause numOfPriority can't be negative, but it is not")
	}

	rrpq, err := NewRoundRobinPriorityQueue(2048, 16)
	if err != nil {
		t.Fatalf("It should not error, instead we got %v", err)
	}

	err = rrpq.PushOrError(common.QItem{Priority: -1})
	if err == nil || err != ErrPriorityOutOfRange {
		t.Fatal("It should error, cause cannot accept negative priority, but it is not")
	}

	err = rrpq.PushOrError(common.QItem{Priority: 16})
	if err == nil || err != ErrPriorityOutOfRange {
		t.Fatal("It should error, cause can only accept priority [0, numOfPriority), but it is not")
	}

	if rrpq.size != 0 {
		t.Fatalf("No item is added yet, but the size is %d", rrpq.size)
	}

	for i := 0; i < 2048; i++ {
		err = rrpq.PushOrError(
			common.QItem{ID: uint64(i), Priority: i % 16})
		if err != nil {
			t.Fatalf("It should not error, because slots left, but instead, at iteration %d, size %d, sizeLimit %d, we got %v", i, rrpq.size, rrpq.sizeLimit, err)
		}
	}

	err = rrpq.PushOrError(common.QItem{ID: 2048, Priority: 1})
	if err == nil {
		t.Fatalf("It should error, because no slots left, but it is not")
	}
}

func TestRoundRobinPriorityQueuePopWait(t *testing.T) {
	rrpq, err := NewRoundRobinPriorityQueue(100, 16)

	c := make(chan bool, 1)
	go func() {
		time.Sleep(200 * time.Millisecond)
		log.Println("timeout, returning")
		c <- false
	}()

	go func() {
		item := rrpq.PopOrWait()
		if item.Priority != 10 {
			log.Printf("We received priority %d\n", item.Priority)
			c <- false
		}
		c <- true
	}()

	time.Sleep(100 * time.Millisecond)
	err = rrpq.PushOrError(common.QItem{Priority: 10})
	if err != nil {
		t.Fatalf("It should not error because slots are available, but we got %v", err)
	}

	result := <-c
	if !result {
		t.Fatal("We should receive true, because all the above are true, but we are not")
	}
}

func BenchmarkRoundRobinPriorityQueue(b *testing.B) {
	rrpq, _ := NewRoundRobinPriorityQueue(1024, 8)

	for i := 0; i < b.N; i++ {
		for j := 0; j < 1024; j++ {
			err := rrpq.PushOrError(
				common.QItem{ID: uint64(j), Priority: j % 8})
			if err != nil {
				b.Fatalf("It should not error, because slots left, but instead, at iteration %d we got %v", j, err)
			}
		}
		for j := 0; j < 1024; j++ {
			rrpq.PopOrWait()
		}
	}
}
