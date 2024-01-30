package main

import (
	"StudentManagementSystem/module"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	db, err := module.InitDB()
	if err != nil {
		log.Fatalf("数据库初始化失败，err:%v\n", err)
		return
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalf("数据库关闭失败，err:%v\n", err)
		}
	}(db)

	// 创建日志文件
	file, err := os.OpenFile("student_management_system.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666) // 如果文件不存在就创建它，然后以只写模式打开，且写入的数据追加到文件末尾
	if err != nil {
		log.Fatal(err)
	}
	//logger = log.New(file, "<New>", log.Lshortfile|log.Ldate|log.Ltime)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("日志关闭失败，err: %v\n", err)
		}
	}(file)

	// 设置日志输出到文件
	log.SetOutput(file)

	// 设置静态文件服务
	fs := http.FileServer(http.Dir("./module/templates"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", module.LoginHandler)
	http.HandleFunc("/home", module.HomeHandler)
	http.HandleFunc("/query", module.QueryRowHandler)
	http.HandleFunc("/insert", module.InsertRowHandler)
	http.HandleFunc("/update", module.UpdateRowHandler)
	http.HandleFunc("/delete", module.DeleteRowHandler)
	http.HandleFunc("/register", module.RegisterStudentHandler)
	http.HandleFunc("/studentPage", module.StudentHandler)

	fmt.Println("学生管理系统运行在： http://127.0.0.1:8080， 按 CTRL + C 退出系统。")
	err = http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		log.Fatal("发生错误:", err)
	}
}
