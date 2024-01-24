package module

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "表单解析错误", http.StatusBadRequest)
			return
		}
		userName := r.FormValue("username")
		pwd := r.FormValue("password")

		// 如果输入正确，则发送cookie 并给出反馈
		if userName == "user" && pwd == "123" {
			cookie := &http.Cookie{
				Name:     "username",
				Value:    userName,
				MaxAge:   0, // cookie 的最大存活时间
				HttpOnly: false,
			}
			http.SetCookie(w, cookie)
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		} else {
			fmt.Fprint(w, "登陆失败，请重新登陆")
		}
	} else {
		// 如果不是post请求，则返回登陆页面
		http.ServeFile(w, r, "./module/templates/login.html")
		return
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// 确保只有根路径 "/" 被这个处理器处理
	if r.URL.Path != "/home" {
		http.NotFound(w, r)
		return
	}
	// 加载并发送 choose.html 文件
	http.ServeFile(w, r, "./module/templates/choose.html")
}

// 获取所有学生信息的Handler
func QueryRowHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "解析表格失败", http.StatusBadRequest)
			return
		}
		idStr := r.FormValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "无效的学生ID", http.StatusBadRequest)
			return
		}
		// 查询学生信息
		stu, err := queryRow(id)
		if err != nil {
			http.Error(w, "查询失败", http.StatusInternalServerError)
			return
		}
		//在模板中使用查询结果
		err = tpl.ExecuteTemplate(w, "query.html", stu)
		if err != nil {
			return
		}
	} else {
		http.Error(w, "无效请求", http.StatusMethodNotAllowed)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return
}

// 添加学生信息的Handler
func InsertRowHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
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
			http.Error(w, "分数格式错误", http.StatusBadRequest)
			return
		}
		err = insertRow(number, name, score)
		if err != nil {
			http.Error(w, "添加学生失败", http.StatusInternalServerError)
			return
		}

		//渲染成功信息的模板
		tmpl, err := template.ParseFiles("module/templates/success.html") // 修改为实际的模板文件路径
		if err != nil {
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
			http.Error(w, "模板渲染错误", http.StatusInternalServerError)
		}
	} else {
		http.ServeFile(w, r, "module/templates/add.html") // 修改为实际的添加学生表单文件路径
	}
}

// 修改学生信息的Handler
func UpdateRowHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		//POST请求（写）
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "解析表格失败", http.StatusBadRequest)
			return
		}
		name := r.FormValue("name")
		score, err := strconv.Atoi(r.FormValue("score"))
		if err != nil {
			http.Error(w, "无效的分数值", http.StatusBadRequest)
			return
		}
		err = updateRow(name, score)
		if err != nil {
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
		err = tpl.ExecuteTemplate(w, "edit.html", nil)
		if err != nil {
			return
		}
	}*/
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return
}

// 删除学生信息的Handler
func DeleteRowHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "表单解析错误", http.StatusBadRequest)
		return
	}
	numberStr := r.FormValue("student_id")
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		http.Error(w, "无效的学生ID", http.StatusBadRequest)
		return
	}
	// 调用deleteRow来删除学生
	if err := deleteRow(number); err != nil {
		// 如果删除过程中出现错误，返回内部服务器错误
		http.Error(w, fmt.Sprintf("删除失败: %v", err), http.StatusInternalServerError)
		return
	}
	// 删除成功后重定向到根路径
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return
}
