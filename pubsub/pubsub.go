package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Message — типизированное сообщение.
type Message[T any] struct {
	Topic string
	Data  T
}

// Subscriber — подписчик с буферизированным каналом.
type Subscriber[T any] struct {
	ID      int
	Channel chan T
}

// Broker — брокер с поддержкой контекста.
type Broker[T any] struct {
	subscribers map[string][]*Subscriber[T]
	mu          sync.RWMutex
}

func NewBroker[T any]() *Broker[T] {
	return &Broker[T]{
		subscribers: make(map[string][]*Subscriber[T]),
	}
}

// Subscribe добавляет подписчика к теме.
func (b *Broker[T]) Subscribe(topic string, sub *Subscriber[T]) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[topic] = append(b.subscribers[topic], sub)
}

// Unsubscribe удаляет подписчика из темы и закрывает его канал.
func (b *Broker[T]) Unsubscribe(topic string, sub *Subscriber[T]) {
	b.mu.Lock()
	defer b.mu.Unlock()

	subs, ok := b.subscribers[topic]
	if !ok {
		return
	}
	for i, s := range subs {
		if s == sub {
			// Удаляем из слайса
			b.subscribers[topic] = append(subs[:i], subs[i+1:]...)
			close(sub.Channel)
			return
		}
	}
}

// Publish публикует сообщение всем подписчикам темы.
// Использует неблокирующую отправку с таймаутом.
func (b *Broker[T]) Publish(ctx context.Context, msg Message[T]) {
	b.mu.RLock()
	subs := b.subscribers[msg.Topic]
	subscribers := make([]*Subscriber[T], len(subs))
	copy(subscribers, subs)
	b.mu.RUnlock()

	for _, sub := range subscribers {
		timer := time.NewTimer(100 * time.Millisecond)
		select {
		case sub.Channel <- msg.Data:
			if !timer.Stop() {
				<-timer.C
			}
		case <-timer.C:
			fmt.Printf("WARN: subscriber %d on topic %s is slow, message dropped\n", sub.ID, msg.Topic)
		case <-ctx.Done():
			timer.Stop()
			return
		}
	}
}

// StartSubscriber запускает обработку сообщений подписчиком.
// Завершается при отмене контекста или закрытии канала.
func (s *Subscriber[T]) Start(ctx context.Context) {
	for {
		select {
		case msg, ok := <-s.Channel:
			if !ok {
				fmt.Printf("Subscriber %d: channel closed\n", s.ID)
				return
			}
			fmt.Printf("Subscriber %d received: %v\n", s.ID, msg)
		case <-ctx.Done():
			fmt.Printf("Subscriber %d: context done\n", s.ID)
			return
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	broker := NewBroker[string]()

	// Создаём подписчиков с буфером 10
	sub1 := &Subscriber[string]{ID: 1, Channel: make(chan string, 10)}
	sub2 := &Subscriber[string]{ID: 2, Channel: make(chan string, 10)}

	broker.Subscribe("news", sub1)
	broker.Subscribe("news", sub2)
	broker.Subscribe("alerts", sub2)

	// Запускаем обработчики
	go sub1.Start(ctx)
	go sub2.Start(ctx)

	// Публикуем сообщения
	broker.Publish(ctx, Message[string]{Topic: "news", Data: "Hello, World!"})
	broker.Publish(ctx, Message[string]{Topic: "alerts", Data: "Disk space low"})

	// Дадим время на обработку
	time.Sleep(1 * time.Second)

	// Отписываем sub1 от "news" (канал закроется)
	broker.Unsubscribe("news", sub1)

	// Ещё одно сообщение в "news" — получит только sub2
	broker.Publish(ctx, Message[string]{Topic: "news", Data: "Second news"})

	time.Sleep(500 * time.Millisecond)

	// Отменяем контекст, чтобы остановить подписчиков
	cancel()
	time.Sleep(100 * time.Millisecond)

	fmt.Println("Program finished")
}
