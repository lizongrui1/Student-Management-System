package module

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

func emit(conn *amqp.Connection, message string) (*amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("创建通道失败: %w", err)
	}
	defer ch.Close()
	err = ch.ExchangeDeclare(
		"logs",
		"fanout",
		true,
		false,
		false,
		false,
		nil)
	failOnError(err, "创建交换器失败")
	err = ch.Publish(
		"logs",
		"",
		false,
		false,
		amqp.Publishing{ContentType: "text/plain", Body: []byte(message)})
	failOnError(err, "发送消息失败")
	return ch, nil
}

func receive(conn *amqp.Connection, chMsg chan string) {
	ch, err := conn.Channel()
	if err != nil {
		fmt.Errorf("创建通道失败: %w", err)
		return
	}
	defer ch.Close()
	err = ch.ExchangeDeclare(
		"logs",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "创建交换器失败")
	q, err := ch.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		fmt.Printf("队列声明失败，err：%s\n", err)
		return
	}
	err = ch.QueueBind(
		q.Name,
		"",
		"logs",
		false,
		nil,
	)
	failOnError(err, "绑定队列失败")
	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "消费者创建失败")
	forever := make(chan bool)
	go func() {
		for m := range msgs {
			log.Printf(" [x] %s", m.Body)
		}
	}()
	<-forever
}
