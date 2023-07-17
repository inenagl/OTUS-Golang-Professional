package hw05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
	"runtime"
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

		require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})

	// Это и дополнительный unit-тест, и задание со звёздочкой: проверка конкурентности без использования time.Sleep
	t.Run("without errors limit", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var inProgressTasksCount int32

		workersCount := 5
		maxErrorsCount := 0

		tasks = append(
			tasks,
			// Эта задача будет проверять, что кол-во параллельно (конкурентно) выполняемых задач больше одной
			func() error {
				atomic.AddInt32(&inProgressTasksCount, 1)
				defer atomic.AddInt32(&inProgressTasksCount, -1)

				atomic.AddInt32(&runTasksCount, 1)
				require.Eventually(t, func() bool {
					return atomic.LoadInt32(&inProgressTasksCount) > 2
				}, time.Second, time.Millisecond, "tasks were run sequentially?")
				return nil
			},
			// Эта задача будет проверять, что кол-во параллельно (конкурентно) выполняемых задач
			// не больше заданного числа горутин
			func() error {
				atomic.AddInt32(&inProgressTasksCount, 1)
				defer atomic.AddInt32(&inProgressTasksCount, -1)

				atomic.AddInt32(&runTasksCount, 1)
				require.Never(t, func() bool {
					return atomic.LoadInt32(&inProgressTasksCount) > int32(workersCount)
				}, time.Second, time.Millisecond, "more simultaneous tasks then required")
				return nil
			},
		)
		// Остальные задачи имитируют работу и возвращают ошибку
		for i := 0; i < tasksCount-2; i++ {
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				atomic.AddInt32(&inProgressTasksCount, 1)
				defer atomic.AddInt32(&inProgressTasksCount, -1)

				// Немножко поспать всё-равно приходится, иначе всё очень быстро пролетает, и проверки не успевают
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(10)))
				atomic.AddInt32(&runTasksCount, 1)
				return err
			})
		}

		err := Run(tasks, workersCount, maxErrorsCount)

		require.NoError(t, err)
		require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
	})

	t.Run("no tasks", func(t *testing.T) {
		tasks := make([]Task, 0)

		workersCount := 5
		maxErrorsCount := 1
		expectedGoroutines := runtime.NumGoroutine()

		err := Run(tasks, workersCount, maxErrorsCount)

		require.NoError(t, err)
		require.Equal(t, expectedGoroutines, runtime.NumGoroutine(), "not all goroutines were finished")
	})

	t.Run("no goroutines", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			tasks = append(tasks, func() error {
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 0
		maxErrorsCount := 1
		expectedGoroutines := runtime.NumGoroutine()

		err := Run(tasks, workersCount, maxErrorsCount)
		require.NoError(t, err)

		require.Equal(t, int32(0), runTasksCount, "some tasks were completed")
		require.Equal(t, expectedGoroutines, runtime.NumGoroutine(), "extra goroutines are running")
	})
}
