package main

import (
	"fmt"
	"sync"
	"time"
)

// generator создаёт канал и отправляет в него числа из слайса.
// Этап 1: источник данных.
func generator(nums []int) <-chan int {
	out := make(chan int)
	go func() {
		for _, n := range nums {
			out <- n
		}
		close(out)
	}()
	return out
}

// worker выполняет обработку одного числа и возвращает результат.
// Каждый воркер будет запущен в отдельной горутине.
func worker(id int, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			// Имитация работы
			fmt.Printf("worker %d processing %d\n", id, n)
			time.Sleep(300 * time.Millisecond)
			out <- n * 2 // пример обработки
		}
	}()
	return out
}

// fanIn объединяет несколько каналов в один.
func fanIn(channels ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup

	// Функция пересылки данных из одного канала в out
	forward := func(ch <-chan int) {
		defer wg.Done()
		for val := range ch {
			out <- val
		}
	}

	wg.Add(len(channels))
	for _, ch := range channels {
		go forward(ch)
	}

	// Закрываем out после завершения всех горутин пересылки
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func main() {
	// Входные данные
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// Этап 1: Генератор
	inputCh := generator(numbers)

	// Этап 2: Fan-Out — запускаем несколько воркеров, каждый получает свой выходной канал
	numWorkers := 3
	workerChannels := make([]<-chan int, numWorkers)
	for i := 0; i < numWorkers; i++ {
		workerChannels[i] = worker(i+1, inputCh)
	}

	// Этап 3: Fan-In — объединяем результаты всех воркеров в один канал
	outputCh := fanIn(workerChannels...)

	// Этап 4: Потребитель — читаем и выводим результаты
	for result := range outputCh {
		fmt.Println("Result:", result)
	}

	fmt.Println("All done.")
}
