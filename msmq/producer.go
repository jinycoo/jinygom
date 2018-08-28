package msmq

import (
	"github.com/streadway/amqp"
	"fmt"
)

func MsgPublish(conn *amqp.Connection) {
	ch, err := conn.Channel()
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()
	defer ch.Close()
	err = ch.Publish(
		"",
		"",
		false,
		false,
		amqp.Publishing{

		},
	)
}