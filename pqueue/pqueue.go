package pqueue

import "container/heap"

type innerQueue[T any] struct {
	items []T
	cmp   func(T, T) bool
}

func (q *innerQueue[T]) Len() int {
	return len(q.items)
}

func (q *innerQueue[T]) Less(i, j int) bool {
	return q.cmp(q.items[i], q.items[j])
}

func (q *innerQueue[T]) Swap(i, j int) {
	q.items[i], q.items[j] = q.items[j], q.items[i]
}

func (q *innerQueue[T]) Push(item any) {
	q.items = append(q.items, item.(T))
}

func (q *innerQueue[T]) Pop() any {
	l := len(q.items)
	if l == 0 {
		return nil
	}
	item := q.items[l-1]
	q.items = q.items[:l-1]
	return item
}

type PQueue[T any] struct {
	inner *innerQueue[T]
}

func New[T any](cmp func(T, T) bool) *PQueue[T] {
	return &PQueue[T]{
		inner: &innerQueue[T]{cmp: cmp},
	}
}

func (q *PQueue[T]) Empty() bool {
	return len(q.inner.items) == 0
}

func (q *PQueue[T]) Push(item T) {
	heap.Push(q.inner, item)
}

func (q *PQueue[T]) Pop() (T, bool) {
	var zero T
	item := heap.Pop(q.inner)
	if item == nil {
		return zero, false
	}
	return item.(T), true
}
