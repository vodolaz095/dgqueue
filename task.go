package dgqueue

import "time"

// Task is element describing things to be done in future
type Task[T comparable] struct {
	ExecuteAt time.Time
	Payload   T
}
