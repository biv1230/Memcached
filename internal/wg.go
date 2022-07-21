package internal

import "sync"

type WaitGroupWrapper struct {
	sync.WaitGroup
}

func (wg WaitGroupWrapper) Wrap(cb func()) {
	wg.Add(1)
	go func() {
		cb()
		wg.Done()
	}()
}
