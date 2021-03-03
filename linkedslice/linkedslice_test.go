package linkedslice

import (
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/aarondwi/prioritize/common"
)

func TestLinkedSlice(t *testing.T) {
	ls := NewLinkedSlice()
	for i := 0; i < 1027; i++ {
		err := ls.PushOrError(common.QItem{ID: uint64(i)})
		if err != nil {
			t.Fatalf("This implementation will only return nil, but instead we got %v", err)
		}
	}
	for i := 0; i < 1027; i++ {
		res := ls.PopOrWait()
		if res.ID != uint64(i) {
			t.Fatalf("We don't receive FIFO as we expected: expected %d, got %d", uint64(i), res.ID)
		}
	}
}

func TestLinkedSlicePopWait(t *testing.T) {
	q := NewLinkedSlice()

	c := make(chan bool, 1)
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

func BenchmarkLinkedSlice(b *testing.B) {
	ls := NewLinkedSlice()

	for i := 0; i < b.N; i++ {
		r := rand.Intn(2048) + 1
		for j := 0; j < r; j++ {
			err := ls.PushOrError(common.QItem{ID: uint64(i*2048 + j)})
			if err != nil {
				b.Fatalf("it should never error because it is unbounded, but we got %v", err)
			}
		}
		s := rand.Intn(r) + 1
		for j := 0; j < s; j++ {
			ls.PopOrWait()
		}
	}
}
