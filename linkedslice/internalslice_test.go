package linkedslice

import (
	"testing"
)

func TestInternalSlice(t *testing.T) {
	is := newInternalSlice()

	// from empty queue
	_, err := is.pop()
	if err == nil || err != errSliceIsEmpty {
		t.Fatalf("it should return `errSliceIsEmpty`, but instead we got %v", err)
	}

	if !is.canPush() {
		t.Fatal("Should be able to push, but we can't")
	}

	for i := 0; i < 128; i++ {
		err := is.push(uint64(i))
		if err != nil {
			t.Fatalf("It should not return error, cause slots still available, but instead we got %v", err)
		}
	}
	for i := 0; i < 128; i++ {
		_, err := is.pop()
		if err != nil {
			t.Fatalf("It should not return error, cause items still available, but instead we got %v", err)
		}
	}

	// after put half
	_, err = is.pop()
	if err == nil || err != errSliceIsEmpty {
		t.Fatalf("it should return `errSliceIsEmpty`, but instead we got %v", err)
	}

	if !is.canPush() {
		t.Fatal("Should be able to push, but we can't")
	}

	for i := 0; i < 128; i++ {
		err := is.push(uint64(i))
		if err != nil {
			t.Fatalf("It should not return error, cause slots still available, but instead we got %v", err)
		}
	}
	for i := 0; i < 128; i++ {
		_, err := is.pop()
		if err != nil {
			t.Fatalf("It should not return error, cause items still available, but instead we got %v", err)
		}
	}

	// after both is used up
	err = is.push(200)
	if err == nil || err != errSliceIsFull {
		t.Fatalf("it should return `errSliceIsFull`, but instead we got %v", err)
	}

	putInternalSlice(is)
}
