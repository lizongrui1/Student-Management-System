package module

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

func publishMessage(conn *amqp.Connection, message string) (*amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("创建通道失败: %w", err)
	}

	q, err := ch.QueueDeclare(
		"hello",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("创建队列失败: %w", err)
	}
	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	if err != nil {
		return nil, fmt.Errorf("发送消息失败: %w", err)
	}
	return ch, nil
}

func ConsumerMessage(conn *amqp.Connection, chMsg chan string) {
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("通道创建失败：", err)
	}
	//defer ch.Close()

	msgs, err := ch.Consume(
		"hello",
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("消费者创建失败：", err)
	}
	forever := make(chan bool)
	go func() {
		for msg := range msgs {
			chMsg <- string(msg.Body)
		}
	}()

	<-forever
}
