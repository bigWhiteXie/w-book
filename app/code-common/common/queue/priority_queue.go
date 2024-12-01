package queue

import (
	"container/heap"
	"errors"
)

type Item[T any] struct {
	value    T   // 泛型数据值
	priority int // 优先级（分数）
	index    int // 用于堆操作的索引
}

// 定义优先级队列（小顶堆），使用泛型
type PriorityQueue[T any] []*Item[T]

func (pq PriorityQueue[T]) Len() int { return len(pq) }

func (pq PriorityQueue[T]) Less(i, j int) bool {
	// 小顶堆，优先级高的元素排在前面
	return pq[i].priority < pq[j].priority
}

func (pq PriorityQueue[T]) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

// Push 添加元素到队列
func (pq *PriorityQueue[T]) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item[T])
	item.index = n
	*pq = append(*pq, item)
}

// Pop 移除队列中的最小元素
func (pq *PriorityQueue[T]) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

// PriorityQueue 是一个固定大小的优先级队列
type FixedSizePriorityQueue[T any] struct {
	queue  PriorityQueue[T]
	maxLen int
}

// NewFixedSizePriorityQueue 创建一个固定大小的优先级队列
func NewFixedSizePriorityQueue[T any](maxLen int) *FixedSizePriorityQueue[T] {
	return &FixedSizePriorityQueue[T]{
		queue:  make(PriorityQueue[T], 0, maxLen),
		maxLen: maxLen,
	}
}

// Enqueue 向队列中添加元素
func (pq *FixedSizePriorityQueue[T]) Enqueue(value T, priority int) error {
	if len(pq.queue) >= pq.maxLen {
		return errors.New("queue is full")
	}
	item := &Item[T]{
		value:    value,
		priority: priority,
	}
	heap.Push(&pq.queue, item)
	return nil
}

// Dequeue 从队列中取出最小优先级的元素
func (pq *FixedSizePriorityQueue[T]) Dequeue() (T, int, error) {
	if len(pq.queue) == 0 {
		var zero T
		return zero, 0, errors.New("queue is empty")
	}
	item := heap.Pop(&pq.queue).(*Item[T])
	return item.value, item.priority, nil
}

// Peek 返回当前队列中优先级最小的元素（但不移除）
func (pq *FixedSizePriorityQueue[T]) Peek() (T, int, error) {
	if len(pq.queue) == 0 {
		var zero T
		return zero, 0, errors.New("queue is empty")
	}
	item := pq.queue[0]
	return item.value, item.priority, nil
}

// GetAll 返回队列中的所有元素（包括它们的优先级）
func (pq *FixedSizePriorityQueue[T]) GetAll() []T {
	if len(pq.queue) == 0 {
		return nil
	}
	values := make([]T, len(pq.queue))
	priorities := make([]int, len(pq.queue))
	for i, item := range pq.queue {
		values[i] = item.value
		priorities[i] = item.priority
	}
	return values
}
