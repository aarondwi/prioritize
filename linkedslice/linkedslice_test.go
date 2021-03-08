package linkedslice

import (
	"log"
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
		res, err := ls.PopOrWaitTillClose()
		if err != nil {
			t.Fatalf("It should not error, because not closed yet, but we got %v", err)
		}
		if res.ID != uint64(i) {
			t.Fatalf("We don't receive FIFO as we expected: expected %d, got %d", uint64(i), res.ID)
		}
	}
	ls.Close()
}

func TestLinkedSlicePopWait(t *testing.T) {
	ls := NewLinkedSlice()

	c := make(chan bool, 1)
	go func() {
		time.Sleep(200 * time.Millisecond)
		log.Println("timeout, returning")
		c <- false
	}()

	go func() {
		item, err := ls.PopOrWaitTillClose()
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
	err := ls.PushOrError(common.QItem{ID: 100})
	if err != nil {
		t.Fatalf("It should not error because slots are available, but we got %v", err)
	}

	result := <-c
	if !result {
		t.Fatal("We should receive true, because all the above are true, but we are not")
	}

	ls.Close()
}

func TestLinkedSliceAfterClose(t *testing.T) {
	ls := NewLinkedSlice()
	ls.Close()

	err := ls.PushOrError(common.QItem{})
	if err == nil || err != common.ErrQueueIsClosed {
		t.Fatalf("It should be error, cause already closed, but it is not")
	}

	_, err = ls.PopOrWaitTillClose()
	if err == nil || err != common.ErrQueueIsClosed {
		t.Fatalf("It should be error, cause already closed, but it is not")
	}
}

func BenchmarkLinkedSlice(b *testing.B) {
	ls := NewLinkedSlice()
	for i := 0; i < b.N; i++ {
		ls.PushOrError(common.QItem{ID: uint64(i)})
		ls.PopOrWaitTillClose()
	}
	ls.Close()
}

func BenchmarkLinkedSliceInLoop(b *testing.B) {
	ls := NewLinkedSlice()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 128; j++ {
			ls.PushOrError(common.QItem{ID: uint64(i*128 + j)})
		}
		for j := 0; j < 128; j++ {
			ls.PopOrWaitTillClose()
		}
	}
	ls.Close()
}

func BenchmarkLinkedSliceParallel(b *testing.B) {
	ls := NewLinkedSlice()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ls.PushOrError(common.QItem{})
			ls.PopOrWaitTillClose()
		}
	})
	ls.Close()
}

func BenchmarkLinkedSliceInLoopParallel(b *testing.B) {
	ls := NewLinkedSlice()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for j := 0; j < 128; j++ {
				ls.PushOrError(common.QItem{ID: uint64(j)})
			}
			for j := 0; j < 128; j++ {
				ls.PopOrWaitTillClose()
			}
		}
	})
	ls.Close()
}
