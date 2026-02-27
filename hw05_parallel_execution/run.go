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
	chBuffered := make(chan struct{}, n)
	var wg sync.WaitGroup

	var accerrors atomic.Int64
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, t := range tasks {
			chBuffered <- struct{}{}
			if accerrors.Load() >= int64(m) {
				break
			}
			wg.Add(1)
			go func(t Task) error {
				defer wg.Done()
				err := t()
				if err != nil {
					accerrors.Add(int64(1))
				}
				<-chBuffered

				return err
			}(t)
		}
	}()
	wg.Wait()
	if accerrors.Load() >= int64(m) {
		return ErrErrorsLimitExceeded
	}
	return nil
}
