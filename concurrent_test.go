package concurrent

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDo_Normal(t *testing.T) {
	var count int64
	err := Do[int](2, 2, []int{1, 2, 3, 4, 5}, func(ctx context.Context, index int, item int) error {
		atomic.AddInt64(&count, 1)
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if atomic.LoadInt64(&count) != 5 {
		t.Fatalf("expected 5 tasks processed, got %d", atomic.LoadInt64(&count))
	}
}

func TestDo_EmptyTasks(t *testing.T) {
	err := Do[int](2, 2, []int(nil), func(ctx context.Context, index int, item int) error {
		t.Fatal("should not be called")
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestDo_InvalidNumGo(t *testing.T) {
	err := Do[int](0, 2, []int{1, 2, 3}, func(ctx context.Context, index int, item int) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error for numGo < 1")
	}
	if err.Error() != "numGo must be greater than 0" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestDo_InvalidNumCh(t *testing.T) {
	err := Do[int](2, 0, []int{1, 2, 3}, func(ctx context.Context, index int, item int) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error for numCh < 1")
	}
	if err.Error() != "numCh must be greater than 0" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestDoContext_Normal(t *testing.T) {
	var count int64
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := DoContext[int](ctx, 2, 2, []int{1, 2, 3, 4, 5}, func(ctx context.Context, index int, item int) error {
		atomic.AddInt64(&count, 1)
		time.Sleep(10 * time.Millisecond)
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if atomic.LoadInt64(&count) != 5 {
		t.Fatalf("expected 5 tasks processed, got %d", atomic.LoadInt64(&count))
	}
}

func TestDoContext_EmptyTasks(t *testing.T) {
	err := DoContext[int](context.Background(), 2, 2, []int{}, func(ctx context.Context, index int, item int) error {
		t.Fatal("should not be called")
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestDoContext_InvalidNumGo(t *testing.T) {
	err := DoContext[int](context.Background(), -1, 2, []int{1}, func(ctx context.Context, index int, item int) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error for numGo < 1")
	}
}

func TestDoContext_InvalidNumCh(t *testing.T) {
	err := DoContext[int](context.Background(), 2, -1, []int{1}, func(ctx context.Context, index int, item int) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error for numCh < 1")
	}
}

func TestDoContext_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	err := DoContext[int](ctx, 2, 2, []int{1, 2, 3, 4, 5}, func(ctx context.Context, index int, item int) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})
	if err == nil {
		t.Fatal("expected context canceled error")
	}
}

func TestDoContext_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	err := DoContext[int](ctx, 2, 2, []int{1, 2, 3, 4, 5}, func(ctx context.Context, index int, item int) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})
	if err == nil {
		t.Fatal("expected context deadline exceeded error")
	}
}

func TestDoContext_GoroutineId(t *testing.T) {
	ids := make(map[int]struct{})
	var mu sync.Mutex
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// Use 10 tasks so all 3 goroutines get work
	err := DoContext[int](ctx, 3, 2, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, func(ctx context.Context, index int, item int) error {
		id := ctx.Value(GoroutineId).(int)
		mu.Lock()
		ids[id] = struct{}{}
		mu.Unlock()
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(ids) != 3 {
		t.Fatalf("expected 3 goroutine ids, got %d: %v", len(ids), ids)
	}
	for id := range ids {
		if id < 0 || id >= 3 {
			t.Fatalf("goroutine id %d out of range [0, 3)", id)
		}
	}
}

func TestDoChanContext_Normal(t *testing.T) {
	var count int64
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	tasks := make(chan Job[int], 2)
	go func() {
		for i := 1; i <= 5; i++ {
			tasks <- Job[int]{index: i, item: i}
		}
		close(tasks)
	}()
	err := DoChanContext[int](ctx, 2, tasks, func(ctx context.Context, index int, item int) error {
		atomic.AddInt64(&count, 1)
		time.Sleep(10 * time.Millisecond)
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if atomic.LoadInt64(&count) != 5 {
		t.Fatalf("expected 5 tasks processed, got %d", atomic.LoadInt64(&count))
	}
}

func TestDoChanContext_InvalidNumGo(t *testing.T) {
	ctx := context.Background()
	tasks := make(chan Job[int])
	close(tasks)
	err := DoChanContext[int](ctx, 0, tasks, func(ctx context.Context, index int, item int) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error for numGo < 1")
	}
	if err.Error() != "numGo must be greater than 0" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestDoChanContext_TaskError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	tasks := make(chan Job[int], 2)
	go func() {
		for i := 1; i <= 5; i++ {
			tasks <- Job[int]{index: i, item: i}
		}
		close(tasks)
	}()
	expectedErr := errors.New("task error")
	err := DoChanContext[int](ctx, 2, tasks, func(ctx context.Context, index int, item int) error {
		if item == 3 {
			return expectedErr
		}
		time.Sleep(10 * time.Millisecond)
		return nil
	})
	if err == nil {
		t.Fatal("expected error from task")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected task error, got %v", err)
	}
}

func TestDoChanContext_ErrSkip(t *testing.T) {
	var count int64
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	tasks := make(chan Job[int], 2)
	go func() {
		for i := 1; i <= 5; i++ {
			tasks <- Job[int]{index: i, item: i}
		}
		close(tasks)
	}()
	err := DoChanContext[int](ctx, 2, tasks, func(ctx context.Context, index int, item int) error {
		atomic.AddInt64(&count, 1)
		if item == 3 {
			return ErrSkip
		}
		time.Sleep(10 * time.Millisecond)
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error with ErrSkip, got %v", err)
	}
	if atomic.LoadInt64(&count) != 5 {
		t.Fatalf("expected 5 tasks processed, got %d", atomic.LoadInt64(&count))
	}
}

func TestDoChanContext_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	tasks := make(chan Job[int], 2)
	go func() {
		for i := 1; i <= 5; i++ {
			tasks <- Job[int]{index: i, item: i}
		}
		close(tasks)
	}()
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	err := DoChanContext[int](ctx, 2, tasks, func(ctx context.Context, index int, item int) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})
	if err == nil {
		t.Fatal("expected context canceled error")
	}
}

func TestDoChanContext_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	tasks := make(chan Job[int], 2)
	go func() {
		for i := 1; i <= 5; i++ {
			tasks <- Job[int]{index: i, item: i}
			time.Sleep(20 * time.Millisecond)
		}
		close(tasks)
	}()
	err := DoChanContext[int](ctx, 2, tasks, func(ctx context.Context, index int, item int) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})
	if err == nil {
		t.Fatal("expected context deadline exceeded error")
	}
}

func TestDoChanContext_EmptyChannel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	tasks := make(chan Job[int])
	close(tasks)
	err := DoChanContext[int](ctx, 2, tasks, func(ctx context.Context, index int, item int) error {
		t.Fatal("should not be called")
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error for empty channel, got %v", err)
	}
}

func TestDoChanContext_GoroutineId(t *testing.T) {
	ids := make(map[int]struct{})
	var mu sync.Mutex
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// Use larger buffer so tasks are distributed across goroutines
	tasks := make(chan Job[int], 10)
	go func() {
		for i := 1; i <= 20; i++ {
			tasks <- Job[int]{index: i, item: i}
		}
		close(tasks)
	}()
	err := DoChanContext[int](ctx, 3, tasks, func(ctx context.Context, index int, item int) error {
		id := ctx.Value(GoroutineId).(int)
		mu.Lock()
		ids[id] = struct{}{}
		mu.Unlock()
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	// Verify all goroutine ids are in valid range [0, 3)
	for id := range ids {
		if id < 0 || id >= 3 {
			t.Fatalf("goroutine id %d out of range [0, 3)", id)
		}
	}
}

func TestDoChanContext_SingleGoroutine(t *testing.T) {
	var count int64
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	tasks := make(chan Job[int], 1)
	go func() {
		for i := 1; i <= 3; i++ {
			tasks <- Job[int]{index: i, item: i}
		}
		close(tasks)
	}()
	err := DoChanContext[int](ctx, 1, tasks, func(ctx context.Context, index int, item int) error {
		atomic.AddInt64(&count, 1)
		time.Sleep(10 * time.Millisecond)
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if atomic.LoadInt64(&count) != 3 {
		t.Fatalf("expected 3 tasks processed, got %d", atomic.LoadInt64(&count))
	}
}

func TestDoChanContext_LargeBuffer(t *testing.T) {
	var count int64
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	tasks := make(chan Job[int], 100)
	go func() {
		for i := 1; i <= 50; i++ {
			tasks <- Job[int]{index: i, item: i}
		}
		close(tasks)
	}()
	err := DoChanContext[int](ctx, 5, tasks, func(ctx context.Context, index int, item int) error {
		atomic.AddInt64(&count, 1)
		time.Sleep(5 * time.Millisecond)
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if atomic.LoadInt64(&count) != 50 {
		t.Fatalf("expected 50 tasks processed, got %d", atomic.LoadInt64(&count))
	}
}

func TestDoChanContext_ConcurrentSendClose(t *testing.T) {
	var count int64
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	tasks := make(chan Job[int], 10)
	go func() {
		for i := 1; i <= 20; i++ {
			tasks <- Job[int]{index: i, item: i}
		}
		close(tasks)
	}()
	err := DoChanContext[int](ctx, 4, tasks, func(ctx context.Context, index int, item int) error {
		atomic.AddInt64(&count, 1)
		time.Sleep(5 * time.Millisecond)
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if atomic.LoadInt64(&count) != 20 {
		t.Fatalf("expected 20 tasks processed, got %d", atomic.LoadInt64(&count))
	}
}

func TestDoChanContext_ErrSkipMultiple(t *testing.T) {
	var count int64
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	tasks := make(chan Job[int], 2)
	go func() {
		for i := 1; i <= 5; i++ {
			tasks <- Job[int]{index: i, item: i}
		}
		close(tasks)
	}()
	err := DoChanContext[int](ctx, 2, tasks, func(ctx context.Context, index int, item int) error {
		atomic.AddInt64(&count, 1)
		if item%2 == 0 {
			return ErrSkip
		}
		time.Sleep(10 * time.Millisecond)
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error with ErrSkip, got %v", err)
	}
	if atomic.LoadInt64(&count) != 5 {
		t.Fatalf("expected 5 tasks processed, got %d", atomic.LoadInt64(&count))
	}
}

func TestDoChanContext_WrappedContextCancel(t *testing.T) {
	baseCtx, baseCancel := context.WithCancel(context.Background())
	ctx, cancel := context.WithTimeout(baseCtx, time.Second)
	defer cancel()
	tasks := make(chan Job[int], 2)
	go func() {
		for i := 1; i <= 5; i++ {
			tasks <- Job[int]{index: i, item: i}
		}
		close(tasks)
	}()
	go func() {
		time.Sleep(50 * time.Millisecond)
		baseCancel()
	}()
	err := DoChanContext[int](ctx, 2, tasks, func(ctx context.Context, index int, item int) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})
	if err == nil {
		t.Fatal("expected context canceled error")
	}
}
