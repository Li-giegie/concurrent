# concurrent

A lightweight, efficient generic concurrent task scheduler for parallel batch processing in Go.

## Features

- Generic support (Go 1.18+)
- Zero third-party dependencies
- Context cancellation and timeout control
- Streaming input via channel
- Error propagation (return `ErrSkip` to skip errors without terminating)

## Install

```bash
go get github.com/Li-giegie/concurrent
```

## Core Concepts

- `numGo`: number of concurrent goroutines
- `numCh`: buffer capacity of the task channel
- `GoroutineId`: context key with `int` value in range `[0, numGo)`, identifies which worker is executing

## API Hierarchy

```
Do          → DoContext        → DoChanContext
(no context)  (task slice)       (streaming/channel)
```

## Usage

### Basic (task slice)

```go
err := Do[int](2, 2, []int{1, 2, 3, 4, 5}, func(ctx context.Context, task int) error {
    time.Sleep(time.Second)
    log.Println(ctx.Value(GoroutineId), "task", task)
    return nil
})
```

### Context control (timeout/cancel)

```go
ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
defer cancel()

err := DoContext[int](ctx, 2, 2, []int{1, 2, 3, 4, 5}, func(ctx context.Context, task int) error {
    time.Sleep(time.Second)
    log.Println(ctx.Value(GoroutineId), "task", task)
    return nil
})
```

### Streaming input (channel)

Suitable for unbounded or streaming task sources:

```go
ctx, cancel := context.WithTimeout(context.Background(), time.Second*6)
defer cancel()

tasks := make(chan int, 2)
go func() {
    for i := 0; i < 5; i++ {
        tasks <- i
    }
    close(tasks)
}()

err := DoChanContext[int](ctx, 2, tasks, func(ctx context.Context, task int) error {
    time.Sleep(time.Second)
    log.Println(ctx.Value(GoroutineId), "task", task)
    return nil
})
```

### Error Handling

- Return `nil`: continue execution
- Return `ErrSkip`: skip error and continue
- Return other error: terminate all workers immediately

```go
err := Do[int](2, 2, []int{1, 2, 3, 4, 5}, func(ctx context.Context, task int) error {
    if task == 2 {
        return ErrSkip // skip, continue other tasks
    }
    if task == 3 {
        return errors.New("fatal") // terminate scheduler
    }
    return nil
})
```

## Notes

- `numGo` and `numCh` must be greater than 0
- The channel passed to `DoChanContext` must be closed after all tasks are submitted
- `GoroutineId` is a reserved context key; do not override it

## 文档

- [中文文档](README_CN.md)
