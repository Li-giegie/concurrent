# concurrent 
一个轻量、高效的泛型并发任务调度器，用于在 Go 中高效地并行处理批任务。

## 概述
``concurrent`` 旨在解决 Go 开发中批量任务并行执行的通用问题：无需手动管理 goroutine 池、无需重复编写并发控制逻辑，只需定义任务处理函数和待处理的任务列表，即可快速实现任务的并行调度。

泛型支持

无任何第三方依赖
## 获取
```
go get github.com/Li-giegie/concurrent
```

## 用法
处理一组打印到屏幕的任务
```go
package concurrent

import (
	"context"
	"log"
	"testing"
	"time"
)

func TestDo(t *testing.T) {
	err := Do[int](2, 2, []int{1, 2, 3, 4, 5}, func(ctx context.Context, task int) {
		time.Sleep(time.Second)
		log.Println(ctx.Value("goroutineId"), "task", task)
	})
	log.Println("done", err)
}

func TestDoContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	err := DoContext[int](ctx, 2, 2, []int{1, 2, 3, 4, 5}, func(ctx context.Context, task int) {
		time.Sleep(time.Second)
		log.Println(ctx.Value("goroutineId"), "task", task)
	})
	log.Println("done", err)
}

func TestDoChanContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*6)
	defer cancel()
	tasks := make(chan int, 2)
	go func() {
		for i := 0; i < 5; i++ {
			tasks <- i
		}
		close(tasks)
	}()
	err := DoChanContext[int](ctx, 2, tasks, func(ctx context.Context, task int) {
		time.Sleep(time.Second)
		log.Println(ctx.Value("goroutineId"), "task", task)
	})
	log.Println("done", err)
}
```