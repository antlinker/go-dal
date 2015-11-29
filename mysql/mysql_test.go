package mysql

import (
	"testing"
	"time"

	"gopkg.in/dal.v1"
)

type Student struct {
	ID       int64
	StuCode  string
	StuName  string
	Sex      int
	Age      int
	Birthday time.Time
	Memo     string
}

func getDB() *MysqlProvider {
	provider := new(MysqlProvider)
	err := provider.InitDB(`{"datasource":"root:123456@tcp(127.0.0.1:3306)/testdb?charset=utf8"}`)
	if err != nil {
		panic(err)
	}
	return provider
}

func TestInsert(t *testing.T) {
	db := getDB()
	birthday, _ := time.Parse("2006-01-02", "1990-10-13")
	student := Student{
		StuCode:  "S002",
		StuName:  "Lyric",
		Sex:      1,
		Age:      26,
		Birthday: birthday,
	}
	result := db.Exec(dal.NewTranAEntity("student", student).Entity)
	if result.Error != nil {
		t.Error(result.Error)
		return
	}
	t.Log("Insert:", result.Result)
}

func TestUpdate(t *testing.T) {
	db := getDB()
	student := map[string]interface{}{
		"StuName": "Lyric1",
		"Sex":     1,
		"Memo":    "Message",
	}
	entityResult := dal.NewTranUEntity("student",
		student,
		dal.NewFieldsKvCondition(map[string]interface{}{"StuCode": "S002"}).Condition)
	if err := entityResult.Error; err != nil {
		t.Error(err)
		return
	}
	result := db.Exec(entityResult.Entity)
	if result.Error != nil {
		t.Error(result.Error)
		return
	}
	t.Log("Update:", result.Result)
}

func TestQueryList(t *testing.T) {
	db := getDB()
	entity := dal.NewQueryEntity("student", dal.NewCondition("").Condition)()
	var studs []Student
	err := db.AssignList(entity.Entity, &studs)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("Student list:", studs)
}

func TestDelete(t *testing.T) {
	db := getDB()
	cond := dal.NewFieldsKvCondition(map[string]interface{}{"StuCode": "S002"})
	entity := dal.NewTranDEntity("student", cond.Condition)
	result := db.Exec(entity.Entity)
	if result.Error != nil {
		t.Error(result.Error)
		return
	}
	t.Log("Delete:", result.Result)
}
