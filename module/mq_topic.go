package module

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

func emitTopic(conn *amqp.Connection, message string) (*amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("创建通道失败: %w", err)
	}
	err = ch.ExchangeDeclare(
		"stu_topic",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "创建交换器失败")
	err = ch.Publish(
		"stu_topic",
		"course.english",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message)})
	failOnError(err, "发送消息失败")
	return ch, nil
}

func ReceiveTopic(conn *amqp.Connection, chMsg chan string) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("创建通道失败: %w", err)
	}
	err = ch.ExchangeDeclare(
		"stu_topic",
		"topic",
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
		return nil
	}
	err = ch.QueueBind(
		q.Name,
		"*.english",
		"stu_topic",
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
			log.Printf("%s", m.Body)
			chMsg <- string(m.Body)
		}
	}()
	<-forever
	return nil
}
