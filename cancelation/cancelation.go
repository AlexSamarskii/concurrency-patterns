package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func doTask(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Task canceled: %v\n", ctx.Err())
			return
		default:
			fmt.Println("Performing the task...")
			select {
			case <-time.After(1 * time.Second):
			case <-ctx.Done():
				fmt.Printf("Task interrupted during work: %v\n", ctx.Err())
				return
			}
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go doTask(ctx, &wg)

	time.Sleep(3 * time.Second)
	fmt.Println("Main: canceling task")
	cancel()

	wg.Wait()

	fmt.Println("Main goroutine completed")
}
