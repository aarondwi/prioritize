package prioritize

import (
	"log"
	"testing"
	"time"
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
		err := q.PushOrError(QItem{priority: item})
		if err != nil {
			t.Fatal("It should not error because slots still available, but it is")
		}
	}
	err := q.PushOrError(QItem{priority: 1})
	if err == nil {
		t.Fatal("It should fail because already full, but it is not")
	}

	for i := 0; i < 100; i++ {
		item := q.PopOrWait()
		if item.priority != descSortedArr[i] {
			t.Fatalf(
				"It should be the same, cause both descending sorted, but instead we got %d and %d",
				item.priority, descSortedArr[i])
		}
	}

	// re-Push, to ensure it works till full again
	for _, item := range randomizedArr {
		err := q.PushOrError(QItem{priority: item})
		if err != nil {
			t.Fatal("It should not error because slots still available, but it is")
		}
	}

	// and still sorted after that
	for i := 0; i < 100; i++ {
		item := q.PopOrWait()
		if item.priority != descSortedArr[i] {
			t.Fatalf(
				"It should be the same, cause both descending sorted, but instead we got %d and %d",
				item.priority, descSortedArr[i])
		}
	}
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
		item := q.PopOrWait()
		if item.priority != 100 {
			log.Printf("We received priority %d", item.priority)
			c <- false
		}
		c <- true
	}()

	time.Sleep(100 * time.Millisecond)
	err := q.PushOrError(QItem{priority: 100})
	if err != nil {
		t.Fatalf("It should not error because slots are available, but we got %v", err)
	}

	result := <-c
	if !result {
		t.Fatal("We should receive true, because all the above are true, but we are not")
	}
}
