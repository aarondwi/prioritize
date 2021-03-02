package prioritize

// QInterface is the interface for queue used inside our main engine
// You may implement this to create custom priority queuing mechanism
//
// Our implementation has different semantic on Push/Pop.
// Push returns error, while Pop waits.
// This is by design, as we want Push to error fast
// (to notify customer and not overburden our system),
// but we want our Pop to wait until a task exists (so can do work).
//
// Those implementing this interface should be thread(goroutine)-safe.
type QInterface interface {
	PushOrError(item QItem) error
	PopOrWait() QItem
}
