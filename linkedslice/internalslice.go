package linkedslice

import (
	"errors"
	"sync"
)

var internalSliceSize = 256

// Bounded one-way slices, not a circular one.
// Designed this way to maintain FIFO semantic, even after it is full.
//
// This struct is NOT thread(goroutine)-safe.
type internalSlice struct {
	head      int
	tail      int
	sizeLimit int
	arr       []uint64
	next      *internalSlice
}

var internalSlicePool = &sync.Pool{
	New: func() interface{} {
		return &internalSlice{
			head:      0,
			tail:      0,
			sizeLimit: internalSliceSize,
			arr:       make([]uint64, internalSliceSize),
		} // 256 * 8 = 2048 bytes / 2KB, a lot already
	},
}

func newInternalSlice() *internalSlice {
	return internalSlicePool.Get().(*internalSlice)
}

func putInternalSlice(is *internalSlice) {
	is.head = 0
	is.tail = 0
	is.next = nil
	internalSlicePool.Put(is)
}

var errSliceIsFull = errors.New("this slice is full")
var errSliceIsEmpty = errors.New("this slice is empty")

func (is *internalSlice) push(n uint64) error {
	if !is.canPush() {
		return errSliceIsFull
	}
	is.arr[is.head] = n
	is.head++
	return nil
}

func (is *internalSlice) pop() (uint64, error) {
	if is.isEmpty() {
		return 0, errSliceIsEmpty
	}
	result := is.arr[is.tail]
	is.tail++
	return result, nil
}

func (is *internalSlice) canPush() bool {
	return is.head < is.sizeLimit
}

func (is *internalSlice) isEmpty() bool {
	return is.head == 0 || is.tail == is.head
}

func (is *internalSlice) slotsUsedUp() bool {
	return is.tail == is.sizeLimit
}
