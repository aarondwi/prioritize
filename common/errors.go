package common

import "errors"

// ErrQueueIsFull is returned to prevent some task to getting too high latency.
//
// Better fail fast than seems as down.
var ErrQueueIsFull = errors.New("queue is full, rejecting new qitem")

// ErrQueueIsClosed is returned when PushOrError() or PopOrWaitTillClose()
// is called after Close() is called
var ErrQueueIsClosed = errors.New("queue is already closed, can't accept new request")

// ErrParamShouldBePositive is returned when either sizeLimit or priority parameter is negative
var ErrParamShouldBePositive = errors.New("sizeLimit and priority given should be positive")

// ErrPriorityOutOfRange is returned if priority given is outside of range
//
// If we accept it, to maintain the guarantee, needs to maintain too much queue,
// and hard to scan over.
var ErrPriorityOutOfRange = errors.New("Roundrobin Priority Queue is full, rejecting new qitem")
