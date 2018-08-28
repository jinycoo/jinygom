package msmq

import (
	"github.com/streadway/amqp"
	"fmt"
)

var (
	Conn *amqp.Connection
	Ch   *amqp.Channel
)

func Init() {
	var err error
	Conn, err = amqp.Dial("amqp://admin:admin@192.168.0.92:5672/")
	if err != nil {
		fmt.Println(err)
	}
	defer Conn.Close()
	Ch, err = Conn.Channel()
	if err != nil {
		fmt.Println(err)
	}
	defer Ch.Close()
}

func Close() {

}