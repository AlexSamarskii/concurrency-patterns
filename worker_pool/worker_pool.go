package main

import (
	"fmt"
	"sync"
	"time"
)

// Task — единица работы.
type Task struct {
	ID   int
	Data string
}

// Result — результат выполнения задачи.
type Result struct {
	TaskID int
	Output string
	Err    error
}

func main() {
	const numWorkers = 3

	// Буферизированные каналы для задач и результатов.
	// Размер буфера можно подобрать под ожидаемую нагрузку.
	taskCh := make(chan Task, 10)
	resultCh := make(chan Result, 10)

	var wg sync.WaitGroup

	// Запускаем воркеры.
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(i, taskCh, resultCh, &wg)
	}

	// В отдельной горутине ждём завершения воркеров и закрываем канал результатов.
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Генерируем и отправляем задачи.
	tasks := generateTasks()
	for _, task := range tasks {
		taskCh <- task
	}
	close(taskCh) // Закрываем канал задач — сигнал воркерам, что больше задач не будет.

	// Читаем и обрабатываем результаты.
	for res := range resultCh {
		if res.Err != nil {
			fmt.Printf("Ошибка в задаче %d: %v\n", res.TaskID, res.Err)
		} else {
			fmt.Printf("Результат задачи %d: %s\n", res.TaskID, res.Output)
		}
	}

	fmt.Println("Все задачи обработаны.")
}

// worker выполняет задачи из taskCh и отправляет результаты в resultCh.
func worker(id int, taskCh <-chan Task, resultCh chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("Воркер %d запущен\n", id)

	for task := range taskCh {
		// Имитация работы.
		time.Sleep(500 * time.Millisecond)

		// Обработка задачи.
		output := fmt.Sprintf("обработан воркером %d: %s", id, task.Data)

		// Отправка результата.
		resultCh <- Result{
			TaskID: task.ID,
			Output: output,
		}
	}

	fmt.Printf("Воркер %d завершил работу\n", id)
}

func generateTasks() []Task {
	return []Task{
		{ID: 1, Data: "Загрузить файл"},
		{ID: 2, Data: "Отправить email"},
		{ID: 3, Data: "Сгенерировать отчёт"},
		{ID: 4, Data: "Обновить кэш"},
		{ID: 5, Data: "Записать лог"},
	}
}
