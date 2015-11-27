package mysql

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/LyricTian/go-dal"
	_ "github.com/go-sql-driver/mysql"
)

var (
	GDB *sql.DB
)

type Config struct {
	DataSource   string `json:"datasource"`
	MaxOpenConns int    `json:"maxopen"`
	MaxIdleConns int    `json:"maxidle"`
}

type MysqlProvider struct{}

func (mp *MysqlProvider) Error(funcName, errInfo string) error {
	return fmt.Errorf("[go-dal:mysql:mysql.go]%s:%s", funcName, errInfo)
}

func (mp *MysqlProvider) InitDB(config string) error {
	funcName := "InitDB"
	var cfg Config
	if err := json.NewDecoder(bytes.NewBufferString(config)).Decode(&cfg); err != nil {
		return mp.Error(funcName, err.Error())
	}
	if cfg.DataSource == "" {
		return mp.Error(funcName, "`datasource` can't be empty!")
	}
	db, err := sql.Open("mysql", cfg.DataSource)
	if err != nil {
		return mp.Error(funcName, err.Error())
	}
	if v := cfg.MaxOpenConns; v > 0 {
		db.SetMaxOpenConns(v)
	}
	if v := cfg.MaxIdleConns; v > 0 {
		db.SetMaxIdleConns(v)
	}
	GDB = db
	return nil
}

func (mp *MysqlProvider) Single(entity dal.QueryEntity) (map[string]string, error) {
	funcName := "Single"
	if entity.ResultType != dal.Single {
		entity.ResultType = dal.Single
	}
	sqlText, values := mp.parseQuerySQL(entity)
	data, err := mp.queryData(sqlText[0], values...)
	if err != nil {
		return nil, mp.Error(funcName, err.Error())
	}
	if len(data) == 0 {
		return make(map[string]string), nil
	}
	return data[0], nil
}

func (mp *MysqlProvider) List(entity dal.QueryEntity) ([]map[string]string, error) {
	funcName := "List"
	if entity.ResultType != dal.List {
		entity.ResultType = dal.List
	}
	sqlText, values := mp.parseQuerySQL(entity)
	data, err := mp.queryData(sqlText[0], values...)
	if err != nil {
		return nil, mp.Error(funcName, err.Error())
	}

	return data, nil
}

func (mp *MysqlProvider) Pager(entity dal.QueryEntity) (dal.QueryPagerResult, error) {
	var qResult dal.QueryPagerResult
	funcName := "Pager"
	if entity.ResultType != dal.Pager {
		entity.ResultType = dal.Pager
	}
	sqlText, values := mp.parseQuerySQL(entity)
	var (
		errs []error
		mux  = new(sync.RWMutex)
		wg   = new(sync.WaitGroup)
	)
	wg.Add(2)

	go func(result *dal.QueryPagerResult, errs *[]error) {
		defer wg.Done()
		data, err := mp.queryData(sqlText[0], values...)
		mux.Lock()
		defer mux.Unlock()
		if err != nil {
			*errs = append(*errs, err)
			return
		}
		(*result).Rows = data

	}(&qResult, &errs)

	go func(result *dal.QueryPagerResult, errs *[]error) {
		defer wg.Done()
		var count int64
		row := GDB.QueryRow(sqlText[1], values...)
		err := row.Scan(&count)
		mux.Lock()
		defer mux.Unlock()
		if err != nil {
			*errs = append(*errs, err)
			return
		}
		(*result).Total = count
	}(&qResult, &errs)

	wg.Wait()

	if len(errs) > 0 {
		return qResult, mp.Error(funcName, errs[0].Error())
	}

	return qResult, nil
}

func (mp *MysqlProvider) Query(entity dal.QueryEntity) (interface{}, error) {
	switch entity.ResultType {
	case dal.Single:
		return mp.Single(entity)
	case dal.List:
		return mp.List(entity)
	case dal.Pager:
		return mp.Pager(entity)
	}
	return nil, mp.Error("Query", "The unknown `ResultType`")
}

func (mp *MysqlProvider) Exec(entity dal.TranEntity) (result dal.TranResult) {
	funcName := "Exec"
	if entity.Table == "" {
		result.Error = mp.Error(funcName, "`Table` can't be empty!")
		return
	}
	var (
		sqlText string
		values  []interface{}
		err     error
	)
	switch entity.Operate {
	case dal.A:
		sqlText, values, err = mp.getInsertSQL(entity)
	case dal.U:
		sqlText, values, err = mp.getUpdateSQL(entity)
	case dal.D:
		sqlText, values, err = mp.getDeleteSQL(entity)
	}
	if err != nil {
		result.Error = err
	}
	sqlResult, err := GDB.Exec(sqlText, values...)
	if err != nil {
		result.Error = mp.Error(funcName, err.Error())
		return
	}
	if entity.Operate == dal.A {
		result.Result, err = sqlResult.LastInsertId()
	} else {
		result.Result, err = sqlResult.RowsAffected()
	}
	if err != nil {
		result.Error = mp.Error(funcName, err.Error())
	}
	return
}

func (mp *MysqlProvider) ExecTrans(entities []dal.TranEntity) (result dal.TranResult) {
	funcName := "ExecTrans"
	if len(entities) == 0 {
		result.Error = mp.Error(funcName, "`entities` can't be empty!")
		return
	}
	var (
		sqlText    string
		values     []interface{}
		err        error
		affectNums int64
	)
	tx, err := GDB.Begin()
	if err != nil {
		result.Error = mp.Error(funcName, err.Error())
		return
	}
	for i, l := 0, len(entities); i < l; i++ {
		entity := entities[i]
		switch entity.Operate {
		case dal.A:
			sqlText, values, err = mp.getInsertSQL(entity)
		case dal.U:
			sqlText, values, err = mp.getUpdateSQL(entity)
		case dal.D:
			sqlText, values, err = mp.getDeleteSQL(entity)
		}
		if err != nil {
			break
		}
		sqlResult, err := tx.Exec(sqlText, values...)
		if err != nil {
			break
		}
		rowsAffected, _ := sqlResult.RowsAffected()
		affectNums += rowsAffected
	}
	if err != nil {
		tx.Rollback()
		result.Error = mp.Error(funcName, err.Error())
		return
	}
	tx.Commit()
	result.Result = affectNums
	return
}

func init() {
	// dal.RegisterDBProvider(dal.MYSQL, new(MysqlProvider))
}
