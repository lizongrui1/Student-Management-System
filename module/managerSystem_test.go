package module

import (
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

// 此函数需要您提前设置好一个测试数据库，并确保测试用的数据存在
func TestQueryRow(t *testing.T) {
	db, err := sql.Open("mysql", "root:12345678@tcp(127.0.0.1:3306)/studb?charset=utf8")
	if err != nil {
		t.Fatalf("无法连接数据库: %s", err)
	}
	defer db.Close()

	expectedStudent := Student{
		Number: 1,
		Name:   "Test Student", // 这里填写实际的学生姓名
		Score:  100,            // 这里填写实际的分数
	}

	student, err := queryRow(1)
	if err != nil {
		t.Fatalf("queryRow 返回了一个错误: %s", err)
	}

	if student != expectedStudent {
		t.Errorf("预期得到的学生: %v, 实际得到的学生: %v", expectedStudent, student)
	}
}
