package module

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestUpdateRowHandler(t *testing.T) {
	// 创建一个表单数据
	formData := url.Values{}
	formData.Set("name", "Test Student")
	formData.Set("score", "90")

	// 创建一个模拟的请求
	req, err := http.NewRequest("POST", "/update", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// 创建一个 ResponseRecorder (模拟的 ResponseWriter) 来记录响应
	rr := httptest.NewRecorder()

	// 实例化一个具体的 http.Handler，并调用它的 ServeHTTP 方法来直接使用我们的模拟请求和 ResponseRecorder
	handler := http.HandlerFunc(UpdateRowHandler)
	handler.ServeHTTP(rr, req)

	// 检查响应状态码
	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}
}
