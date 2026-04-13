package main

import (
	"fmt"
	"sync"
	"time"
)

// Barrier представляет многоразовый барьер для синхронизации горутин.
type Barrier struct {
	size       int        // общее число участников
	arrived    int        // сколько уже прибыло
	mu         sync.Mutex // защита arrived
	cond       *sync.Cond // ожидание / уведомление
	generation int        // поколение (фаза)
}

// NewBarrier создаёт новый барьер для n горутин.
func NewBarrier(n int) *Barrier {
	b := &Barrier{size: n}
	b.cond = sync.NewCond(&b.mu)
	return b
}

// Wait блокирует вызывающую горутину до тех пор, пока все n горутин не вызовут Wait.
// После разблокировки барьер автоматически сбрасывается для следующего использования.
func (b *Barrier) Wait() {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Запоминаем текущее поколение
	gen := b.generation

	b.arrived++
	if b.arrived == b.size {
		// Все прибыли — будим всех и сбрасываем счётчик для следующей фазы
		b.arrived = 0
		b.generation++
		b.cond.Broadcast()
		return
	}

	// Ждём, пока поколение не изменится (т.е. пока последняя горутина не вызовет Broadcast)
	for gen == b.generation {
		b.cond.Wait()
	}
}

func main() {
	const numWorkers = 5
	const numPhases = 3

	barrier := NewBarrier(numWorkers)

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func(id int) {
			defer wg.Done()
			for phase := 1; phase <= numPhases; phase++ {
				// Имитация работы в текущей фазе
				workTime := time.Duration(id+1) * 200 * time.Millisecond
				fmt.Printf("Worker %d начал фазу %d (работа %v)\n", id, phase, workTime)
				time.Sleep(workTime)
				fmt.Printf("Worker %d завершил работу в фазе %d, ждёт остальных...\n", id, phase)

				// Ждём на барьере
				barrier.Wait()

				fmt.Printf("Worker %d прошёл барьер фазы %d\n", id, phase)
			}
		}(i)
	}

	wg.Wait()
	fmt.Println("Все воркеры завершили все фазы")
}
