package module

import (
	"context"
	"database/sql"
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
var tid int64

type myUsualType interface{}

type Student struct {
	Number int    `json:"number"`
	Name   string `json:"name"`
	Score  int    `json:"score"`
}

type UserSign struct {
}

func executeDBOperations(db *sql.DB, id int) {
	tx, err := db.Begin()
	if err != nil {
		fmt.Printf("事务初始错误: %v\n", err)
		return
	}
	defer tx.Rollback()

	//_, err = tx.Exec("INSERT INTO sms (column_names) VALUES (values)")
	//if err != nil {
	//	fmt.Printf("Exec insert error: %v\n", err)
	//	return
	//}

	_, err = tx.Exec("UPDATE sms SET `score` = 80 WHERE `id` = 26 AND `number` = 2416")
	if err != nil {
		fmt.Printf("Exec update error: %v\n", err)
		return
	}

	//_, err = tx.Exec("DELETE FROM sms WHERE id = ?", 25)
	//if err != nil {
	//	fmt.Printf("Exec delete error: %v\n", err)
	//	return
	//}

	err = tx.Commit()
	if err != nil {
		fmt.Printf("事务提交失败: %v\n", err)
		return
	}

	fmt.Printf("Goroutine %d 已完成\n", id)
}

func executeDBOperationsWithSharedLock(db *sql.DB, id int) {
	tx, err := db.Begin()
	if err != nil {
		fmt.Printf("事务初始错误: %v\n", err)
		return
	}
	defer tx.Rollback()

	//_, err = tx.Exec("INSERT INTO sms (column_names) VALUES (values)")
	//if err != nil {
	//	fmt.Printf("Exec insert error: %v\n", err)
	//	return
	//}

	_, err = tx.Exec("SELECT `score` FROM sms WHERE `id` = 26 AND `number` = 2416 LOCK IN SHARE MODE")
	if err != nil {
		fmt.Printf("Exec update error: %v\n", err)
		return
	}

	//_, err = tx.Exec("DELETE FROM sms WHERE id = ?", 25)
	//if err != nil {
	//	fmt.Printf("Exec delete error: %v\n", err)
	//	return
	//}

	err = tx.Commit()
	if err != nil {
		fmt.Printf("事务提交失败: %v\n", err)
		return
	}

	fmt.Printf("Goroutine %d 已完成\n", id)
}

func executeDBOperationsWithExclusiveLock1(db *sql.DB, id int) {
	tx, err := db.Begin()
	if err != nil {
		fmt.Println("开启事务失败:", err)
		return
	}
	defer tx.Rollback()

	var score int
	err = tx.QueryRow("SELECT `score` FROM sms WHERE `id` = 28 AND `number` = 2418 FOR UPDATE").Scan(&score)
	if err != nil {
		fmt.Printf("Query error: %v\n", err)
		return
	}

	//_, err = tx.Exec("UPDATE sms SET `score` = 80 WHERE `id` = 28 AND `number` = 2418")
	//if err != nil {
	//	fmt.Printf("Exec update error: %v\n", err)
	//	return
	//}

	//_, err = tx.Exec("DELETE FROM sms WHERE `id` = 27")
	//if err != nil {
	//	fmt.Printf("Exec delete error: %v\n", err)
	//	return
	//}

	err = tx.Commit()
	if err != nil {
		fmt.Printf("事务提交失败: %v\n", err)
		return
	}

	fmt.Printf("Goroutine %d 已完成\n", id)
}

func executeDBOperationsWithExclusiveLock2(db *sql.DB, id int) {
	tx, err := db.Begin()
	if err != nil {
		fmt.Println("开启事务失败:", err)
		return
	}
	defer tx.Rollback()

	//var score int
	//err = tx.QueryRow("SELECT `score` FROM sms WHERE `id` = 28 AND `number` = 2418 FOR UPDATE").Scan(&score)
	//if err != nil {
	//	fmt.Printf("Query error: %v\n", err)
	//	return
	//}

	_, err = tx.Exec("UPDATE sms SET `score` = 70 WHERE `id` = 28 AND `number` = 2418")
	if err != nil {
		fmt.Printf("Exec update error: %v\n", err)
		return
	}

	//_, err = tx.Exec("DELETE FROM sms WHERE `id` = 27")
	//if err != nil {
	//	fmt.Printf("Exec delete error: %v\n", err)
	//	return
	//}

	err = tx.Commit()
	if err != nil {
		fmt.Printf("事务提交失败: %v\n", err)
		return
	}

	fmt.Printf("Goroutine %d 已完成\n", id)
}

func Lock1(db *sql.DB) {
	tx, err := db.Begin()
	if err != nil {
		fmt.Println("开启事务失败:", err)
		return
	}
	defer tx.Rollback()
	_, err = tx.Exec("SELECT * FROM sms WHERE id = 30 FOR UPDATE")
	if err != nil {
		fmt.Println("获取锁失败:", err)
		return
	}
	_, err = tx.Exec("UPDATE sms SET points = points + 10 WHERE id = 30")
	if err != nil {
		fmt.Println("更新失败:", err)
		return
	}
	time.Sleep(time.Second * 10)
	err = tx.Commit()
	if err != nil {
		fmt.Printf("事务提交失败:", err)
		return
	}
}

func Lock2(db *sql.DB) {
	tx, err := db.Begin()
	if err != nil {
		fmt.Println("实物开启失败:", err)
		return
	}
	defer tx.Rollback()
	_, err = tx.Exec("SELECT * FROM sms WHERE id = 30 FOR UPDATE")
	if err != nil {
		fmt.Println("获取锁失败:", err)
		return
	}
	_, err = tx.Exec("UPDATE sms SET points = points - 5 WHERE id = 30")
	if err != nil {
		fmt.Println("更新失败:", err)
		return
	}
	time.Sleep(time.Second * 1)
	err = tx.Commit()
	if err != nil {
		fmt.Println("事务提交失败:", err)
		return
	}
}

func Lock3(db *sql.DB) {
	tx, err := db.Begin()
	if err != nil {
		fmt.Println("开启事务失败:", err)
		return
	}
	defer tx.Rollback()
	_, err = tx.Exec("SELECT * FROM sms WHERE id = 25 FOR UPDATE")
	if err != nil {
		fmt.Println("获取锁失败:", err)
		return
	}
	_, err = tx.Exec("UPDATE sms SET points = points + 10 WHERE id = 25")
	if err != nil {
		fmt.Println("更新失败:", err)
		return
	}
	time.Sleep(time.Second * 10)
	err = tx.Commit()
	if err != nil {
		fmt.Printf("事务提交失败:", err)
		return
	}
}

func Lock4(db *sql.DB) {
	tx, err := db.Begin()
	if err != nil {
		fmt.Println("实物开启失败:", err)
		return
	}
	defer tx.Rollback()
	_, err = tx.Exec("SELECT * FROM sms WHERE id = 25 FOR UPDATE")
	if err != nil {
		fmt.Println("获取锁失败:", err)
		return
	}
	_, err = tx.Exec("UPDATE sms SET points = points - 5 WHERE id = 25")
	if err != nil {
		fmt.Println("更新失败:", err)
		return
	}
	time.Sleep(time.Second * 1)
	err = tx.Commit()
	if err != nil {
		fmt.Println("事务提交失败:", err)
		return
	}
}

func Lock5(db *sql.DB) {
	tx, err := db.Begin()
	if err != nil {
		fmt.Println("开启事务失败:", err)
		return
	}
	defer tx.Rollback()
	_, err = tx.Exec("SELECT * FROM sms WHERE id > 24 FOR UPDATE")
	if err != nil {
		fmt.Println("获取锁失败:", err)
		return
	}
	time.Sleep(time.Second * 10)
	err = tx.Commit()
	if err != nil {
		fmt.Printf("事务提交失败:", err)
		return
	}
}

func Lock6(db *sql.DB) {
	tx, err := db.Begin()
	if err != nil {
		fmt.Println("实物开启失败:", err)
		return
	}
	defer tx.Rollback()
	_, err = tx.Exec("SELECT * FROM sms WHERE id = 28 FOR UPDATE")
	if err != nil {
		fmt.Println("获取锁失败:", err)
		return
	}
	time.Sleep(time.Second * 1)
	err = tx.Commit()
	if err != nil {
		fmt.Println("事务提交失败:", err)
		return
	}
}

// 签到功能
func (u UserSign) DoSign(ctx context.Context, id int) (bool, string, error) {
	var offset = time.Now().Local().Day() - 1 //其中减1是为了得到一个从0开始的索引
	var keys = u.buildSignKey(id)
	signed, err := rdb.SetBit(ctx, keys, int64(offset), 1).Result()
	if err != nil {
		return false, "", err
	}
	if signed == 1 {
		return true, "签到成功", nil
	}
	return false, "签到失败", nil
}

// 判断学生是否都已经签到了
func (u UserSign) CheckSign(id int) (int64, error) {
	var offset = time.Now().Local().Day() - 1
	var keys = u.buildSignKey(id)
	return rdb.GetBit(ctx, keys, int64(offset)).Result()
}

// 获取学生签到的次数
func (u UserSign) GetSignCount(id int) (int64, error) {
	var keys = u.buildSignKey(id)
	count := redis.BitCount{Start: 0, End: 31}
	return rdb.BitCount(ctx, keys, &count).Result()
}

// 获取学生首次签到的日期
func (u UserSign) GetFirstSignDate(uid int) (string, error) {
	var keys = u.buildSignKey(uid)
	pos, err := rdb.BitPos(ctx, keys, 1).Result() //获取第一位为1的位置
	if err != nil {
		return "", err
	}
	pos = pos + 1

	var day = time.Now().Local().Day()

	var offsetDay = (day - int(pos)) * -1

	return time.Now().AddDate(0, 0, offsetDay).Format("2006-01-02"), nil
}

// 获取学生当月签到情况
func (u UserSign) GetSignInfo(uid int) (interface{}, error) {
	var keys = u.buildSignKey(uid)
	var day = time.Now().Local().Day()
	var dddd = fmt.Sprintf("u%d", day)
	st, _ := rdb.Do(ctx, keys, "GET", dddd, 0).Result()
	f := st.([]interface{})
	var res = make([]bool, 0)
	var days = make([]string, 0)
	var v = f[0].(int64)
	fmt.Println(v)
	for i := day; i > 0; i-- {
		var pos = (day - i) * -1
		var keys = time.Now().Local().AddDate(0, 0, pos).Format("2006-01-02")
		days = append(days, keys)
		var value = v>>1<<1 != v
		res = append(res, value)
		v >>= 1
	}
	fmt.Println(res)
	fmt.Println(days)
	return nil, nil
}

// 构建一个用于签到的键值
func (u UserSign) buildSignKey(id int) string {
	var nowDate = u.formatDate()
	return fmt.Sprintf("stu:sign:%d:%s", id, nowDate)
}

// 获取当前的日期
func (u UserSign) formatDate() string {
	return time.Now().Format("2006-01")
}

// 点赞功能
func getKey(tid int64) string {
	return fmt.Sprintf("teacher:like:%d", tid)
}

// tid 需要点赞教师的ID   id 学生ID
func GiveLike(ctx context.Context, tid int64, id int64) (bool, error) {
	keys := getKey(tid)
	res, err := rdb.GetBit(ctx, keys, id-1).Result()
	if err != nil {
		return false, err
	}

	if res == 1 {
		return false, err
	}

	_, err = rdb.SetBit(ctx, keys, id-1, 1).Result()
	if err != nil {
		return false, err
	}

	return true, nil
}

// 查询是否已经点赞了
func GiveLikeSelect(tid int64, id int64) (bool, error) {
	var keys = getKey(tid)
	res, err := rdb.GetBit(context.Background(), keys, id-1).Result()
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
	results, err := rdb.ZRevRangeWithScores(ctx, "students", 0, -1).Result()
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
