package queue

import (
	"container/heap"
	"errors"
	"sort"
)

// Item 定义队列中的元素
type Item[T any] struct {
	value    T   // 泛型数据值
	priority int // 优先级（分数）
	index    int // 用于堆操作的索引
}

// PriorityQueue 定义优先级队列（小顶堆）
type PriorityQueue[T any] []*Item[T]

func (pq PriorityQueue[T]) Len() int { return len(pq) }

func (pq PriorityQueue[T]) Less(i, j int) bool {
	return pq[i].priority < pq[j].priority // 小顶堆：优先级低的元素排在前面
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

// Pop 移除队列中的最小优先级元素
func (pq *PriorityQueue[T]) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

// FixedSizePriorityQueue 是一个固定大小的小顶堆优先级队列
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
	if len(pq.queue) < pq.maxLen {
		// 队列未满，直接添加
		item := &Item[T]{
			value:    value,
			priority: priority,
		}
		heap.Push(&pq.queue, item)
		return nil
	}

	// 队列已满，比较新元素的优先级与当前最小优先级
	if pq.queue[0].priority < priority {
		// 替换最小优先级的元素
		heap.Pop(&pq.queue)
		item := &Item[T]{
			value:    value,
			priority: priority,
		}
		heap.Push(&pq.queue, item)
		return nil
	}

	// 新元素优先级不高于当前队列中的元素，不添加
	return nil
}

// Peek 返回当前队列中优先级最低的元素（但不移除）
func (pq *FixedSizePriorityQueue[T]) Peek() (T, int, error) {
	if len(pq.queue) == 0 {
		var zero T
		return zero, 0, errors.New("queue is empty")
	}
	item := pq.queue[0]
	return item.value, item.priority, nil
}

// GetAll 返回队列中的所有元素，按优先级从高到低排序
func (pq *FixedSizePriorityQueue[T]) GetAll() []T {
	if len(pq.queue) == 0 {
		return nil
	}

	// 复制队列以保持原始堆结构
	copyQueue := make([]*Item[T], len(pq.queue))
	copy(copyQueue, pq.queue)

	// 按优先级从高到低排序
	sort.Slice(copyQueue, func(i, j int) bool {
		return copyQueue[i].priority > copyQueue[j].priority // 从高到低排序
	})

	// 提取排序后的值
	values := make([]T, len(copyQueue))
	for i, item := range copyQueue {
		values[i] = item.value
	}
	return values
}
