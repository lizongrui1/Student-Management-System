package module

import (
	"bytes"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"time"
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
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(message),
		})
	if err != nil {
		return nil, fmt.Errorf("发送消息失败: %w", err)
	}
	log.Printf(" [x] Sent %s", message)
	return ch, nil
}

func Worker(conn *amqp.Connection, chMsg chan string) {
	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("通道创建失败，err：%s\n", err)
		return
	}
	q, err := ch.QueueDeclare(
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
	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Printf("ch.Consume 创建失败，err: %s\n", err)
		return
	}
	forever := make(chan bool)
	go func() {
		for m := range msgs {
			log.Printf("已接收到消息： %s\n", m.Body)
			dotCount := bytes.Count(m.Body, []byte("."))
			t := time.Duration(dotCount)
			time.Sleep(t * time.Second)
			log.Printf("Done")
			m.Ack(false)
			chMsg <- string(m.Body)
		}
	}()
	<-forever
}
