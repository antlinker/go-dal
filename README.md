# GO-DAL

> 提供基于Golang的数据访问层

## 获取

``` bash
$ go get gopkg.in/dal.v1
```

## 针对MySQL数据库的CRUD范例

``` go
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
	dal.RegisterProvider(dal.MYSQL, `{"datasource":"root:123456@tcp(127.0.0.1:3306)/testdb?charset=utf8","maxopen":100,"maxidle":50}`)
	insert()
	update()
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

```

## License

	Copyright 2015.All rights reserved.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.