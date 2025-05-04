package dgqueue

import (
	"time"
)

type queue[T comparable] struct {
	nextOn time.Time
	tasks  []*Task[T]
}

// Len is used for implementing sort.Interface.
func (dq *queue[T]) Len() int {
	return len(dq.tasks)
}

// Less is used for implementing sort.Interface.
func (dq *queue[T]) Less(i, j int) bool {
	return dq.tasks[i].ExecuteAt.Before(dq.tasks[j].ExecuteAt)
}

// Swap is used for implementing sort.Interface.
func (dq *queue[T]) Swap(i, j int) {
	dq.tasks[i], dq.tasks[j] = dq.tasks[j], dq.tasks[i]
}

// Push is used for implementing heap.Interface.
func (dq *queue[T]) Push(x any) {
	item, ok := x.(*Task[T])
	if !ok {
		return
	}
	if item.ExecuteAt.Before(time.Now()) {
		return
	}
	if item.ExecuteAt.Before(dq.nextOn) {
		dq.nextOn = item.ExecuteAt
	}
	dq.tasks = append(dq.tasks, item)
}

// Pop is used for implementing heap.Interface.
func (dq *queue[T]) Pop() any {
	old := dq.tasks
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	dq.tasks = old[0 : n-1]
	dq.nextOn = item.ExecuteAt
	return item
}

func (dq *queue[T]) prune() {
	dq.nextOn = time.Now().Add(maxNextInterval)
	dq.tasks = nil
}
