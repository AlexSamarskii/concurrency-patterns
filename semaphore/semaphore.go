package main

import (
	"fmt"
	"sync"
	"time"
)

type Semaphore struct {
	count int // текущее количество доступных разрешений
	mutex sync.Mutex
	cond  *sync.Cond // условная переменная для ожидания/пробуждения
}

func NewSemaphore(initialCount int) *Semaphore {
	s := &Semaphore{count: initialCount}
	s.cond = sync.NewCond(&s.mutex)
	return s
}

func (s *Semaphore) Acquire() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for s.count <= 0 {
		s.cond.Wait()
	}
	s.count--
}

func (s *Semaphore) Release() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.count++
	s.cond.Signal()
}

type ChanSemaphore chan struct{}

func NewChanSemaphore(limit int) ChanSemaphore {
	return make(ChanSemaphore, limit)
}

func (s ChanSemaphore) Acquire() {
	s <- struct{}{}
}

func (s ChanSemaphore) Release() {
	<-s
}

func (s ChanSemaphore) TryAcquire() bool {
	select {
	case s <- struct{}{}:
		return true
	default:
		return false
	}
}

func ExampleTask(id int, semaphore *Semaphore) {
	fmt.Printf("Задача %d ожидает семафор\n", id)
	semaphore.Acquire()
	defer semaphore.Release()

	fmt.Printf("Задача %d захватила семафор\n", id)
	time.Sleep(time.Second) // имитация работы
	fmt.Printf("Задача %d освободила семафор\n", id)
}

func main() {
	semaphore := NewSemaphore(2)

	var wg sync.WaitGroup
	wg.Add(4)

	for i := 1; i <= 4; i++ {
		go func(id int) {
			defer wg.Done()
			ExampleTask(id, semaphore)
		}(i)
	}

	wg.Wait()
}
