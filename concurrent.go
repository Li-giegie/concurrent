package concurrent

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

const GoroutineId = "GoroutineId"

type DoFunc[T any] func(ctx context.Context, index int, item T) error

var (
	// ErrSkip 如果 DoFunc 返回 error != nil 终止调用，如果不想 则返回 ErrSkip 跳过错误
	ErrSkip = errors.New("skip")
)

// Do 泛型并发任务调度器，用于在 Go 中高效地并行处理一批任务
// numGo：goroutine数量必须大于0、numCh：chan缓冲区容量必须大于0、tasks：等待执行的任务列表，fn 工作携程调度的回调
func Do[T any](numGo, numCh int, tasks []T, fn DoFunc[T]) error {
	return DoContext(context.Background(), numGo, numCh, tasks, fn)
}

type Job[T any] struct {
	index int
	item  T
}

// DoContext 泛型并发任务调度器，用于在 Go 中高效地并行处理一批任务
// ctx：context、numGo：goroutine数量必须大于0、numCh：chan缓冲区容量必须大于0、tasks：等待执行的任务列表，fn 工作携程调度的回调
func DoContext[T any](ctx context.Context, numGo, numCh int, tasks []T, fn DoFunc[T]) (err error) {
	if len(tasks) == 0 {
		return nil
	}
	if numGo < 1 {
		return fmt.Errorf("numGo must be greater than 0")
	}
	if numCh < 1 {
		return fmt.Errorf("numCh must be greater than 0")
	}
	in := make(chan Job[T], numCh)
	go func() {
		defer close(in)
		for i := 0; i < len(tasks); i++ {
			select {
			case <-ctx.Done():
				return
			default:
				in <- Job[T]{
					index: i,
					item:  tasks[i],
				}
			}
		}
	}()
	return DoChanContext(ctx, numGo, in, fn)
}

// DoChanContext 泛型并发任务调度器，用于在 Go 中高效地并行处理一批任务
// ctx：context、numGo：goroutine数量必须大于0、in：任务输入管道，处理完必须关闭管道、fn 工作携程调度的回调
func DoChanContext[T any](ctx context.Context, numGo int, in chan Job[T], fn DoFunc[T]) (err error) {
	if numGo < 1 {
		return fmt.Errorf("numGo must be greater than 0")
	}
	defer func() {
		for range in {
		}
	}()
	errCtx, errCancel := context.WithCancel(context.TODO())
	defer errCancel()

	wg := new(sync.WaitGroup)
	for i := 0; i < numGo; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			gCtx := context.WithValue(errCtx, GoroutineId, i)
			for {
				select {
				case <-ctx.Done():
					return
				case <-errCtx.Done():
					return
				case task, ok := <-in:
					if !ok {
						return
					}
					if rErr := fn(gCtx, task.index, task.item); rErr != nil {
						if !errors.Is(rErr, ErrSkip) {
							err = rErr
							errCancel()
							return
						}
					}
				}
			}
		}(i)
	}
	doneCtx, doneCancel := context.WithCancel(context.TODO())
	defer doneCancel()
	go func() {
		wg.Wait()
		doneCancel()
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-errCtx.Done():
		return
	case <-doneCtx.Done():
		return nil
	}
}
