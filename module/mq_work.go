package module

import (
	"fmt"
	"github.com/streadway/amqp"
)

func publishWorker(conn *amqp.Connection, message string) (*amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("创建通道失败: %w", err)
	}
	q, err := ch.QueueDeclare(
		"work",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Printf("队列声明失败，err：%s\n", err)
		return nil, err
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

func Worker(conn *amqp.Connection, chMsg chan string) {
	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("通道创建失败，err：%s\n", err)
		return
	}
	msgs, err := ch.QueueDeclare(
		"work",
		true,
		false,
		false,
		false,
		nil)
	err = ch.Qos(
		1,
		0,
		false,
	)
	if err != nil {
		fmt.Printf("ch.Qos创建失败，err：%s\n", err)
		return
	}
	msgs, err = ch.Consume()
}
