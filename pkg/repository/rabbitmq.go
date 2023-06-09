package repository

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

func NewRabbitMQ(cfg Config) (*amqp.Connection, error) {
	conn := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		cfg.Login, cfg.Password, cfg.Host, cfg.Port)

	mq, err := amqp.Dial(conn)
	if err != nil {
		log.Fatal(err)
	}

	return mq, err
}
