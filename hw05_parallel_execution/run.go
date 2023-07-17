package hw05parallelexecution

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if n == 0 || len(tasks) == 0 {
		return nil
	}

	ch := make(chan Task, len(tasks))
	for _, task := range tasks {
		ch <- task
	}
	close(ch)

	wg := sync.WaitGroup{}
	wg.Add(n)

	var errorsCount atomic.Uint32

	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()

			for {
				if m > 0 && errorsCount.Load() >= uint32(m) {
					return
				}

				task, ok := <-ch
				if !ok {
					return
				}

				if err := task(); err != nil {
					errorsCount.Add(1)
				}
			}
		}()
	}

	wg.Wait()

	if m > 0 && errorsCount.Load() >= uint32(m) {
		return ErrErrorsLimitExceeded
	}

	return nil
}
