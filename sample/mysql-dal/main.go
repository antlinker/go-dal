package main

import (
	"fmt"
	"time"

	"gopkg.in/dal.v1"
	_ "gopkg.in/dal.v1/mysql"
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

func main() {
	dal.RegisterProvider(dal.MYSQL, `{"datasource":"root:123456@tcp(192.168.33.70:3306)/testdb?charset=utf8","maxopen":100,"maxidle":50}`)
	insert()
	list()
	update()
	list()
	delete()
	list()
}

func insert() {
	stud := Student{
		StuCode:  "S001",
		StuName:  "Lyric",
		Sex:      1,
		Age:      25,
		Birthday: time.Now(),
		Memo:     "Message...",
	}
	entity := dal.NewTranAEntity("student", stud).Entity
	result := dal.GDAL.Exec(entity)
	if err := result.Error; err != nil {
		panic(err)
	}
	fmt.Println("===> Student Insert:", result.Result)
}

func update() {
	stud := map[string]interface{}{
		"StuName": "Lyric01",
		"Sex":     1,
		"Age":     26,
	}
	cond := dal.NewFieldsKvCondition(map[string]interface{}{"StuCode": "S001"}).Condition
	entity := dal.NewTranUEntity("student", stud, cond).Entity
	result := dal.GDAL.Exec(entity)
	if err := result.Error; err != nil {
		panic(err)
	}
	fmt.Println("===> Student Update:", result.Result)
}

func delete() {
	cond := dal.NewFieldsKvCondition(Student{StuCode: "S001"}).Condition
	entity := dal.NewTranDEntity("student", cond).Entity
	result := dal.GDAL.Exec(entity)
	if err := result.Error; err != nil {
		panic(err)
	}
	fmt.Println("===> Student Delete:", result.Result)
}

func list() {
	entity := dal.NewQueryEntity("student", dal.QueryCondition{}, "*")(dal.List).Entity
	var stuData []Student
	err := dal.GDAL.AssignList(entity, &stuData)
	if err != nil {
		panic(err)
	}
	fmt.Println("===> Student List:", stuData)
}
