package concurrent

import (
	"context"
	"fmt"
	"sync"
)

// Do 泛型并发任务调度器，用于在 Go 中高效地并行处理一批任务
// numGo：goroutine数量必须大于0、numCh：chan缓冲区容量必须大于0、tasks：等待执行的任务列表，fn 工作携程调度的回调
func Do[T any](numGo, numCh int, tasks []T, fn func(ctx context.Context, task T)) error {
	return DoContext[T](context.TODO(), numGo, numCh, tasks, fn)
}

// DoContext 泛型并发任务调度器，用于在 Go 中高效地并行处理一批任务
// ctx：context、numGo：goroutine数量必须大于0、numCh：chan缓冲区容量必须大于0、tasks：等待执行的任务列表，fn 工作携程调度的回调
func DoContext[T any](ctx context.Context, numGo, numCh int, tasks []T, fn func(ctx context.Context, task T)) error {
	if len(tasks) == 0 {
		return nil
	}
	if numGo < 1 {
		return fmt.Errorf("numGo must be greater than 0")
	}
	if numCh < 1 {
		return fmt.Errorf("numCh must be greater than 0")
	}
	queue := make(chan T, numCh)
	wg := new(sync.WaitGroup)
	wg.Add(numGo)
	for i := 0; i < numGo; i++ {
		go do(context.WithValue(ctx, "goroutineId", i), wg, queue, fn)
	}
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	for i := 0; i < len(tasks); i++ {
		queue <- tasks[i]
	}
	close(queue)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

func do[T any](ctx context.Context, wg *sync.WaitGroup, queue chan T, fn func(ctx context.Context, task T)) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-queue:
			if !ok {
				return
			}
			fn(ctx, task)
		}
	}
}
