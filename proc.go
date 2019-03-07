package main

import "sync"

// Proc runs create channel and runs consumer function fn n times.
// Return channel to be feed, and function done - which shall be called
// when weed is complete.
func Proc(fn func(chan interface{}), n ...int) (chan interface{}, func()) {
	m := 1
	for _, n := range n {
		m = n
	}

	chanSize := 512
	ch := make(chan interface{}, chanSize)

	wg := sync.WaitGroup{}
	for i := 0; i < m; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fn(ch)
		}()
	}

	done := func() {
		close(ch)
		wg.Wait()
	}

	return ch, done
}
