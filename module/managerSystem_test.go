package module

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

//func TestQueryRow(t *testing.T) {
//	db, err := sql.Open("mysql", "root:12345678@tcp(127.0.0.1:3306)/studb?charset=utf8")
//	if err != nil {
//		t.Fatalf("无法连接数据库: %s", err)
//	}
//	defer db.Close()
//
//	expectedStudent := Student{
//		Number: 2000,
//		Name:   "李一",
//		Score:  90,
//	}
//
//	student, err := queryRow(2000)
//	if err != nil {
//		t.Fatalf("queryRow 返回了一个错误: %s", err)
//	}
//	if student != expectedStudent {
//		t.Errorf("预期得到的学生: %v, 实际得到的学生: %v", expectedStudent, student)
//	}
//}

func TestQueryRow(t *testing.T) {
	db, mock, err := sqlmock.New() // 创建模拟的数据库连接
	if err != nil {
		t.Fatalf("创建模拟数据库连接时发生错误: %s", err)
	}
	defer db.Close()

	// 设置期望的模拟行为
	rows := sqlmock.NewRows([]string{"number", "name", "score"}).
		AddRow(2000, "李一", 90)
	mock.ExpectQuery("SELECT number, name, score FROM sms WHERE number = ?").
		WithArgs(2000).
		WillReturnRows(rows)

	// 运行要测试的函数
	student, err := queryRow(2000)
	if err != nil {
		t.Fatalf("queryRow 返回了一个错误: %s", err)
	}

	// 检查结果
	expectedStudent := Student{Number: 2000, Name: "李一", Score: 90}
	if student != expectedStudent {
		t.Errorf("预期得到的学生: %v, 实际得到的学生: %v", expectedStudent, student)
	}

	// 确保所有期望的模拟操作都已执行
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("存在未满足的期望: %s", err)
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
