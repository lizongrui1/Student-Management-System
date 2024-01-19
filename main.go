package main

import (
	"StudentManagementSystem/module"
	"fmt"
	"log"
	"net/http"
)

func main() {
	db, err := module.InitDB()
	if err != nil {
		log.Fatalf("数据库初始化失败: %v\n", err)
		return
	}
	defer db.Close()

	// 设置静态文件服务
	fs := http.FileServer(http.Dir("./module/templates"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// 设置路由
	http.HandleFunc("/", module.LoginHandler)
	http.HandleFunc("/home", module.HomeHandler)
	http.HandleFunc("/query", module.QueryRowHandler)
	http.HandleFunc("/insert", module.InsertRowHandler)
	http.HandleFunc("/update", module.UpdateRowHandler)
	http.HandleFunc("/delete", module.DeleteRowHandler)

	fmt.Println("学生管理系统运行在： http://127.0.0.1:8080， 按 CTRL + C 退出系统。")
	err = http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		log.Fatal("发生错误:", err)
	}
}
