package main

import (
	"fmt"
	"sync"
	"time"
)

type Request struct {
	OrderID int
	Amount  float64
	RespCh  chan Response // канал для ответа
}

type Response struct {
	OrderID int
	Status  string
}

// OrderProcessor — контекст обработки заказов.
type OrderProcessor struct {
	id         int
	paymentReq chan Request // канал для отправки запросов в PaymentService
	shutdown   chan struct{}
	wg         *sync.WaitGroup
}

// PaymentService — контекст обработки платежей.
type PaymentService struct {
	id       int
	requests chan Request // входящие запросы
	shutdown chan struct{}
	wg       *sync.WaitGroup
}

func NewOrderProcessor(id int, paymentReq chan Request, wg *sync.WaitGroup) *OrderProcessor {
	return &OrderProcessor{
		id:         id,
		paymentReq: paymentReq,
		shutdown:   make(chan struct{}),
		wg:         wg,
	}
}

func (op *OrderProcessor) Start() {
	defer op.wg.Done()
	fmt.Printf("OrderProcessor %d started\n", op.id)

	// Имитация поступления заказов
	orders := []int{101, 102, 103}
	for _, orderID := range orders {
		// Создаём канал для ответа конкретно на этот запрос
		respCh := make(chan Response)
		req := Request{OrderID: orderID, Amount: float64(orderID) * 10.0, RespCh: respCh}

		fmt.Printf("OrderProcessor %d: sending order %d for payment\n", op.id, orderID)

		select {
		case op.paymentReq <- req:
			// Ждём ответ от платёжного сервиса
			select {
			case resp := <-respCh:
				fmt.Printf("OrderProcessor %d: received payment response for order %d: %s\n", op.id, resp.OrderID, resp.Status)
			case <-op.shutdown:
				return
			}
		case <-op.shutdown:
			return
		}
	}

	fmt.Printf("OrderProcessor %d finished sending orders\n", op.id)
}

func (op *OrderProcessor) Stop() {
	close(op.shutdown)
}

func NewPaymentService(id int, wg *sync.WaitGroup) *PaymentService {
	return &PaymentService{
		id:       id,
		requests: make(chan Request),
		shutdown: make(chan struct{}),
		wg:       wg,
	}
}

func (ps *PaymentService) Start() {
	defer ps.wg.Done()
	fmt.Printf("PaymentService %d started\n", ps.id)

	for {
		select {
		case req := <-ps.requests:
			fmt.Printf("PaymentService %d: processing payment for order %d, amount %.2f\n", ps.id, req.OrderID, req.Amount)
			time.Sleep(500 * time.Millisecond) // имитация работы
			status := "success"
			if req.Amount > 1000 {
				status = "declined"
			}
			// Отправляем ответ в канал, который предоставил отправитель
			req.RespCh <- Response{OrderID: req.OrderID, Status: status}
		case <-ps.shutdown:
			fmt.Printf("PaymentService %d shutting down\n", ps.id)
			return
		}
	}
}

func (ps *PaymentService) Stop() {
	close(ps.shutdown)
}

func main() {
	var wg sync.WaitGroup

	// Канал для связи OrderProcessor -> PaymentService
	paymentReqCh := make(chan Request)

	// Создаём контексты
	ps := NewPaymentService(1, &wg)
	op := NewOrderProcessor(1, paymentReqCh, &wg)

	// Подключаем PaymentService к каналу запросов
	ps.requests = paymentReqCh

	// Запускаем контексты
	wg.Add(2)
	go ps.Start()
	go op.Start() // Останавливаем контексты

	time.Sleep(3 * time.Second)

	op.Stop()
	ps.Stop()

	wg.Wait()
	fmt.Println("All contexts shut down")
}
