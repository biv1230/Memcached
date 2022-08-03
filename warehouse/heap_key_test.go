package warehouse

import (
	"container/heap"
	"testing"
)

func TestKeyQueue(t *testing.T) {
	q := &keyQueue{
		items: make([]*keyItem, 0, 10),
	}
	q.items = append(q.items, &keyItem{
		key:      "xxxx",
		pastTime: 11,
	})
	q.items = append(q.items, &keyItem{
		key:      "xxxx",
		pastTime: 10,
	})
	q.items = append(q.items, &keyItem{
		key:      "xxxx",
		pastTime: 3,
	})
	q.items = append(q.items, &keyItem{
		key:      "xxxx",
		pastTime: 17,
	})
	q.items = append(q.items, &keyItem{
		key:      "xxxx",
		pastTime: 6,
	})
	q.items = append(q.items, &keyItem{
		key:      "xxxx",
		pastTime: 180,
	})
	q.items = append(q.items, &keyItem{
		key:      "xxxx",
		pastTime: 1,
	})
	q.items = append(q.items, &keyItem{
		key:      "xxxx",
		pastTime: 12,
	})
	q.items = append(q.items, &keyItem{
		key:      "xxxx",
		pastTime: 19,
	})
	q.items = append(q.items, &keyItem{
		key:      "xxxx",
		pastTime: 2,
	})
	q.items = append(q.items, &keyItem{
		key:      "xxxx",
		pastTime: 11,
	})

	t.Log(len(q.items))

	heap.Init(q)

	if q.items[0].pastTime != 1 {
		t.Error("heap err")
	}

}
