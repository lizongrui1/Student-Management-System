package module

import (
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestQueryRow(t *testing.T) {
	db, err := sql.Open("mysql", "root:12345678@tcp(127.0.0.1:3306)/studb?charset=utf8")
	if err != nil {
		t.Fatalf("无法连接数据库: %s", err)
	}
	defer db.Close()

	expectedStudent := Student{
		Number: 2000,
		Name:   "李一",
		Score:  90,
	}

	student, err := queryRow(2000)
	if err != nil {
		t.Fatalf("queryRow 返回了一个错误: %s", err)
	}
	if student != expectedStudent {
		t.Errorf("预期得到的学生: %v, 实际得到的学生: %v", expectedStudent, student)
	}
}

func TestInsertRow(t *testing.T) {
	db, err := sql.Open("mysql", "root:12345678@tcp(127.0.0.1:3306)/studb?charset=utf8")
	if err != nil {
		t.Fatalf("无法连接数据库: %s", err)
	}
	defer db.Close()

	want := Student{
		Number: 2010,
		Name:   "张三",
		Score:  60,
	}
	insertRow(2010, "张三", 60)
	student, err := queryRow(2010)
	if err != nil {
		t.Fatalf("queryRow 返回了一个错误: %s", err)
	}
	if student != want {
		t.Errorf("预期得到的学生: %v, 实际得到的学生: %v", want, student)
	}
}
