# concurrent

一个轻量、高效的泛型并发任务调度器，用于在 Go 中高效地并行处理批任务。

## 特性

- 泛型支持（Go 1.18+）
- 无任何第三方依赖
- 支持 context 取消与超时控制
- 支持 channel 模式流式输入
- 错误传播机制（返回 `ErrSkip` 可跳过错误而非终止调度）

## 安装

```bash
go get github.com/Li-giegie/concurrent
```

## 核心概念

- `numGo`：同时运行的 goroutine 数量
- `numCh`：任务 channel 的缓冲区容量
- `GoroutineId`：context 中的 key，value 类型为 `int`，范围 `[0, numGo)`，用于标识当前执行任务的 worker

## API 层次

```
Do          → DoContext        → DoChanContext
（无context）  （任务切片）          （任务流/Channel）
```

## 用法

### 基本用法（任务切片）

```go
err := Do[int](2, 2, []int{1, 2, 3, 4, 5}, func(ctx context.Context, task int) error {
    time.Sleep(time.Second)
    log.Println(ctx.Value(GoroutineId), "task", task)
    return nil
})
```

### context 控制（超时/取消）

```go
ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
defer cancel()

err := DoContext[int](ctx, 2, 2, []int{1, 2, 3, 4, 5}, func(ctx context.Context, task int) error {
    time.Sleep(time.Second)
    log.Println(ctx.Value(GoroutineId), "task", task)
    return nil
})
```

### 流式输入（channel）

适用于任务源不确定或无限流的场景：

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

### 错误处理

- 返回 `nil`：继续执行
- 返回 `ErrSkip`：跳过错误，继续执行
- 返回其他 error：立即终止所有 worker，并返回该错误

```go
err := Do[int](2, 2, []int{1, 2, 3, 4, 5}, func(ctx context.Context, task int) error {
    if task == 2 {
        return ErrSkip // 跳过，继续其他任务
    }
    if task == 3 {
        return errors.New("fatal") // 终止调度
    }
    return nil
})
```

## 注意事项

- `numGo` 和 `numCh` 必须大于 0
- `DoChanContext` 传入的 channel 必须在任务完成后关闭
- `GoroutineId` 是保留 context key，请勿覆盖
