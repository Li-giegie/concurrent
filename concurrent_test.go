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
