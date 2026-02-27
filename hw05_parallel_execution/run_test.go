package hw05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("if were errors in first M tasks, than finished not more N+M tasks", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				atomic.AddInt32(&runTasksCount, 1)
				return err
			})
		}

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount), "extra tasks were started")
	})

	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 5
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		require.NoError(t, err)

		require.Equal(t, runTasksCount, int32(tasksCount), "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})
}

func TestNegativeAllowedErrors(t *testing.T) {
	t.Run("m is negative, expected errors limit exceeded", func(t *testing.T) {
		workersCount := 10
		maxErrorsCount := -5
		tasksCount := 50

		tasks := make([]Task, 0, tasksCount)

		for i := 0; i < tasksCount; i++ {
			tasks = append(tasks, func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				return nil
			})
		}
		err := Run(tasks, workersCount, maxErrorsCount)
		require.ErrorIs(t, ErrErrorsLimitExceeded, err)
	})
}

// Если задачи работают без ошибок, то выполнятся `len(tasks)` задач, т.е. все задачи.
func TestAllTaskCompleted(t *testing.T) {
	t.Run("all tasks completed without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 5
		maxErrorsCount := 1

		err := Run(tasks, workersCount, maxErrorsCount)
		require.NoError(t, err)

		require.Equal(t, runTasksCount, int32(tasksCount), "not all tasks were completed")
	})
}

// Если в первых выполненных m задачах (или вообще всех) происходят ошибки, то всего выполнится не более n+m задач.
func TestFirstTasksWithErrors(t *testing.T) {
	t.Run("first m tasks with errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep
			err := fmt.Errorf("error from task %d", i)

			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return err
			})
		}

		workersCount := 5
		maxErrorsCount := 7

		err := Run(tasks, workersCount, maxErrorsCount)
		require.ErrorIs(t, ErrErrorsLimitExceeded, err)

		require.LessOrEqual(t, runTasksCount, int32(maxErrorsCount+workersCount), "not all tasks were completed")
	})
}

// Дополнительное задание: написать тест на concurrency без time.Sleep.
func TestWorkersCount(t *testing.T) {
	t.Run("concurrency limit should not be exceeded", func(t *testing.T) {
		tasksCount := 50
		workersCount := 5
		maxErrorsCount := 1000

		var active int64
		var maxActive int64 // максимальное значение active за всё время
		var completed int64 // количество завершённых задач

		tasks := make([]Task, tasksCount)
		for i := 0; i < tasksCount; i++ {
			tasks[i] = func() error {
				atomic.AddInt64(&active, 1)
				time.Sleep(30 * time.Millisecond)
				atomic.AddInt64(&active, -1)
				atomic.AddInt64(&completed, 1)
				return nil
			}
		}
		stopCh := make(chan struct{})
		// Мониторинг максимума активных задач
		go func() {
			ticker := time.NewTicker(1 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					cur := atomic.LoadInt64(&active)
					for {
						old := atomic.LoadInt64(&maxActive)
						if cur <= old {
							break
						}
						if atomic.CompareAndSwapInt64(&maxActive, old, cur) {
							break
						}
					}
				case <-stopCh:
					return
				}
			}
		}()

		err := Run(tasks, workersCount, maxErrorsCount)
		require.NoError(t, err)

		require.Eventually(t, func() bool {
			return atomic.LoadInt64(&completed) == int64(tasksCount)
		}, 2*time.Second, 10*time.Millisecond, "not all tasks completed")

		close(stopCh)

		require.LessOrEqual(t, atomic.LoadInt64(&maxActive), int64(workersCount),
			"number of concurrent workers exceeded limit")
	})
}
