package module

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/streadway/amqp"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// 全局变量来存储最新的消息
var lastMessage string
var lastMessageMutex sync.Mutex // 用于同步访问lastMessage
var messageChannel = make(chan string)

func init() {
	go func() {
		for msg := range messageChannel {
			lastMessageMutex.Lock()
			lastMessage = msg
			lastMessageMutex.Unlock()
		}
	}()
}

func ConcurrencyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	//var wg sync.WaitGroup
	//concurrencyLevel := 100
	//wg.Add(concurrencyLevel)
	//for i := 0; i < concurrencyLevel; i++ {
	//	go func(i int) {
	//		defer wg.Done()
	//		//executeDBOperationsWithSharedLock(db, i)
	//		executeDBOperationsWithExclusiveLock(db, i)
	//	}(i)
	//}
	//
	//wg.Wait()

	//go task1()
	//go task2()
	//time.Sleep(time.Second * 2)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		Lock5(db)
	}()

	go func() {
		defer wg.Done()
		Lock6(db)
	}()
	wg.Wait()
	w.Write([]byte("并发操作已完成"))
}

func task1() {
	for i := 0; i < 5; i++ {
		executeDBOperationsWithExclusiveLock1(db, i)
		fmt.Println("Task 1 -", i)
		time.Sleep(time.Millisecond * 500)
	}
}

func task2() {
	for i := 0; i < 5; i++ {
		executeDBOperationsWithExclusiveLock2(db, i)
		fmt.Println("Task 2 -", i)
		time.Sleep(time.Millisecond * 500)
	}
}

func ShowStudentHandler(w http.ResponseWriter, r *http.Request) {
	names, err := obtainStudent(context.Background(), db)
	if err != nil {
		http.Error(w, "获取学生姓名失败", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(names)
	if err != nil {
		http.Error(w, "Failed to encode student names to JSON", http.StatusInternalServerError)
	}
}

func MessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Printf("表单解析错误: %v\n", err)
			http.Error(w, "表单解析错误", http.StatusBadRequest)
			return
		}
		message := r.FormValue("message")
		fmt.Printf("收到消息: %s\n", message)
		messageChannel <- message
		fmt.Fprintf(w, "消息已发送")
	} else {
		http.ServeFile(w, r, "module/templates/sendMessage.html")
	}
}

func MqHandler(conn *amqp.Connection, w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "解析表单失败", http.StatusInternalServerError)
			return
		}
		message := r.FormValue("message")
		if _, err := emit(conn, message); err != nil {
			http.Error(w, "消息发送失败", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "消息发送成功")
	} else {
		http.ServeFile(w, r, "module/templates/sendMessage.html")
	}
}

//func ShowMessageHandler(w http.ResponseWriter, r *http.Request) {
//	lastMessageMutex.Lock()
//	msg := lastMessage
//	lastMessageMutex.Unlock()
//
//	tmpl, err := template.ParseFiles("module/templates/studentPage.html")
//	if err != nil {
//		http.Error(w, "模板解析错误: "+err.Error(), http.StatusInternalServerError)
//		return
//	}
//	data := struct {
//		Message string
//	}{
//		Message: msg,
//	}
//	if err := tmpl.Execute(w, data); err != nil {
//		http.Error(w, "模板执行错误: "+err.Error(), http.StatusInternalServerError)
//	}
//}

func StudentSelectHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, rdb *redis.Client) {
	if r.Method == http.MethodGet {
		renderTemplate(w, nil)
	} else if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "解析表单错误", http.StatusBadRequest)
			return
		}

		studentNumberStr, err := rdb.Get(ctx, "student_id").Result()
		if err != nil {
			log.Printf("获取学号失败，err：%v", err)
			http.Error(w, "内部服务器错误", http.StatusInternalServerError)
			return
		}

		studentNumber, err := strconv.Atoi(studentNumberStr)
		if err != nil {
			fmt.Printf("转换失败，err：%v", err)
			http.Error(w, "内部服务器错误", http.StatusInternalServerError)
			return
		}

		courseOption := r.FormValue("student-options")
		_, err = db.Exec("UPDATE sms SET course = ? WHERE number = ?", courseOption, studentNumber)
		if err != nil {
			http.Error(w, "课程添加失败", http.StatusInternalServerError)
			return
		}

		var successMessage string
		switch courseOption {
		case "数学课":
			successMessage = "你已成功选取数学课程"
		case "语文课":
			successMessage = "你已成功选取语文课程"
		case "英语课":
			successMessage = "你已成功选取英语课程"
		case "政治课":
			successMessage = "你已成功选取政治课程"
		case "地理课":
			successMessage = "你已成功选取地理课程"
		case "化学课":
			successMessage = "你已成功选取化学课程"
		}

		renderTemplate(w, map[string]string{
			"SuccessMessage": successMessage,
		})
	} else {
		http.Error(w, "err", http.StatusMethodNotAllowed)
	}
}

func renderTemplate(w http.ResponseWriter, data interface{}) {
	tmpl, err := template.ParseFiles("./module/templates/studentSelect.html")
	if err != nil {
		http.Error(w, "Error loading page", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		return
	}
}

func StudentPageHandler(conn *amqp.Connection, chMsg chan string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "只允许GET方法", http.StatusMethodNotAllowed)
			return
		}
		// 尝试从Cookie获取学生ID
		cookie, err := r.Cookie("student_id")
		if err != nil {
			http.Error(w, "未授权访问", http.StatusUnauthorized)
			return
		}

		studentID, err := strconv.ParseInt(cookie.Value, 10, 64)
		if err != nil {
			http.Error(w, "无效的学生ID", http.StatusBadRequest)
			return
		}

		favoriteTeacher := r.URL.Query().Get("favoriteTeacher")
		if favoriteTeacher != "" {
			//var tid int64
			switch favoriteTeacher {
			case "math":
				tid = 1
			case "chinese":
				tid = 2
			case "english":
				tid = 3
			}

			_, err = GiveLike(ctx, tid, studentID)
			if err != nil {
				http.Error(w, fmt.Sprintf("保存投票结果失败：%v", err), http.StatusInternalServerError)
				return
			}

			alreadyLiked, err := GiveLikeSelect(tid, studentID)
			if err != nil {
				http.Error(w, fmt.Sprintf("查询点赞状态失败：%v", err), http.StatusInternalServerError)
				return
			}
			if alreadyLiked {
				fmt.Fprintf(w, "你已经给这位老师投过票了")
				return
			}
		}

		cookie, err = r.Cookie("student_id")
		if err != nil {
			http.Error(w, "未授权访问", http.StatusUnauthorized)
			return
		}
		number, err := strconv.Atoi(cookie.Value)
		if err != nil {
			http.Error(w, "无效的学生ID", http.StatusBadRequest)
			return
		}
		stu, err := queryRow(number)
		if err != nil {
			log.Printf("查询失败，err：%v\n", err)
			http.Error(w, "查询失败", http.StatusInternalServerError)
			return
		}

		lastMessageMutex.Lock()
		msg := lastMessage
		lastMessageMutex.Unlock()

		signCount, err := getSignCount(rdb, strconv.Itoa(stu.Number), "2024-03")
		if err != nil {
			fmt.Println("登录次数计算失败:", err)
		} else {
			fmt.Println("登录次数为:", signCount)
		}
		data := struct {
			Name      string
			Number    int
			Score     int
			Message   string
			SignCount int
		}{
			Name:      stu.Name,
			Number:    stu.Number,
			Score:     stu.Score,
			Message:   msg,
			SignCount: signCount,
		}
		tmpl, err := template.ParseFiles("module/templates/studentPage.html")
		if err != nil {
			log.Printf("模板解析错误：%v\n", err)
			http.Error(w, "内部服务器错误", http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("模板渲染错误，err：%v\n", err)
			//http.Error(w, "模板执行错误", http.StatusInternalServerError)
		}
		// 启动一个新的goroutine来运行Worker函数？？？
		go Receive(conn, chMsg)
	}
}

func getSignCount(rdb *redis.Client, studentID string, yearMonth string) (int, error) {
	key := fmt.Sprintf("stu:sign:%s:%s", studentID, yearMonth)
	ctx := context.Background()
	result, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	} else if err != nil {
		return 0, err
	}
	signCount, err := strconv.Atoi(result)
	if err != nil {
		return 0, err
	}

	return signCount, err
}

func SignInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只允许POST方法", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("student_id")
	if err != nil {
		http.Error(w, "未授权访问", http.StatusUnauthorized)
		return
	}
	studentID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		http.Error(w, "无效的学生ID", http.StatusBadRequest)
		return
	}

	var u UserSign
	signed, message, err := u.DoSign(ctx, studentID)
	if err != nil {
		http.Error(w, fmt.Sprintf("学生签到失败：%v", err), http.StatusInternalServerError)
		return
	}
	if signed {
		w.Write([]byte(message))
		return
	}
	count, err := u.GetSignCount(studentID)
	data := struct {
		SignCount int64
	}{
		SignCount: count,
	}
	tmpl, err := template.ParseFiles("module/templates/studentPage.html")
	if err != nil {
		log.Printf("模板解析错误：%v\n", err)
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("模板渲染错误，err：%v\n", err)
		http.Error(w, "模板执行错误", http.StatusInternalServerError)
	}
}

func RegisterStudentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		err := r.ParseForm()
		if err != nil {
			log.Printf("表单解析错误，err:%v\n", err)
			return
		}
		student_id := r.FormValue("number")
		pwd := r.FormValue("password")
		err = register(student_id, pwd)
		if err != nil {
			http.Error(w, "注册失败，请重新输入正确的学号或密码", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/register", http.StatusSeeOther)
	} else {
		http.ServeFile(w, r, "./module/templates/studentRegister.html")
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Printf("表单解析错误，err：%v\n", err)
			return
		}
		userName := r.FormValue("username")
		pwd := r.FormValue("password")
		action := r.FormValue("action")
		var userType string
		switch action {
		case "学生登录":
			userType = "student"
		case "管理员登录":
			userType = "teacher"
		default:
			fmt.Fprint(w, "未知登录类型")
			return
		}
		isValid, err := validate(userName, pwd, userType)
		if err != nil {
			log.Printf("登录验证过程中出错：%v", err)
			http.Error(w, "内部服务器错误", http.StatusInternalServerError)
			return
		}
		if isValid {
			sessionKey := fmt.Sprintf("%s_id", userType)
			err := rdb.Set(ctx, sessionKey, userName, 24*time.Hour).Err()
			if err != nil {
				log.Printf("无法将用户信息存储到Redis：%v", err)
				http.Error(w, "内部服务器错误", http.StatusInternalServerError)
				return
			}

			switch action {
			case "学生登录":
				http.SetCookie(w, &http.Cookie{
					Name:  "student_id",
					Value: userName,
					Path:  "/",
				})
				http.Redirect(w, r, "/studentPage", http.StatusSeeOther)
			case "管理员登录":
				http.SetCookie(w, &http.Cookie{
					Name:  "teacher_id",
					Value: userName,
					Path:  "/",
				})
				http.Redirect(w, r, "/home", http.StatusSeeOther)
			}
		} else {
			fmt.Fprint(w, "用户名或者密码错误，请重新登录。")
		}
	} else {
		http.ServeFile(w, r, "./module/templates/login.html")
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// 确保只有根路径 "/" 被这个处理器处理
	if r.URL.Path != "/home" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "./module/templates/home.html")
}

func QueryRowHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "./module/templates/query.html")
		return
	} else if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "解析表格失败", http.StatusBadRequest)
			return
		}
		idStr := r.FormValue("id")
		number, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "无效的学生ID", http.StatusBadRequest)
			return
		}
		stu, err := queryRow(number)
		if err != nil {
			log.Printf("查询失败，err：%v\n", err)
			http.Error(w, "查询失败", http.StatusInternalServerError)
			return
		}
		// 在模板中使用查询结果
		tmpl, err := template.ParseFiles("./module/templates/querySuccess.html")
		if err != nil {
			log.Printf("模板解析错误：%v\n", err)
			http.Error(w, fmt.Sprintf("模板解析错误: %v", err), http.StatusInternalServerError)
			return
		}
		data := struct {
			Number int
			Name   string
			Score  int
		}{
			Number: stu.Number,
			Name:   stu.Name,
			Score:  stu.Score,
		}
		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, "模板渲染错误", http.StatusInternalServerError)
			return
		}
	}
}

func QueryAllRowHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		ctx := r.Context()
		//students, err := queryMultiRow()
		students, err := studentsScore(ctx, db, rdb)
		if err != nil {
			http.Error(w, "内部服务器错误", http.StatusInternalServerError)
			return
		}
		tmpl, err := template.ParseFiles("module/templates/queryAll.html")
		if err != nil {
			log.Printf("模板解析错误: %v\n", err)
			http.Error(w, "内部服务器错误", http.StatusInternalServerError)
			return
		}
		err = tmpl.Execute(w, students)
		if err != nil {
			log.Printf("模板执行错误: %v\n", err)
			http.Error(w, "模板渲染错误", http.StatusInternalServerError)
		}
	}
}

// 添加学生信息的Handler
func InsertRowHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Printf("InsertRowHandler：表单解析出错: %v\n", err)
			http.Error(w, "解析表单失败", http.StatusBadRequest)
			return
		}
		numberStr := r.FormValue("student_id")
		number, err := strconv.Atoi(numberStr)
		if err != nil {
			http.Error(w, "学号错误", http.StatusBadRequest)
		}
		name := r.FormValue("name")
		scoreStr := r.FormValue("score")
		score, err := strconv.Atoi(scoreStr)
		if err != nil {
			log.Printf("InsertRowHandler：分数格式错误:%v\n", err)
			http.Error(w, "分数格式错误", http.StatusBadRequest)
			return
		}
		err = insertRow(number, name, score)
		if err != nil {
			log.Printf("InsertRowHandler: 添加学生失败: %v", err)
			http.Error(w, "学生添加失败", http.StatusInternalServerError)
			return
		}
		log.Printf("InsertRowHandler: 学生添加成功，学号: %d", number)
		tmpl, err := template.ParseFiles("./module/templates/addSuccess.html")
		if err != nil {
			log.Printf("InsertRowHandler：模板解析错误:%v\n", err)
			http.Error(w, "模板解析错误", http.StatusInternalServerError)
			return
		}
		//使用模板渲染成功信息，并传递学生ID
		data := struct {
			InsertedID int
		}{
			InsertedID: number,
		}
		err = tmpl.Execute(w, data)
		if err != nil {
			log.Printf("InsertRowHandler：模板渲染错误：%v\n", err)
			http.Error(w, "模板渲染错误", http.StatusInternalServerError)
			return
		}
		log.Println("InsertRowHandler：学生添加成功！")
	} else {
		log.Println("InsertRowHandler: 显示添加学生页面")
		http.ServeFile(w, r, "module/templates/add.html")
	}
}

// 修改学生信息的Handler
func UpdateRowHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		//POST请求（写）
		err := r.ParseForm()
		if err != nil {
			log.Printf("UpdateRowHandler: 解析表格失败,err:%v\n", err)
			http.Error(w, "解析表格失败", http.StatusBadRequest)
			return
		}
		number, err := strconv.Atoi(r.FormValue("number"))
		if err != nil {
			http.Error(w, "无效的学号", http.StatusBadRequest)
			return
		}
		score, err := strconv.Atoi(r.FormValue("score"))
		if err != nil {
			http.Error(w, "无效的分数值", http.StatusBadRequest)
			return
		}
		err = updateRow(number, score)
		if err != nil {
			log.Printf("UpdateRowHandler: 更新学生信息失败: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl, err := template.ParseFiles("./module/templates/updateSuccess.html")
		data := struct {
			InsertedID int
			Score      int
		}{
			InsertedID: number,
			Score:      score,
		}
		err = tmpl.Execute(w, data)
		if err != nil {
			log.Printf("InsertRowHandler：模板渲染错误：%v\n", err)
			http.Error(w, "模板渲染错误", http.StatusInternalServerError)
			return
		}
		log.Println("UpdateRowHandler：学生成绩修改成功！")
	} else {
		http.ServeFile(w, r, "./module/templates/update.html")
	}
}

// 删除学生信息
func DeleteRowHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Printf("表单解析错误: %v\n", err)
			http.Error(w, "表单解析错误", http.StatusBadRequest)
			return
		}
		numberStr := r.FormValue("student_id")
		number, err := strconv.Atoi(numberStr)
		if err != nil {
			log.Printf("无效的学生ID: %v\n", err)
			http.Error(w, "无效的学生ID", http.StatusBadRequest)
			return
		}
		if err := deleteRow(number); err != nil {
			log.Printf("DeleteRowHandler: 学生删除失败: %v", err)
			// 如果删除过程中出现错误，返回内部服务器错误
			http.Error(w, fmt.Sprintf("删除失败: %v", err), http.StatusInternalServerError)
			return
		}
		tmpl, err := template.ParseFiles("./module/templates/deleteSuccess.html")
		if err != nil {
			log.Printf("模板解析错误:%v\n", err)
			http.Error(w, "模板解析错误", http.StatusInternalServerError)
			return
		}
		data := struct {
			DeleteID int
		}{
			DeleteID: number,
		}
		err = tmpl.Execute(w, data)
		if err != nil {
			log.Printf("模板渲染错误：%v\n", err)
			http.Error(w, "模板渲染错误", http.StatusInternalServerError)
			return
		}
		log.Println("学号为", number, "的学生删除成功！")
	} else {
		http.ServeFile(w, r, "./module/templates/delete.html")
	}
}

func validate(username, password, userType string) (bool, error) {
	var dbPassword string
	var err error
	if userType == "student" {
		err = db.QueryRow("SELECT password FROM stu WHERE student_id = ?", username).Scan(&dbPassword)
	} else if userType == "teacher" {
		err = db.QueryRow("SELECT password FROM teachers WHERE tname = ?", username).Scan(&dbPassword)
	} else {
		return false, fmt.Errorf("未知的用户类型: %s", userType)
	}
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
