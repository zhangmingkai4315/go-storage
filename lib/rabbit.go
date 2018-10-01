package lib

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// FailOnError will panic when err not nil
func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s:%s", msg, err)
		panic(fmt.Sprintf("%s:%s", msg, err))
	}
}

// RabbitMQ define the struct of rabbit infomation
type RabbitMQ struct {
	channel  *amqp.Channel
	Name     string
	exchange string
}

// NewRabbitMQ create a new RabbitMQ with server ip
func NewRabbitMQ(server string) *RabbitMQ {
	conn, err := amqp.Dial(server)
	FailOnError(err, "fail to connect to rabbitmq server")

	ch, err := conn.Channel()
	FailOnError(err, "fail to connect channgel")
	q, err := ch.QueueDeclare(
		"",
		false,
		true,
		false,
		false,
		nil,
	)
	FailOnError(err, "fail to declare queue")

	mq := new(RabbitMQ)
	mq.channel = ch
	mq.Name = q.Name
	return mq
}

// DeclareExchange will declare a new exchange
func (mq *RabbitMQ) DeclareExchange(exchange string, typeOfExchange string) {
	err := mq.channel.ExchangeDeclare(
		exchange,       // name
		typeOfExchange, // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	FailOnError(err, "declare exchange error")
	return
}

// Bind will bind the rabbitmq struct with new exchange
func (mq *RabbitMQ) Bind(exchange string) {
	err := mq.channel.QueueBind(
		mq.Name,
		"",
		exchange,
		false,
		nil,
	)
	FailOnError(err, "bind channel fail")
	mq.exchange = exchange
}

// Send will send the body infomation to queue
func (mq *RabbitMQ) Send(queue string, body interface{}) {
	str, err := json.Marshal(body)
	FailOnError(err, "json marshal error")
	err = mq.channel.Publish("", queue, false, false, amqp.Publishing{
		ReplyTo: mq.Name,
		Body:    []byte(str),
	})
	FailOnError(err, "publish message error")
}

// Publish will publish the new body to exchange
func (mq *RabbitMQ) Publish(exchange string, body interface{}) {
	str, err := json.Marshal(body)
	FailOnError(err, "json marshal error")
	err = mq.channel.Publish(
		exchange,
		"",
		false,
		false,
		amqp.Publishing{
			ReplyTo: mq.Name,
			Body:    []byte(str),
		},
	)
	FailOnError(err, "publish message error")
}

// Consume will create a new consumer channel
func (mq *RabbitMQ) Consume() <-chan amqp.Delivery {
	c, err := mq.channel.Consume(
		mq.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	FailOnError(err, "consume error")
	return c
}

// Close current rabbitmq connection
func (mq *RabbitMQ) Close() {
	mq.channel.Close()
}
