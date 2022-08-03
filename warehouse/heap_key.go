package warehouse

import (
	"container/heap"
)

type keyItem struct {
	key      string
	index    int
	pastTime int64
}

func newKeyItem(key string, et int64) *keyItem {
	return &keyItem{
		key:      key,
		pastTime: et,
	}
}

type keyQueue struct {
	items []*keyItem
}

func newKewQueue() *keyQueue {
	kq := &keyQueue{
		items: make([]*keyItem, 0, 100),
	}
	heap.Init(kq)
	return kq
}

func (kq *keyQueue) Len() int {
	return len(kq.items)
}

func (kq *keyQueue) Less(i, j int) bool {
	return kq.items[i].pastTime < kq.items[j].pastTime
}

func (kq *keyQueue) Swap(i, j int) {
	kq.items[i], kq.items[j] = kq.items[j], kq.items[i]
	kq.items[i].index = i
	kq.items[j].index = j
}

func (kq *keyQueue) Push(x any) {
	n := len(kq.items)
	item := x.(*keyItem)
	item.index = n
	kq.items = append(kq.items, item)
}

func (kq *keyQueue) Pop() any {
	n := len(kq.items)
	item := kq.items[n-1]
	item.index = -1
	kq.items = kq.items[0 : n-1]
	return item
}

func (kq *keyQueue) add(x any) {
	heap.Push(kq, x)
}

func (kq *keyQueue) removeFirst() any {
	return heap.Pop(kq)
}

func (kq *keyQueue) remove(i int) any {
	return heap.Remove(kq, i)
}

func (kq *keyQueue) fix(i int) {
	heap.Fix(kq, i)
}

func (kq *keyQueue) first() *keyItem {
	return kq.items[0]
}

func (kq *keyQueue) len() int {
	return len(kq.items)
}
