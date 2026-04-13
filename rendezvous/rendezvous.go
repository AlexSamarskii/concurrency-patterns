package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	rendezvousCh1 := make(chan string)
	rendezvousCh2 := make(chan string)

	wg.Add(2)
	go goroutine1(rendezvousCh1, rendezvousCh2, &wg)
	go goroutine2(rendezvousCh1, rendezvousCh2, &wg)

	wg.Wait()
}

func goroutine1(rendezvousCh1 chan<- string, rendezvousCh2 <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	message := "Hello from Goroutine 1!"
	rendezvousCh1 <- message    // отправка
	response := <-rendezvousCh2 // приём
	fmt.Println("Goroutine 1:", response)
}

func goroutine2(rendezvousCh1 <-chan string, rendezvousCh2 chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	message := <-rendezvousCh1 // приём
	fmt.Println("Goroutine 2:", message)
	response := "Hello from Goroutine 2!"
	rendezvousCh2 <- response // отправка
}
