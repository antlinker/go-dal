package mysql

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	"gopkg.in/dal.v1"
	"gopkg.in/dal.v1/utils"
	_ "github.com/go-sql-driver/mysql"
)

const (
	DefaultMaxOpenConns = 0
	DefaultMaxIdleConns = 500
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

func (mp *MysqlProvider) Error(errInfo string) error {
	return fmt.Errorf("[go-dal:mysql:mysql.go]%s", errInfo)
}

func (mp *MysqlProvider) InitDB(config string) error {
	var cfg Config
	if err := json.NewDecoder(bytes.NewBufferString(config)).Decode(&cfg); err != nil {
		return mp.Error(err.Error())
	}
	if cfg.DataSource == "" {
		return mp.Error("`datasource` can't be empty!")
	}
	db, err := sql.Open("mysql", cfg.DataSource)
	if err != nil {
		return mp.Error(err.Error())
	}
	if v := cfg.MaxOpenConns; v < 0 {
		cfg.MaxOpenConns = DefaultMaxOpenConns
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	if v := cfg.MaxIdleConns; v < 0 {
		cfg.MaxIdleConns = DefaultMaxIdleConns
	}
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	GDB = db
	return nil
}

func (mp *MysqlProvider) Single(entity dal.QueryEntity) (map[string]string, error) {
	if entity.ResultType != dal.Single {
		entity.ResultType = dal.Single
	}
	sqlText, values := mp.parseQuerySQL(entity)
	data, err := mp.queryData(sqlText[0], values...)
	if err != nil {
		return nil, mp.Error(err.Error())
	}
	if len(data) == 0 {
		return make(map[string]string), nil
	}
	return data[0], nil
}

func (mp *MysqlProvider) AssignSingle(entity dal.QueryEntity, output interface{}) error {
	data, err := mp.Single(entity)
	if err != nil {
		return err
	}
	return utils.NewDecoder(data).Decode(output)
}

func (mp *MysqlProvider) List(entity dal.QueryEntity) ([]map[string]string, error) {
	if entity.ResultType != dal.List {
		entity.ResultType = dal.List
	}
	sqlText, values := mp.parseQuerySQL(entity)
	data, err := mp.queryData(sqlText[0], values...)
	if err != nil {
		return nil, mp.Error(err.Error())
	}

	return data, nil
}

func (mp *MysqlProvider) AssignList(entity dal.QueryEntity, output interface{}) error {
	data, err := mp.List(entity)
	if err != nil {
		return err
	}
	return utils.NewDecoder(data).Decode(output)
}

func (mp *MysqlProvider) Pager(entity dal.QueryEntity) (dal.QueryPagerResult, error) {
	var qResult dal.QueryPagerResult
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
		return qResult, mp.Error(errs[0].Error())
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
	return nil, mp.Error("The unknown `ResultType`")
}

func (mp *MysqlProvider) Exec(entity dal.TranEntity) (result dal.TranResult) {
	if entity.Table == "" {
		result.Error = mp.Error("`Table` can't be empty!")
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
		result.Error = mp.Error(err.Error())
		return
	}
	if entity.Operate == dal.A {
		result.Result, err = sqlResult.LastInsertId()
	} else {
		result.Result, err = sqlResult.RowsAffected()
	}
	if err != nil {
		result.Error = mp.Error(err.Error())
	}
	return
}

func (mp *MysqlProvider) ExecTrans(entities []dal.TranEntity) (result dal.TranResult) {
	if len(entities) == 0 {
		result.Error = mp.Error("`entities` can't be empty!")
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
		result.Error = mp.Error(err.Error())
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
		result.Error = mp.Error(err.Error())
		return
	}
	tx.Commit()
	result.Result = affectNums
	return
}

func init() {
	dal.RegisterDBProvider(dal.MYSQL, new(MysqlProvider))
}
