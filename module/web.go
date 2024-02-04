package module

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

func StudentPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		cookie, err := r.Cookie("student_id")
		if err != nil {
			// 处理cookie不存在的情况
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
		tmpl, err := template.ParseFiles("./module/templates/studentPage.html")
		if err != nil {
			log.Printf("模板解析错误：%v\n", err)
			http.Error(w, "内部服务器错误", http.StatusInternalServerError)
			return
		}
		data := struct {
			Name   string
			Number int
			Score  int
		}{
			Name:   stu.Name,
			Number: stu.Number,
			Score:  stu.Score,
		}
		err = tmpl.Execute(w, data)
		if err != nil {
			log.Printf("模板渲染错误，err：%v\n", err)
			return
		}
	}
}

func RegisterStudentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
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
		fmt.Fprint(w, "注册成功，请返回登录页面。")
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
		isValid, err := validate(userName, pwd)
		if err != nil {
			log.Printf("登录验证过程中出错：%v", err)
			http.Error(w, "内部服务器错误", http.StatusInternalServerError)
			return
		}
		if isValid {
			switch action {
			case "学生登录":
				// 设置cookie
				http.SetCookie(w, &http.Cookie{
					Name:  "student_id",
					Value: userName,
					Path:  "/",
				})
				http.Redirect(w, r, "/studentPage", http.StatusSeeOther)
			case "管理员登录":
				http.Redirect(w, r, "/home", http.StatusSeeOther)
			default:
				fmt.Fprint(w, "未知登录类型")
			}
		} else {
			fmt.Fprint(w, "学号或者密码错误，请重新登录。")
		}

	} else {
		http.ServeFile(w, r, "./module/templates/select.html")
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
	if r.Method != "POST" {
		http.Error(w, "仅支持POST请求", http.StatusMethodNotAllowed)
		return
	}
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
	//else {
	//	http.ServeFile(w, r, "./module/templates/query.html")
	//}
}

func QueryAllRowHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		students, err := queryMultiRow()
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

//http.Redirect(w, r, "/", http.StatusSeeOther)
//return

// 添加学生信息的Handler
func InsertRowHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
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
	if r.Method == "POST" {
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
	if r.Method == "POST" {
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
