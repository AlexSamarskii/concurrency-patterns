package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func doTask(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Task canceled:", ctx.Err())
			return
		case <-ticker.C:
			fmt.Println("Performing the task...")
		}
	}
}

func main() {
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	wg.Add(1)
	go doTask(ctx, &wg)

	wg.Wait()
	fmt.Println("Main goroutine completed")
}
