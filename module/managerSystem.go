package module

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"time"
)

var db, _ = InitDB()

type myUsualType interface{}

type Student struct {
	Number int
	Name   string
	Score  int
}

type StudentID struct {
	ID  int
	pwd string
}

func register(number string, password string) (err error) {
	currentTime := time.Now()
	ret, err := db.Exec("INSERT INTO stu (student_id, password) VALUES (?, ?)", number, password)
	if err != nil {
		log.Printf("学生账号添加失败: %v\n", err)
		return
	}
	newID, err := ret.LastInsertId()
	if err != nil {
		log.Printf("获取新注册学生ID失败: %v\n", err)
	}
	log.Printf("%s 注册成功, 新注册的学生学号为：%d\n", currentTime.Format("2006/01/02 15:04:05"), newID)
	return
}

// 查看学生
func queryRow(number int) (student Student, err error) {
	var stu Student
	err = db.QueryRow("SELECT number, name, score FROM sms WHERE number = ?", number).Scan(&stu.Number, &stu.Name, &stu.Score)
	if err != nil {
		fmt.Printf("查询失败, err: %v\n", err)
		return
	}
	return stu, nil
}

// 全部查看
func queryMultiRow() ([]Student, error) {
	rows, err := db.Query("SELECT * FROM sms")
	if err != nil {
		fmt.Printf("查询失败, err:%v\n", err)
		return nil, err
	}
	defer rows.Close()

	var students []Student // 创建一个空的Student切片用于存储学生数据

	for rows.Next() {
		var stu Student // 使用Student结构体来存储每行的数据
		err := rows.Scan(&stu.Number, &stu.Name, &stu.Score)
		if err != nil {
			fmt.Printf("赋值失败, err:%v\n", err)
			continue // 发生错误时继续处理下一行
		}
		students = append(students, stu) // 将学生对象添加到切片中
	}

	if err := rows.Err(); err != nil {
		fmt.Printf("iteration failed, err:%v\n", err)
		return nil, err
	}

	return students, nil // 返回学生切片和错误信息
}

// 增加学生
func insertRow(number int, name string, score int) (err error) {
	currentTime := time.Now()
	ret, err := db.Exec("INSERT INTO sms (number, name, score) VALUES (?, ?, ?)", number, name, score)
	if err != nil {
		fmt.Printf("添加失败, err:%v\n", err)
		return
	}
	insertedId, err := ret.LastInsertId()
	if err != nil {
		fmt.Printf("获取插入ID失败, err:%v\n", err)
		return
	}
	fmt.Printf("%s 加入成功, 新加入的学生序号为：%d\n", currentTime.Format("2006/01/02 15:04:05"), insertedId)
	return
}

// 修改学生
func updateRow(name string, newValue myUsualType) (err error) {
	sqlStr := "UPDATE sms SET score = ? WHERE name = ?"
	ret, err := db.Exec(sqlStr, newValue, name)
	if err != nil {
		fmt.Printf("更新失败, error: %v\n", err)
		return
	}
	rowsAffected, err := ret.RowsAffected()
	if err != nil {
		fmt.Printf("获取更新行数时发生错误: %v\n", err)
		return
	}
	if rowsAffected == 0 {
		fmt.Println("没有找到对应的ID, 未进行更新")
		return
	}
	fmt.Printf("更新成功, 受影响行数:%d\n", rowsAffected)
	return
}

// 删除学生
func deleteRow(number int) (err error) {
	currentTime := time.Now()
	// 首先检查学生是否存在
	_, err = queryRow(number)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("没有找到学号为 %d 的学生", number)
		}
		fmt.Printf("查询学生时出错, err: %v\n", err)
		return
	}
	ret, err := db.Exec("DELETE FROM sms WHERE number = ?", number)
	if err != nil {
		fmt.Printf("删除失败, err:%v\n", err)
		return
	}
	n, err := ret.RowsAffected()
	if err != nil {
		fmt.Printf("get RowsAffected failed, err:%v\n", err)
		return
	}
	if n == 0 {
		return fmt.Errorf("没有学号为 %d 的学生", number)
	}
	fmt.Printf("%s 删除成功, 删除的学生学号为：%d", currentTime.Format("2006/01/02 15:04:05"), number)
	return
}

func validate(username, password string) (bool, error) {
	var dbPassword string
	err := db.QueryRow("SELECT password FROM stu WHERE student_id = ?", username).Scan(&dbPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // 用户名不存在
		}
		return false, err // 数据库查询出错
	}
	if password == dbPassword {
		return true, nil
	}
	return false, nil
}
