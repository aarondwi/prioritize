package heap

import (
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/aarondwi/prioritize/common"
)

func swap(arr []int, i1, i2 int) {
	// no need to check error
	// our test code guaranteed this succeeds
	//
	// this is just randomize tool
	temp := arr[i1]
	arr[i1] = arr[i2]
	arr[i2] = temp
}

func TestHeapMainFlow(t *testing.T) {
	descSortedArr := make([]int, 100)
	randomizedArr := make([]int, 100)
	for i := 0; i < 100; i++ {
		descSortedArr[i] = 100 - i
		randomizedArr[i] = 100 - i
	}
	for i := 0; i < 50; i++ {
		swap(randomizedArr, i, i+50)
	}

	q := NewHeapPriorityQueue(100)
	for _, item := range randomizedArr {
		err := q.PushOrError(common.QItem{Priority: item})
		if err != nil {
			t.Fatal("It should not error because slots still available, but it is")
		}
	}
	err := q.PushOrError(common.QItem{Priority: 1})
	if err == nil {
		t.Fatal("It should fail because already full, but it is not")
	}

	for i := 0; i < 100; i++ {
		item, err := q.PopOrWaitTillClose()
		if err != nil {
			t.Fatalf("it should not error, cause not closed yet, but we got %v", err)
		}
		if item.Priority != descSortedArr[i] {
			t.Fatalf(
				"It should be the same, cause both descending sorted, but instead we got %d and %d",
				item.Priority, descSortedArr[i])
		}
	}

	// re-Push, to ensure it works till full again
	for _, item := range randomizedArr {
		err := q.PushOrError(common.QItem{Priority: item})
		if err != nil {
			t.Fatal("It should not error because slots still available, but it is")
		}
	}

	// and still sorted after that
	for i := 0; i < 100; i++ {
		item, err := q.PopOrWaitTillClose()
		if err != nil {
			if err != nil {
				t.Fatalf("it should not error, cause not closed yet, but we got %v", err)
			}
		}
		if item.Priority != descSortedArr[i] {
			t.Fatalf(
				"It should be the same, cause both descending sorted, but instead we got %d and %d",
				item.Priority, descSortedArr[i])
		}
	}

	q.Close()
}

func TestPopWait(t *testing.T) {
	q := NewHeapPriorityQueue(100)

	c := make(chan bool, 1)
	go func() {
		time.Sleep(200 * time.Millisecond)
		log.Println("timeout, returning")
		c <- false
	}()

	go func() {
		item, err := q.PopOrWaitTillClose()
		if err != nil {
			c <- false
			return
		}
		if item.Priority != 100 {
			log.Printf("We received priority %d\n", item.Priority)
			c <- false
			return
		}
		c <- true
	}()

	time.Sleep(100 * time.Millisecond)
	err := q.PushOrError(common.QItem{Priority: 100})
	if err != nil {
		t.Fatalf("It should not error because slots are available, but we got %v", err)
	}

	result := <-c
	if !result {
		t.Fatal("We should receive true, because all the above are true, but we are not")
	}

	q.Close()
}

func TestHeapAfterClose(t *testing.T) {
	q := NewHeapPriorityQueue(100)
	q.Close()

	err := q.PushOrError(common.QItem{})
	if err == nil || err != common.ErrQueueIsClosed {
		t.Fatalf("It should be error, cause already closed, but it is not")
	}

	_, err = q.PopOrWaitTillClose()
	if err == nil || err != common.ErrQueueIsClosed {
		t.Fatalf("It should be error, cause already closed, but it is not")
	}
}

func BenchmarkHeapPQ(b *testing.B) {
	q := NewHeapPriorityQueue(1024)

	for i := 0; i < b.N; i++ {
		for j := 0; j < 128; j++ {
			q.PushOrError(common.QItem{Priority: rand.Intn(64)})
		}
		for j := 0; j < 128; j++ {
			q.PopOrWaitTillClose()
		}
	}

	q.Close()
}

func BenchmarkHeapPQParallel(b *testing.B) {
	q := NewHeapPriorityQueue(1024 * 2)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < 128; i++ {
				q.PushOrError(common.QItem{Priority: rand.Intn(64)})
			}
			for i := 0; i < 128; i++ {
				q.PopOrWaitTillClose()
			}
		}
	})
	q.Close()
}
