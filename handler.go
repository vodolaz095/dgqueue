package dgqueue

import (
	"container/heap"
	"sort"
	"sync"
	"time"
)

// Handler is deferred queue handler used to store and retrieve (in time) tasks to be executed
type Handler[T comparable] struct {
	mu          sync.Mutex
	data        *queue[T]
	initialized bool
}

// New creates new deferred queue Handler.
func New[T comparable]() Handler[T] {
	data := queue[T]{}
	data.nextOn = time.Now().Add(maxNextInterval)
	heap.Init(&data)
	return Handler[T]{data: &data, initialized: true}
}

func (h *Handler[T]) checkInitialized() {
	if !h.initialized {
		panic("in order to use dqueue.Handler it should be created via dqueue.New()")
	}
}

// Len is thread save function to see how much tasks are in queue.
func (h *Handler[T]) Len() int {
	h.checkInitialized()
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.data.Len()
}

// Get extracts one of tasks from queue in thread safe manner. If second
// argument is true, it means task is ready to be executed and is removed from
// queue. If there are no ready tasks in queue, first argument is zero Task struct
// and second one - false.
func (h *Handler[T]) Get() (task Task[T], ready bool) {
	h.checkInitialized()
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.data.Len() == 0 {
		return Task[T]{}, false
	}
	if h.data.nextOn.After(time.Now()) {
		return Task[T]{}, false
	}
	item := heap.Pop(h.data).(*Task[T])
	if item.ExecuteAt.Before(time.Now()) { // ready
		return *item, true
	}
	heap.Push(h.data, item)
	return Task[T]{}, false
}

// ExecuteAt schedules task for execution on time desired, it returns true,
// if task is accepted.
func (h *Handler[T]) ExecuteAt(payload T, when time.Time) (ok bool) {
	h.checkInitialized()
	if when.Before(time.Now()) {
		return false
	}
	task := Task[T]{ExecuteAt: when, Payload: payload}
	h.mu.Lock()
	heap.Push(h.data, &task)
	h.mu.Unlock()
	return true
}

// ExecuteAfter schedules task for execution on after time.Duration provided, it
// returns true, if task is accepted.
func (h *Handler[T]) ExecuteAfter(payload T, after time.Duration) (ok bool) {
	return h.ExecuteAt(payload, time.Now().Add(after))
}

// Dump returns copy of contents of the queue in sorted manner, leaving queue intact.
func (h *Handler[T]) Dump() (ret []Task[T]) {
	h.checkInitialized()
	ret = make([]Task[T], h.data.Len())
	for i := range h.data.tasks {
		ret[i] = *h.data.tasks[i]
	}
	sort.SliceStable(ret, func(i, j int) bool {
		return ret[i].ExecuteAt.Before(ret[j].ExecuteAt)
	})
	return ret
}

// Prune resets queue.
func (h *Handler[T]) Prune() {
	h.checkInitialized()
	h.mu.Lock()
	h.data.prune()
	h.mu.Unlock()
}
