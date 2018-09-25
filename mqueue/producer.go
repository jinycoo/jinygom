package mqueue

import (
	"sync"
	"github.com/streadway/amqp"
)

type Producer interface {
	Produce(data []byte)
}

type producer struct {
	sync.Mutex
	workerStatus

	channel         *amqp.Channel
	errorChannel    chan<- error
	exchange        string
	mandatory       bool
	immediate       bool
	options         Options
	publishChannel  chan []byte
	routingKey      string
	shutdownChannel chan struct{}
}

func newProducer(channel *amqp.Channel, errorChannel chan<- error, config ProducerConfig) *producer {
	return &producer{
		channel:         channel,
		errorChannel:    errorChannel,
		exchange:        config.Exchange,
		options:         config.Options,
		mandatory:       config.Mandatory,
		immediate:       config.Immediate,
		publishChannel:  make(chan []byte, config.BufferSize),
		routingKey:      config.RoutingKey,
		shutdownChannel: make(chan struct{}),
	}
}

func (producer *producer) worker() {
	producer.markAsRunning()

	for {
		select {
		case message := <-producer.publishChannel:
			err := producer.produce(message)
			if err != nil {
				producer.errorChannel <- err
				// TODO Resend message.
			}
		case <-producer.shutdownChannel:
			// TODO It is necessary to guarantee the message delivery order.
			producer.closeChannel()

			return
		}
	}
}

func (producer *producer) setChannel(channel *amqp.Channel) {
	producer.Lock()
	producer.channel = channel
	producer.Unlock()
}

func (producer *producer) closeChannel() {
	producer.Lock()
	if err := producer.channel.Close(); err != nil {
		producer.errorChannel <- err
	}
	producer.Unlock()
}

func (producer *producer) Produce(message []byte) {
	producer.publishChannel <- message
}

func (producer *producer) produce(message []byte) error {
	producer.Lock()
	defer producer.Unlock()

	var msg =  amqp.Publishing{}
	if pub, ok := producer.options["contentType"]; ok {
		msg.ContentType = pub.(string)
	} else {
		msg.ContentType = "application/json"
	}
	if pub, ok := producer.options["deliveryMode"]; ok {
		msg.DeliveryMode = uint8(pub.(int))
	} else {
		msg.DeliveryMode = 1
	}
	msg.Body = message
	return producer.channel.Publish(producer.exchange, producer.routingKey, producer.mandatory, producer.immediate, msg)
}

func (producer *producer) Stop() {
	if producer.markAsStoppedIfCan() {
		producer.shutdownChannel <- struct{}{}
	}
}