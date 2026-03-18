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
		log.Println(ctx.Value("id"), "task", task)
	})
	log.Println("done", err)
}
