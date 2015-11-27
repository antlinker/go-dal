package mysql

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"

	"github.com/LyricTian/go-dal"
)

func (mp *MysqlProvider) getInsertSQL(entity dal.TranEntity) (sqlText string, values []interface{}, err error) {
	if len(entity.FieldsValue) == 0 {
		err = mp.Error("getInsertSql", "`FieldsValue` can't be empty!")
		return
	}
	var (
		fields       []string
		placeholders []string
	)
	for k, v := range entity.FieldsValue {
		fields = append(fields, k)
		placeholders = append(placeholders, "?")
		values = append(values, v)
	}
	sqlText = fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", entity.Table, strings.Join(fields, ","), strings.Join(placeholders, ","))
	return
}

func (mp *MysqlProvider) getUpdateSQL(entity dal.TranEntity) (sqlText string, values []interface{}, err error) {
	if len(entity.FieldsValue) == 0 {
		err = mp.Error("getUpdateSql", "`FieldsValue` can't be empty!")
		return
	}
	var (
		fields []string
	)
	for k, v := range entity.FieldsValue {
		fields = append(fields, fmt.Sprintf("%s=?", k))
		values = append(values, v)
	}
	condSQL, condValues, err := mp.parseCondition(entity.Condition)
	if err != nil {
		return
	}
	values = append(values, condValues)
	sqlText = fmt.Sprintf("UPDATE %s SET %s %s", entity.Table, strings.Join(fields, ","), condSQL)
	return
}

func (mp *MysqlProvider) getDeleteSQL(entity dal.TranEntity) (sqlText string, values []interface{}, err error) {
	sqlText, values, err = mp.parseCondition(entity.Condition)
	if err != nil {
		return
	}
	sqlText = fmt.Sprintf("DELETE FROM %s %s", entity.Table, sqlText)
	return
}

func (mp *MysqlProvider) parseCondition(cond dal.QueryCondition) (sqlText string, values []interface{}, err error) {
	funcName := "parseCondition"
	switch cond.CType {
	case dal.COND_KV:
		if len(cond.FieldsKv) == 0 {
			err = mp.Error(funcName, "`FieldsKv` can't be empty!")
			return
		}
		var (
			fields []string
		)
		for k, v := range cond.FieldsKv {
			fields = append(fields, fmt.Sprintf("%s=?", k))
			values = append(values, v)
		}
		sqlText = fmt.Sprintf("WHERE %s", strings.Join(fields, " and "))
	case dal.COND_CV:
		if cond.Condition == "" {
			err = mp.Error(funcName, "`Condition` can't be empty!")
			return
		}
		sqlText = cond.Condition
		values = cond.Values
	default:
		err = mp.Error(funcName, "`QueryCondition` can't be empty!")
	}
	return
}

func (mp *MysqlProvider) parseQuerySQL(entity dal.QueryEntity) (sqlText []string, values []interface{}) {
	if entity.FieldsSelect == "" {
		entity.FieldsSelect = "*"
	}
	condSQL, condValues, _ := mp.parseCondition(entity.Condition)

	querySQL := fmt.Sprintf("SELECT %s FROM %s %s", entity.FieldsSelect, entity.Table, condSQL)
	switch entity.ResultType {
	case dal.Single:
		sqlText = append(sqlText, fmt.Sprintf("SELECT * FROM (%s) AS NewTable LIMIT 1", sqlText))
	case dal.Pager:
		pageSize := entity.PagerParam.PageSize
		pageIndex := entity.PagerParam.PageIndex
		sqlText = append(sqlText, fmt.Sprintf("SELECT * FROM (%s) AS NewTable LIMIT %d,%d", sqlText, (pageIndex-1)*pageSize+1, pageSize))
		sqlText = append(sqlText, fmt.Sprintf("SELECT COUNT(*) 'Count' FROM %s %s", entity.Table, condSQL))
	default:
		sqlText = append(sqlText, querySQL)
	}
	values = append(values, condValues...)

	return
}

func (mp *MysqlProvider) parseQueryRows(rows *sql.Rows) ([]map[string]sql.RawBytes, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	var (
		datas []map[string]sql.RawBytes
		l     = len(columns)
	)
	values := make([]sql.RawBytes, l)
	scanArgs := make([]interface{}, l)
	for i := 0; i < l; i++ {
		scanArgs[i] = &values[i]
	}
	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, err
		}
		data := make(map[string]sql.RawBytes)
		for i := 0; i < l; i++ {
			data[columns[i]] = values[i]
		}
		datas = append(datas, data)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return datas, nil
}

func (mp *MysqlProvider) convQueryData(data *[]map[string]sql.RawBytes) (rData []map[string]string) {
	l := len(*data)
	rData = make([]map[string]string, l)
	for i := 0; i < l; i++ {
		rItem := make(map[string]string)
		for k, v := range (*data)[i] {
			var val string
			if v != nil {
				val = bytes.NewBuffer(v).String()
			}
			rItem[k] = val
		}
		rData[i] = rItem
	}
	return
}

func (mp *MysqlProvider) queryData(query string, values ...interface{}) ([]map[string]string, error) {
	rows, err := GDB.Query(query, values...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	datas, err := mp.parseQueryRows(rows)
	if err != nil {
		return nil, err
	}
	if len(datas) == 0 {
		return make([]map[string]string, 0), nil
	}
	rData := mp.convQueryData(&datas)
	return rData, nil
}
