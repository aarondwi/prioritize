package common

import "errors"

// ErrQueueIsFull is returned to prevent some task to getting too high latency.
// Better fail fast than seems as down
var ErrQueueIsFull = errors.New("queue is full, rejecting new qitem")
