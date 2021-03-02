package prioritize

import "math"

// QItem is the item we put into our priority queue implementation.
// It is basically an index equivalent in usual DBMS.
//
// Given this is small (8 bytes for uint64 and usually 4 bytes for int),
// it gonna results in 12 bytes (let's assume 16 bytes, cause buffer or 64-bit alignment, shall we?).
// For 1000 items (which is a lot of task waiting for most webserver/batch), it will only be 16KB,
// well far below the usual size of L1 cache (64KB).
// So checking and swapping will be really fast.
//
// Of course, as long as not be used as a pointer individually.
type QItem struct {
	id       uint64
	priority int
}

// MinQItem is a holder
// for the lowest possible priority for an item
var MinQItem = QItem{priority: math.MinInt32}
