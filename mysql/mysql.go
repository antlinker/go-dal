package mysql

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/antlinker/go-dal"
	"github.com/antlinker/go-dal/utils"

	// 引入mysql驱动
	_ "github.com/go-sql-driver/mysql"
)

// 定义默认值
const (
	DefaultMaxOpenConns    = 0
	DefaultMaxIdleConns    = 500
	DefaultConnMaxLifetime = time.Hour * 2
)

// 定义全局变量
var (
	GDB *sql.DB
)

// Config 配置参数
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

type mysqlProvider struct {
	config Config
	lg     *log.Logger
}

func (mp *mysqlProvider) PrintSQL(query string, values ...interface{}) {
	msg := fmt.Sprintf("Query SQL:\n%s \nQuery Params:%v", query, values)
	mp.lg.Println(msg)
}

func (mp *mysqlProvider) InitDB(config string) error {
	var cfg Config
	if err := json.NewDecoder(bytes.NewBufferString(config)).Decode(&cfg); err != nil {
		return err
	}
	if cfg.DataSource == "" {
		return errors.New("`datasource` can't be empty")
	}
	db, err := sql.Open("mysql", cfg.DataSource)
	if err != nil {
		return err
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

func (mp *mysqlProvider) Single(entity dal.QueryEntity) (map[string]string, error) {
	if entity.ResultType != dal.QSingle {
		entity.ResultType = dal.QSingle
	}
	sqlText, values := mp.parseQuerySQL(entity)
	if mp.config.IsPrint {
		mp.PrintSQL(sqlText[0], values...)
	}
	data, err := mp.queryData(sqlText[0], values...)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return make(map[string]string), nil
	}
	return data[0], nil
}

func (mp *mysqlProvider) SingleWithSQL(sql string, values ...interface{}) (data map[string]string, err error) {
	if mp.config.IsPrint {
		mp.PrintSQL(sql, values...)
	}
	datas, err := mp.queryData(sql, values...)
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		data = datas[0]
	}
	return
}

func (mp *mysqlProvider) AssignSingle(entity dal.QueryEntity, output interface{}) error {
	data, err := mp.Single(entity)
	if err != nil {
		return err
	}
	return utils.NewDecoder(&data).Decode(output)
}

func (mp *mysqlProvider) AssignSingleWithSQL(sql string, values []interface{}, output interface{}) (err error) {
	data, err := mp.SingleWithSQL(sql, values...)
	if err != nil {
		return
	}
	err = utils.NewDecoder(&data).Decode(output)
	return
}

func (mp *mysqlProvider) ListWithSQL(sql string, values ...interface{}) (data []map[string]string, err error) {
	if mp.config.IsPrint {
		mp.PrintSQL(sql, values...)
	}
	data, err = mp.queryData(sql, values...)
	return
}

func (mp *mysqlProvider) List(entity dal.QueryEntity) ([]map[string]string, error) {
	if entity.ResultType != dal.QList {
		entity.ResultType = dal.QList
	}
	sqlText, values := mp.parseQuerySQL(entity)
	if mp.config.IsPrint {
		mp.PrintSQL(sqlText[0], values...)
	}
	data, err := mp.queryData(sqlText[0], values...)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (mp *mysqlProvider) AssignList(entity dal.QueryEntity, output interface{}) error {
	data, err := mp.List(entity)
	if err != nil {
		return err
	}
	return utils.NewDecoder(&data).Decode(output)
}

func (mp *mysqlProvider) AssignListWithSQL(sql string, values []interface{}, output interface{}) (err error) {
	if mp.config.IsPrint {
		mp.PrintSQL(sql, values...)
	}
	data, err := mp.queryData(sql, values...)
	if err != nil {
		return
	}
	err = utils.NewDecoder(&data).Decode(output)
	return
}

func (mp *mysqlProvider) Pager(entity dal.QueryEntity) (qResult dal.QueryPagerResult, err error) {
	if entity.ResultType != dal.QPager {
		entity.ResultType = dal.QPager
	}

	sqlText, values := mp.parseQuerySQL(entity)

	var count int64
	if mp.config.IsPrint {
		mp.PrintSQL(sqlText[1], values...)
	}
	row := GDB.QueryRow(sqlText[1], values...)
	err = row.Scan(&count)
	if err != nil {
		return
	} else if count == 0 {
		return
	}
	qResult.Total = count

	if mp.config.IsPrint {
		mp.PrintSQL(sqlText[0], values...)
	}

	rData := make([]map[string]interface{}, 0)
	data, err := mp.queryData(sqlText[0], values...)
	if err != nil {
		return
	}
	if len(data) > 0 {
		err = utils.NewDecoder(data).Decode(&rData)
		if err != nil {
			return
		}
	}
	qResult.Rows = rData

	return
}

func (mp *mysqlProvider) Query(entity dal.QueryEntity) (interface{}, error) {
	switch entity.ResultType {
	case dal.QSingle:
		return mp.Single(entity)
	case dal.QList:
		return mp.List(entity)
	case dal.QPager:
		return mp.Pager(entity)
	}
	return nil, errors.New("The unknown `ResultType`")
}

func (mp *mysqlProvider) Exec(entity dal.TranEntity) (result dal.TranResult) {
	if entity.Table == "" {
		result.Error = errors.New("`Table` can't be empty")
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
		result.Error = err
		return
	}
	if entity.Operate == dal.TA {
		result.Result, err = sqlResult.LastInsertId()
	} else {
		result.Result, err = sqlResult.RowsAffected()
	}
	if err != nil {
		result.Error = err
	}
	return
}

func (mp *mysqlProvider) ExecTrans(entities []dal.TranEntity) (result dal.TranResult) {
	if len(entities) == 0 {
		result.Error = errors.New("`entities` can't be empty")
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
		result.Error = err
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
		result.Error = err
		return
	}
	tx.Commit()
	result.Result = affectNums
	return
}

func init() {
	dal.RegisterDBProvider(dal.MYSQL, new(mysqlProvider))
}
