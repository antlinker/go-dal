package mysql

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/antlinker/go-dal"
	"github.com/antlinker/go-dal/utils"

	_ "github.com/go-sql-driver/mysql"
)

const (
	DefaultMaxOpenConns    = 0
	DefaultMaxIdleConns    = 500
	DefaultConnMaxLifetime = time.Hour * 2
)

var (
	GDB *sql.DB
)

type Config struct {
	// DataSource 数据库连接
	DataSource string `json:"datasource"`
	// MaxOpenConns 打开最大连接数
	MaxOpenConns int `json:"maxopen"`
	// MaxIdleConns 连接池保持连接数量
	MaxIdleConns int `json:"maxidle"`
	// ConnMaxLifetime 连接池的生命周期
	ConnMaxLifetime time.Duration `json:"maxlifetime"`
	// IsPrint 是否打印SQL
	IsPrint bool `json:"print"`
}

type MysqlProvider struct {
	config Config
	lg     *log.Logger
}

func (mp *MysqlProvider) Error(errInfo string) error {
	return fmt.Errorf("[go-dal:mysql]%s", errInfo)
}

func (mp *MysqlProvider) PrintSQL(query string, values ...interface{}) {
	msg := fmt.Sprintf("Query SQL:\n%s \nQuery Params:%v", query, values)
	mp.lg.Println(msg)
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
	err = db.Ping()
	if err != nil {
		return err
	}
	if v := cfg.MaxOpenConns; v < 0 {
		cfg.MaxOpenConns = DefaultMaxOpenConns
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	if v := cfg.MaxIdleConns; v <= 0 {
		cfg.MaxIdleConns = DefaultMaxIdleConns
	}
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	if v := cfg.ConnMaxLifetime; v <= 0 {
		cfg.ConnMaxLifetime = DefaultConnMaxLifetime
	}
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	mp.config = cfg
	mp.lg = log.New(os.Stdout, "[go-dal-mysql]", log.Ltime)
	GDB = db
	return nil
}

func (mp *MysqlProvider) Single(entity dal.QueryEntity) (map[string]string, error) {
	if entity.ResultType != dal.QSingle {
		entity.ResultType = dal.QSingle
	}
	sqlText, values := mp.parseQuerySQL(entity)
	if mp.config.IsPrint {
		mp.PrintSQL(sqlText[0], values...)
	}
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
	return utils.NewDecoder(&data).Decode(output)
}

func (mp *MysqlProvider) List(entity dal.QueryEntity) ([]map[string]string, error) {
	if entity.ResultType != dal.QList {
		entity.ResultType = dal.QList
	}
	sqlText, values := mp.parseQuerySQL(entity)
	if mp.config.IsPrint {
		mp.PrintSQL(sqlText[0], values...)
	}
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
	return utils.NewDecoder(&data).Decode(output)
}

func (mp *MysqlProvider) Pager(entity dal.QueryEntity) (dal.QueryPagerResult, error) {
	var qResult dal.QueryPagerResult
	if entity.ResultType != dal.QPager {
		entity.ResultType = dal.QPager
	}
	sqlText, values := mp.parseQuerySQL(entity)
	if mp.config.IsPrint {
		mp.PrintSQL(sqlText[0], values...)
		mp.PrintSQL(sqlText[1], values...)
	}
	var (
		errs []error
		mux  = new(sync.RWMutex)
		wg   = new(sync.WaitGroup)
	)
	wg.Add(2)

	go func(result *dal.QueryPagerResult, errs *[]error) {
		defer wg.Done()
		rData := make([]map[string]interface{}, 0)
		data, err := mp.queryData(sqlText[0], values...)
		mux.Lock()
		defer mux.Unlock()
		if err != nil {
			*errs = append(*errs, err)
			return
		}
		if len(data) > 0 {
			err = utils.NewDecoder(data).Decode(&rData)
			if err != nil {
				*errs = append(*errs, err)
				return
			}
		}
		(*result).Rows = rData
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
	case dal.QSingle:
		return mp.Single(entity)
	case dal.QList:
		return mp.List(entity)
	case dal.QPager:
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
	case dal.TA:
		sqlText, values, err = mp.getInsertSQL(entity)
	case dal.TU:
		sqlText, values, err = mp.getUpdateSQL(entity)
	case dal.TD:
		sqlText, values, err = mp.getDeleteSQL(entity)
	}
	if err != nil {
		result.Error = err
	}
	if mp.config.IsPrint {
		mp.PrintSQL(sqlText, values...)
	}
	sqlResult, err := GDB.Exec(sqlText, values...)
	if err != nil {
		result.Error = mp.Error(err.Error())
		return
	}
	if entity.Operate == dal.TA {
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
		case dal.TA:
			sqlText, values, err = mp.getInsertSQL(entity)
		case dal.TU:
			sqlText, values, err = mp.getUpdateSQL(entity)
		case dal.TD:
			sqlText, values, err = mp.getDeleteSQL(entity)
		}
		if err != nil {
			break
		}
		if mp.config.IsPrint {
			mp.PrintSQL(sqlText, values...)
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
