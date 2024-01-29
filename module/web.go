package module

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

func RegisterStudentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "表单解析错误", http.StatusBadRequest)
			return
		}
		idStr := r.FormValue("number")
		student_id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "无效的学生ID", http.StatusBadRequest)
			return
		}
		pwd := r.FormValue("password")
		err = register(student_id, pwd)
		if err != nil {
			log.Printf("注册失败，err：%v\n", err)
			http.Error(w, "注册失败", http.StatusInternalServerError)
			return
		}
		// 显示注册成功消息
		fmt.Fprint(w, "注册成功，请返回登录页面。")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		http.ServeFile(w, r, "./module/templates/studentRegister.html")
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "表单解析错误", http.StatusBadRequest)
			return
		}
		userName := r.FormValue("username")
		pwd := r.FormValue("password")
		isValid, err := validate(userName, pwd)
		if err != nil {
			log.Printf("登录验证过程中出错：%v", err)
			http.Error(w, "内部服务器错误", http.StatusInternalServerError)
			return
		}
		if isValid {
			cookie := &http.Cookie{
				Name:     "username",
				Value:    userName,
				MaxAge:   0,
				HttpOnly: false,
			}
			http.SetCookie(w, cookie)
			http.Redirect(w, r, "/home", http.StatusSeeOther)
		} else {
			fmt.Fprint(w, "登录失败，请重新登录。")
		}
	} else {
		http.ServeFile(w, r, "./module/templates/select.html")
	}
}

//func LoginHandler(w http.ResponseWriter, r *http.Request) {
//	if r.Method == http.MethodPost {
//		err := r.ParseForm()
//		if err != nil {
//			log.Printf("LoginHandler: 表单解析错误: %v", err)
//			http.Error(w, "表单解析错误", http.StatusBadRequest)
//			return
//		}
//		userName := r.FormValue("username")
//		pwd := r.FormValue("password")
//		studentID, err := strconv.Atoi(userName)
//		if err != nil {
//			log.Printf("LoginHandler: 用户名转换为学生ID时出错: %v", err)
//			http.Error(w, "无效的用户名", http.StatusBadRequest)
//			return
//		}
//		_, err = StudentLogin(studentID, pwd)
//		if err != nil {
//			if errors.Is(err, sql.ErrNoRows) || err.Error() == "密码不匹配" {
//				log.Printf("LoginHandler: 登录失败 - 用户名或密码错误: %v", err)
//				http.Error(w, "用户名或密码错误", http.StatusUnauthorized)
//				return
//			}
//			log.Printf("LoginHandler: 登录时出错: %v", err)
//			http.Error(w, "内部服务器错误", http.StatusInternalServerError)
//			return
//		}
//
//		log.Printf("LoginHandler: 用户 %s 登录成功", userName)
//		cookie := &http.Cookie{
//			Name:     "username",
//			Value:    userName,
//			MaxAge:   0,
//			HttpOnly: false,
//		}
//		http.SetCookie(w, cookie)
//		http.Redirect(w, r, "/home", http.StatusSeeOther)
//		return
//	} else {
//		log.Println("LoginHandler: 处理GET请求")
//		http.ServeFile(w, r, "./module/templates/select.html")
//		return
//	}
//}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// 确保只有根路径 "/" 被这个处理器处理
	if r.URL.Path != "/home" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "./module/templates/choose.html")
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
	tmpl, err := template.ParseFiles("module/templates/querySuccess.html")
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
		//渲染成功信息的模板
		tmpl, err := template.ParseFiles("module/templates/addSuccess.html") // 修改为实际的模板文件路径
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
		name := r.FormValue("name")
		score, err := strconv.Atoi(r.FormValue("score"))
		if err != nil {
			log.Printf("UpdateRowHandler: 无效的分数值,err:%v\n", err)
			http.Error(w, "无效的分数值", http.StatusBadRequest)
			return
		}
		err = updateRow(name, score)
		if err != nil {
			log.Printf("UpdateRowHandler: 更新学生信息失败: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} /*else {
		GET请求（读）
		id, err := strconv.Atoi(r.FormValue("id"))
		if err != nil {
			http.Error(w, "无效的学生ID值", http.StatusBadRequest)
			return
		}
		err = queryRow(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = tpl.ExecuteTemplate(w, "update.html", nil)
		if err != nil {
			return
		}
	}*/
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return
}

// 删除学生信息的Handler
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
		// 调用deleteRow来删除学生
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
