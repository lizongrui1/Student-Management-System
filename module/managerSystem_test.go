package module

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"testing"
)

func initDB() *sql.DB {
	//wd, err := os.Getwd()
	//if err != nil {
	//	log.Fatalf("无法获取当前工作目录: %s", err)
	//}
	//fmt.Println("当前工作目录:", wd)
	db, err := sql.Open("mysql", "root:12345678@tcp(127.0.0.1:3306)/studb?charset=utf8")
	if err != nil {
		log.Fatalf("无法连接数据库: %s", err)
	}
	return db
}

func setup(db *sql.DB) error {
	_, err := db.Exec("INSERT INTO sms (number, name, score) VALUES (9997, '测试1', 88), (9998, '测试2', 92), (9999, '测试3', 75)")
	return err
}

func deleteData(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM sms WHERE number IN (9997, 9998, 9999)")
	return err
}

func TestQueryRow(t *testing.T) {
	db := initDB()
	defer db.Close()
	if err := setup(db); err != nil {
		t.Fatalf("设置测试数据失败: %s", err)
	}
	defer deleteData(db)
	student, err := queryRow(9997)
	if err != nil {
		t.Fatalf("queryRow返回了一个错误: %s", err)
	}
	expectedStudent := Student{Number: 9997, Name: "测试1", Score: 88}
	if student != expectedStudent {
		t.Errorf("预期得到的学生: %v名, 实际得到的学生: %v名", expectedStudent, student)
	}
}

func TestQueryMultiRow(t *testing.T) {
	db := initDB()
	defer db.Close()
	if err := setup(db); err != nil {
		t.Fatalf("设置测试数据失败: %s", err)
	}
	defer deleteData(db)
	students, err := queryMultiRow()
	if err != nil {
		t.Fatalf("queryMultiRow 返回了一个错误: %s", err)
	}
	expectedStudents := map[int]Student{
		9997: {Number: 9997, Name: "测试1", Score: 88},
		9998: {Number: 9998, Name: "测试2", Score: 92},
		9999: {Number: 9999, Name: "测试3", Score: 75},
	}
	found := 0
	for _, student := range students {
		if expStu, ok := expectedStudents[student.Number]; ok {
			if expStu.Name != student.Name || expStu.Score != student.Score {
				t.Errorf("预期得到的学生信息: %v, 实际得到的学生信息: %v", expStu, student)
			}
			found++
		}
	}
	if found != len(expectedStudents) {
		t.Errorf("预期查询到的特定学生数量: %d, 实际查询到的特定学生数量: %d", len(expectedStudents), found)
	}
}
