package mqueue

import (
	"fmt"
	"net"
	"time"
	"io/ioutil"
	"sync/atomic"
	"gopkg.in/yaml.v2"
	"github.com/streadway/amqp"

	"github.com/jinygo/log"
)

const (
	statusReadyForReconnect int32 = iota
	statusReconnecting
)

const (
	statusStopped int32 = iota
	statusRunning
)

var (
	Mqueue MQ
	mqCfg  *MQConfig
)

type (
	MQ interface {
		GetConsumer(name string) (Consumer, error)
		SetConsumerHandler(name string, handler ConsumerHandler) error
		GetProducer(name string) (Producer, error)
		Error() <-chan error
		Close()
	}
)

type mq struct {
	conn            *amqp.Connection
	channel         *amqp.Channel
	config          *Config

	errorChannel         chan error
	internalErrorChannel chan error
	consumers       *consumersRegistry
	producers       *producersRegistry
	reconnectStatus int32
}

func Init(file string) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		log.Warn(file + "文件读取失败")
	}
	err = yaml.Unmarshal(buf, &mqCfg)
	if err != nil {
		log.Warn(file + "解析失败")
	}
	if mqCfg.Queues != nil && len(mqCfg.Queues) > 0 {
		if rc, ok := mqCfg.Queues["rabbit"]; ok {
			Mqueue, err = New(rc)
		}
	} else {
		log.Warn("队列配置参数缺失")
	}
}

func New(config *Config) (MQ, error) {
	config.normalize()
	mq := &mq{
		config:               config,
		errorChannel:         make(chan error),
		internalErrorChannel: make(chan error),
		consumers:            newConsumersRegistry(len(config.Consumers)),
		producers:            newProducersRegistry(len(config.Producers)),
	}
	if err := mq.connect(); err != nil {
		return nil, err
	}

	go mq.errorHandler()

	return mq, mq.initialSetup()
}

func (mq *mq) GetConsumer(name string) (consumer Consumer, err error) {
	consumer, ok := mq.consumers.Get(name)
	if !ok {
		err = fmt.Errorf("consumer '%s' is not registered. Check your configuration", name)
	}

	return
}

func (mq *mq) SetConsumerHandler(name string, handler ConsumerHandler) error {
	consumer, err := mq.GetConsumer(name)
	if err != nil {
		return err
	}

	consumer.Consume(handler)

	return nil
}

func (mq *mq) GetProducer(name string) (producer Producer, err error) {
	producer, ok := mq.producers.Get(name)
	if !ok {
		err = fmt.Errorf("producer '%s' is not registered. Check your configuration", name)
	}

	return
}

func (mq *mq) Error() <-chan error {
	return mq.errorChannel
}

func (mq *mq) Close() {
	mq.stopProducersAndConsumers()

	if mq.channel != nil {
		mq.channel.Close()
	}

	if mq.conn != nil {
		mq.conn.Close()
	}
}

func (mq *mq) connect() error {
	connection, err := amqp.Dial(mq.config.DSN)
	if err != nil {
		return err
	}

	channel, err := connection.Channel()
	if err != nil {
		connection.Close()
		return err
	}
	mq.conn = connection
	mq.channel = channel

	go mq.handleCloseEvent()

	return nil
}

func (mq *mq) handleCloseEvent() {
	err := <-mq.conn.NotifyClose(make(chan *amqp.Error))
	if err != nil {
		mq.internalErrorChannel <- err
	}
}

func (mq *mq) errorHandler() {
	for err := range mq.internalErrorChannel {
		select {
		case mq.errorChannel <- err:
		default:
		}
		mq.processError(err)
	}
}

func (mq *mq) processError(err interface{}) {
	switch err.(type) {
	case *net.OpError:
		go mq.reconnect()
	case *amqp.Error:
		rmqErr, _ := err.(*amqp.Error)
		if rmqErr.Server == false {
			go mq.reconnect()
		}
	default:
	}
}

func (mq *mq) initialSetup() error {
	if err := mq.setupExchanges(); err != nil {
		return err
	}

	if err := mq.setupQueues(); err != nil {
		return err
	}

	if err := mq.setupProducers(); err != nil {
		return err
	}

	return mq.setupConsumers()
}


func (mq *mq) setupAfterReconnect() error {
	if err := mq.setupExchanges(); err != nil {
		return err
	}

	if err := mq.setupQueues(); err != nil {
		return err
	}

	mq.producers.GoEach(func(producer *producer) {
		if err := mq.reconnectProducer(producer); err != nil {
			mq.internalErrorChannel <- err
		}
	})

	mq.consumers.GoEach(func(consumer *consumer) {
		if err := mq.reconnectConsumer(consumer); err != nil {
			mq.internalErrorChannel <- err
		}
	})

	return nil
}

func (mq *mq) setupExchanges() error {
	for _, config := range mq.config.Exchanges {
		if err := mq.declareExchange(config); err != nil {
			return err
		}
	}

	return nil
}

func (mq *mq) declareExchange(config ExchangeConfig) error {
	var durable, autoDelete, internal, noWait bool
	var args amqp.Table
	if op, ok := config.Options["durable"]; ok {
		durable = op.(bool)
	} else {
		durable = opts["durable"].(bool)
	}
	if op, ok := config.Options["autoDelete"]; ok {
		autoDelete = op.(bool)
	} else {
		autoDelete = opts["autoDelete"].(bool)
	}
	if op, ok := config.Options["internal"]; ok {
		internal = op.(bool)
	} else {
		internal = opts["internal"].(bool)
	}
	if op, ok := config.Options["noWait"]; ok {
		noWait = op.(bool)
	} else {
		noWait = opts["noWait"].(bool)
	}
	if op, ok := config.Options["args"]; ok {
		args = op.(amqp.Table)
	} else {
		args = nil
	}

	return mq.channel.ExchangeDeclare(config.Name, config.Type, durable, autoDelete, internal, noWait, args)
}

func (mq *mq) setupQueues() error {
	for _, config := range mq.config.Queues {
		if err := mq.declareQueue(config); err != nil {
			return err
		}
	}

	return nil
}

func (mq *mq) declareQueue(config QueueConfig) error {
	var durable, autoDelete, exclusive, noWait bool
	var args amqp.Table
	if op, ok := config.Options["durable"]; ok {
		durable = op.(bool)
	} else {
		durable = opts["durable"].(bool)
	}
	if op, ok := config.Options["autoDelete"]; ok {
		autoDelete = op.(bool)
	} else {
		autoDelete = opts["autoDelete"].(bool)
	}
	if op, ok := config.Options["exclusive"]; ok {
		exclusive = op.(bool)
	} else {
		exclusive = opts["exclusive"].(bool)
	}
	if op, ok := config.Options["noWait"]; ok {
		noWait = op.(bool)
	} else {
		noWait = opts["noWait"].(bool)
	}
	if op, ok := config.Options["args"]; ok {
		args = op.(amqp.Table)
	} else {
		args = nil
	}
	if _, err := mq.channel.QueueDeclare(config.Name, durable, autoDelete, exclusive, noWait, args); err != nil {
		return err
	}
	var boNoWait bool
	var boArgs amqp.Table
	if op, ok := config.BindingOptions["noWait"]; ok {
		boNoWait = op.(bool)
	} else {
		boNoWait = opts["noWait"].(bool)
	}
	if op, ok := config.BindingOptions["args"]; ok {
		boArgs = op.(amqp.Table)
	} else {
		boArgs = nil
	}
	return mq.channel.QueueBind(config.Name, config.RoutingKey, config.Exchange, boNoWait, boArgs)
}

func (mq *mq) setupProducers() error {
	for _, config := range mq.config.Producers {
		if err := mq.registerProducer(config); err != nil {
			return err
		}
	}

	return nil
}

func (mq *mq) registerProducer(config ProducerConfig) error {
	if _, ok := mq.producers.Get(config.Name); ok {
		return fmt.Errorf(`producer with name "%s" is already registered`, config.Name)
	}

	channel, err := mq.conn.Channel()
	if err != nil {
		return err
	}

	producer := newProducer(channel, mq.internalErrorChannel, config)

	go producer.worker()
	mq.producers.Set(config.Name, producer)

	return nil
}

func (mq *mq) reconnectProducer(producer *producer) error {
	channel, err := mq.conn.Channel()
	if err != nil {
		return err
	}

	producer.setChannel(channel)
	go producer.worker()

	return nil
}

func (mq *mq) setupConsumers() error {
	for _, config := range mq.config.Consumers {
		if err := mq.registerConsumer(config); err != nil {
			return err
		}
	}
	return nil
}

func (mq *mq) registerConsumer(config ConsumerConfig) error {
	if _, ok := mq.consumers.Get(config.Name); ok {
		return fmt.Errorf(`consumer with name "%s" is already registered`, config.Name)
	}
	if config.Workers == 0 {
		config.Workers = 1
	}

	consumer := newConsumer(config)
	consumer.prefetchCount = config.PrefetchCount
	consumer.prefetchSize = config.PrefetchSize

	for i := 0; i < config.Workers; i++ {
		worker := newWorker(mq.internalErrorChannel)

		if err := mq.initializeConsumersWorker(consumer, worker); err != nil {
			return err
		}

		consumer.workers[i] = worker
	}

	mq.consumers.Set(config.Name, consumer) // Workers will start after consumer.Consume method call.

	return nil
}

func (mq *mq) reconnectConsumer(consumer *consumer) error {
	for _, worker := range consumer.workers {
		if err := mq.initializeConsumersWorker(consumer, worker); err != nil {
			return err
		}

		go worker.Run(consumer.handler)
	}

	return nil
}

func (mq *mq) initializeConsumersWorker(consumer *consumer, worker *worker) error {
	channel, err := mq.conn.Channel()
	if err != nil {
		return err
	}

	if err := channel.Qos(consumer.prefetchCount, consumer.prefetchSize, false); err != nil {
		return err
	}
	var noLocal, autoAck, exclusive, noWait bool
	var args amqp.Table
	if op, ok := consumer.options["noLocal"]; ok {
		noLocal = op.(bool)
	} else {
		noLocal = opts["noLocal"].(bool)
	}
	if op, ok := consumer.options["noAck"]; ok {
		autoAck = op.(bool)
	} else {
		autoAck = opts["noAck"].(bool)
	}
	if op, ok := consumer.options["exclusive"]; ok {
		exclusive = op.(bool)
	} else {
		exclusive = opts["exclusive"].(bool)
	}
	if op, ok := consumer.options["noWait"]; ok {
		noWait = op.(bool)
	} else {
		noWait = opts["noWait"].(bool)
	}
	if op, ok := consumer.options["args"]; ok {
		args = op.(amqp.Table)
	} else {
		args = nil
	}
	deliveries, err := channel.Consume(consumer.queue, "", autoAck, exclusive, noLocal, noWait, args)
	if err != nil {
		return err
	}

	worker.setChannel(channel)
	worker.deliveries = deliveries

	return nil
}


func (mq *mq) reconnect() {
	notBusy := atomic.CompareAndSwapInt32(&mq.reconnectStatus, statusReadyForReconnect, statusReconnecting)
	if !notBusy {
		return
	}

	defer func() {
		atomic.StoreInt32(&mq.reconnectStatus, statusReadyForReconnect)
	}()

	time.Sleep(mq.config.ReconnectDelay)

	mq.stopProducersAndConsumers()

	if err := mq.connect(); err != nil {
		mq.internalErrorChannel <- err
		return
	}

	if err := mq.setupAfterReconnect(); err != nil {
		mq.internalErrorChannel <- err
	}
}

func (mq *mq) stopProducersAndConsumers() {
	mq.producers.GoEach(func(producer *producer) {
		producer.Stop()
	})

	mq.consumers.GoEach(func(consumer *consumer) {
		consumer.Stop()
	})
}

// Describes worker state: running or stopped and provides an ability to change it atomically.
type workerStatus struct {
	value int32
}

// markAsRunning changes status to running.
func (status *workerStatus) markAsRunning() {
	atomic.StoreInt32(&status.value, statusRunning)
}

// markAsStoppedIfCan changes status to stopped if current status is running.
// Returns true on success if status was changed to stopped
// or false if status is already changed to stopped.
func (status *workerStatus) markAsStoppedIfCan() bool {
	return atomic.CompareAndSwapInt32(&status.value, statusRunning, statusStopped)
}