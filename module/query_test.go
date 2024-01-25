package module

import (
	"github.com/DATA-DOG/go-sqlmock"
	"testing"
)

func TestQueryRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建mock数据库连接时出错: %s", err)
	}
	defer db.Close()

	// 设置预期的查询
	rows := sqlmock.NewRows([]string{"id", "name", "score"}).
		AddRow(1, "Alice", 90)
	mock.ExpectQuery("^select \\* from `sms` where id=\\?").
		WithArgs(1).
		WillReturnRows(rows)

	// 调用queryRow函数
	stu, err := queryRow(1)
	if err != nil {
		t.Errorf("期望无错误，但得到: %s", err)
	}
	if stu.Number != 1 || stu.Name != "Alice" || stu.Score != 90 {
		t.Errorf("查询结果不符合预期: %+v", stu)
	}

	// 确保所有预期动作都被满足
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("存在未满足的预期: %s", err)
	}
}
