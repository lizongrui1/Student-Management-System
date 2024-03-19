package module

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"log"
	"os"
	"time"
)

//var tpl *template.Template

func publishMessage(conn *amqp.Connection, message string) (*amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("创建通道失败: %w", err)
	}

	_, err = ch.QueueDeclare(
		"hello",
		false,
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
		"hello",
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
	defer ch.Close()

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
	go func() {
		for msg := range msgs {
			chMsg <- string(msg.Body)
		}
	}()
}

func InitDB() (*sql.DB, error) {
	Initconfig()
	username := viper.GetString("database.username")
	password := viper.GetString("database.password")
	host := viper.GetString("database.host")
	port := viper.GetString("database.port")
	database := viper.GetString("database.name")
	charset := viper.GetString("database.charset")

	//tpl = template.Must(template.ParseGlob("./module/templates/*.html"))
	//db, err := sql.Open("mysql", "root:12345678@tcp(127.0.0.1:3306)/studb?charset=utf8")
	//if err != nil {
	//	return nil, err
	//}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s", username, password, host, port, database, charset)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		err := db.Close()
		if err != nil {
			return nil, err
		}
		return nil, err
	}
	return db, nil
}

func InitRDB() error {
	rdb = redis.NewClient(&redis.Options{
		Network:  "tcp",
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Fatalf("无法连接到Redis: %v", err)
	}

	var err error
	db, err = InitDB()
	if err != nil {
		log.Fatalf("无法连接到MySQL数据库: %v", err)
	}

	go func() {
		for {
			select {
			case msg := <-messageChannel:
				// 存储消息到Redis
				err := rdb.Set(context.Background(), "messageKey", msg, 0).Err()
				if err != nil {
					log.Printf("存储消息失败: %v", err)
				}
			}
		}
	}()

	return nil
}

func Initconfig() {
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("初始化配置文件失败，err：%s", err.Error())
	}

	viper.SetConfigName("database")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(workDir + "/config")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败，err：%s", err.Error())
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("someFunction failed:%s,%s", msg, err)
	}
}
