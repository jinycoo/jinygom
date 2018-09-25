package mqueue

import (
	"sync"
	"github.com/streadway/amqp"
)

type (
	ConsumerHandler func(message []byte)
    Consumer interface { Consume(handler ConsumerHandler) }
)

type (
	worker struct {
		sync.Mutex
		workerStatus

		channel         *amqp.Channel
		deliveries      <-chan amqp.Delivery
		errorChannel    chan<- error
		shutdownChannel chan struct{}
	}
	consumer struct {
		handler ConsumerHandler
		once    sync.Once
		workers []*worker

		queue         string
		name          string
		options       Options
		prefetchCount int
		prefetchSize  int
	}
)


func newConsumer(config ConsumerConfig) *consumer {
	return &consumer{
		workers: make([]*worker, config.Workers),
		queue:   config.Queue,
		name:    config.Name,
		options: config.Options,
	}
}

func (consumer *consumer) Consume(handler ConsumerHandler) {
	consumer.once.Do(func() {
		consumer.handler = handler

		for _, worker := range consumer.workers {
			go worker.Run(handler)
		}
	})
}

func (consumer *consumer) Stop() {
	for _, worker := range consumer.workers {
		worker.Stop()
	}
}

func newWorker(errorChannel chan<- error) *worker {
	return &worker{
		errorChannel:    errorChannel,
		shutdownChannel: make(chan struct{}),
	}
}

func (worker *worker) Run(handler ConsumerHandler) {
	worker.markAsRunning()

	for {
		select {
		case message := <-worker.deliveries:
			if message.Acknowledger == nil {
				if worker.markAsStoppedIfCan() {
					return
				}
				continue
			}

			handler(message.Body)
		case <-worker.shutdownChannel:
			worker.closeChannel()

			return
		}
	}
}

func (worker *worker) setChannel(channel *amqp.Channel) {
	worker.Lock()
	worker.channel = channel
	worker.Unlock()
}

func (worker *worker) closeChannel() {
	worker.Lock()
	if err := worker.channel.Close(); err != nil {
		worker.errorChannel <- err
	}
	worker.Unlock()
}

func (worker *worker) Stop() {
	if worker.markAsStoppedIfCan() {
		worker.shutdownChannel <- struct{}{}
	}
}