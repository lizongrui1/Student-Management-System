package module

import (
	"context"
	"database/sql"
<<<<<<< HEAD
<<<<<<< HEAD
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"log"
	"os"
=======
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"log"
>>>>>>> 7fee29fce5c6d0e2c2bb376910a3d3b621e5ec1f
=======
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"log"
>>>>>>> 7fee29fce5c6d0e2c2bb376910a3d3b621e5ec1f
	"time"
)

//var tpl *template.Template

func InitDB() (*sql.DB, error) {
<<<<<<< HEAD
<<<<<<< HEAD
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

=======
=======
>>>>>>> 7fee29fce5c6d0e2c2bb376910a3d3b621e5ec1f
	//tpl = template.Must(template.ParseGlob("./module/templates/*.html"))
	db, err := sql.Open("mysql", "root:12345678@tcp(127.0.0.1:3306)/studb?charset=utf8")
	if err != nil {
		return nil, err
	}
<<<<<<< HEAD
>>>>>>> 7fee29fce5c6d0e2c2bb376910a3d3b621e5ec1f
=======
>>>>>>> 7fee29fce5c6d0e2c2bb376910a3d3b621e5ec1f
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
<<<<<<< HEAD
<<<<<<< HEAD

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
=======
	return nil
}
>>>>>>> 7fee29fce5c6d0e2c2bb376910a3d3b621e5ec1f
=======
	return nil
}
>>>>>>> 7fee29fce5c6d0e2c2bb376910a3d3b621e5ec1f
