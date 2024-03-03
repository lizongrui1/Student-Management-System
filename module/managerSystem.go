package module

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/now"
	"github.com/redis/go-redis/v9"
	"log"
	"strconv"
	"strings"
	"time"
)

var db, _ = InitDB()

var rdb *redis.Client
var ctx = context.Background()

type myUsualType interface{}

type Student struct {
	Number int    `json:"number"`
	Name   string `json:"name"`
	Score  int    `json:"score"`
}

// 投票功能
func getKey(tid int64) string {
	return fmt.Sprintf("teacher:like:%d", tid)
}

// tid 需要点赞教师的ID   id 学生ID
func GiveLike(ctx context.Context, tid int64, id int64) (bool, error) {
	keys := getKey(tid)
	res, err := rdb.GetBit(ctx, keys, (id - 1)).Result()
	if err != nil {
		return false, err
	}

	if res == 1 {
		return true, nil
	}

	_, err = rdb.SetBit(ctx, keys, (id - 1), 1).Result()
	if err != nil {
		return false, err
	}

	return true, nil
}

// 查询是否已经点赞了
func GiveLikeSelect(tid int64, id int64) (bool, error) {
	var keys = getKey(tid)
	res, err := rdb.GetBit(context.Background(), keys, (id - 1)).Result()
	if err != nil {
		return false, err
	}

	if res == 1 {
		return true, nil
	}

	return false, nil
}

// 点赞数量
func GiveLikeCount(tid int64) (int64, error) {
	var keys = getKey(tid)
	count := redis.BitCount{Start: 0, End: -1}
	return rdb.BitCount(context.Background(), keys, &count).Result()
}

func obtainStudent(ctx context.Context, db *sql.DB) ([]string, error) {
	var names []string
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	rows, err := db.QueryContext(timeoutCtx, "SELECT name FROM sms")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Fatal(err)
			return nil, err
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
		return nil, err
	}
	return names, nil
}

func studentsScore(ctx context.Context, db *sql.DB, rdb *redis.Client) ([]Student, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(timeoutCtx, "SELECT number, name, score FROM sms")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var number int
		var name string
		var score float64
		if err := rows.Scan(&number, &name, &score); err != nil {
			log.Fatal(err)
			return nil, err
		}
		member := fmt.Sprintf("%d:%s", number, name)
		if err := rdb.ZAdd(timeoutCtx, "students", redis.Z{
			Score:  score,
			Member: member,
		}).Err(); err != nil {
			log.Fatal(err)
			return nil, err
		}
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
		return nil, err
	}
	results, err := rdb.ZRangeWithScores(ctx, "students", 0, -1).Result()
	if err != nil {
		log.Printf("获取学生分数失败, err:%v\n", err)
		return nil, err
	}
	var students []Student
	for _, result := range results {
		split := strings.Split(result.Member.(string), ":")
		if len(split) < 2 {
			log.Printf("成员格式错误: %v\n", result.Member)
			continue
		}
		number, err := strconv.Atoi(split[0])
		if err != nil {
			log.Printf("转换编号失败, err:%v\n", err)
			continue
		}
		students = append(students, Student{
			Number: number,
			Name:   split[1],
			Score:  int(result.Score),
		})
	}
	return students, err
}

func register(number string, password string) (err error) {
	time.Now()
	ret, err := db.Exec("INSERT INTO stu (student_id, password) VALUES (?, ?)", number, password)
	if err != nil {
		log.Printf("学生账号添加失败: %v\n", err)
		return
	}
	newID, err := ret.LastInsertId()
	if err != nil {
		log.Printf("新注册学生ID失败: %v\n", err)
	}
	log.Printf("%s注册成功, 新注册的学生学号为：%d\n", now.BeginningOfMinute(), newID)
	return
}

// 查看学生
//
//	func queryRow(number int) (student Student, err error) {
//		var stu Student
//		err = db.QueryRow("SELECT number, name, score FROM sms WHERE number = ?", number).Scan(&stu.Number, &stu.Name, &stu.Score)
//		if err != nil {
//			fmt.Printf("查询失败, err: %v\n", err)
//			return
//		}
//		return stu, nil
//	}
func queryRow(number int) (student Student, err error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	val, err := rdb.Get(timeoutCtx, fmt.Sprintf("student:%d", number)).Result()
	if err == redis.Nil {
		err = db.QueryRow("SELECT number, name, score FROM sms WHERE number = ?", number).Scan(&student.Number, &student.Name, &student.Score)
		if err != nil {
			log.Fatalf("查询失败, err: %v\n", err)
			return
		}
		studentJSON, _ := json.Marshal(student)
		err = rdb.Set(timeoutCtx, fmt.Sprintf("student:%d", number), studentJSON, 30*time.Minute).Err()
		if err != nil {
			log.Printf("缓存设置失败, err: %v\n", err)
		}
		return student, nil
	} else if err != nil {
		log.Printf("从Redis查询失败, err: %v\n", err)
		return
	} else {
		err = json.Unmarshal([]byte(val), &student)
		if err != nil {
			log.Printf("反序列化失败, err: %v\n", err)
			return
		}
		return student, nil
	}
}

// 全部查看
func queryMultiRow() ([]Student, error) {
	var students []Student
	ret, err := db.Query("SELECT number, name, score FROM sms")
	if err != nil {
		log.Printf("查询失败, err:%v\n", err)
		return nil, err
	}
	defer ret.Close()
	for ret.Next() {
		var stu Student
		err := ret.Scan(&stu.Number, &stu.Name, &stu.Score)
		if err != nil {
			log.Printf("赋值失败, err:%v\n", err)
			continue
		}
		students = append(students, stu)
	}
	if err := ret.Err(); err != nil {
		log.Printf("迭代失败, err:%v\n", err)
		return nil, err
	}
	return students, nil
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
	log.Printf("%s 加入成功, 新加入的学生序号为：%d\n", currentTime.Format("2006/01/02 15:04:05"), insertedId)
	return
}

// 修改学生
func updateRow(number int, newScore myUsualType) (err error) {
	sqlStr := "UPDATE sms SET score = ? WHERE number = ?"
	ret, err := db.Exec(sqlStr, newScore, number)
	if err != nil {
		log.Fatalf("更新失败, error: %v\n", err)
		return
	}
	rowsAffected, err := ret.RowsAffected()
	if err != nil {
		log.Printf("获取更新行数时发生错误: %v\n", err)
		return
	}
	if rowsAffected == 0 {
		fmt.Println("没有找到对应的学号, 未进行更新")
		return
	}
	fmt.Printf("更新成功, 受影响行数:%d\n", rowsAffected)
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	student, err := queryRow(number)
	if err != nil {
		log.Printf("从数据库获取更新后的学生信息失败, err: %v\n", err)
		return
	}

	studentJSON, err := json.Marshal(student)
	if err != nil {
		log.Printf("序列化学生信息失败, err: %v\n", err)
		return
	}

	err = rdb.Set(timeoutCtx, fmt.Sprintf("student:%d", number), studentJSON, 30*time.Minute).Err()
	if err != nil {
		log.Printf("更新Redis缓存失败, err: %v\n", err)
		return
	}
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
