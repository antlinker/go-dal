package dal

import (
	"errors"

	"github.com/antlinker/go-dal/utils"
)

var (
	ErrInvalidValue = errors.New("Invalid values!")
)

// CondType 查询条件类型标识
type CondType byte

const (
	COND_KV CondType = iota + 1
	COND_CV
)

// NewFieldsKvCondition 获取键值查询条件实例
// fieldsKv 数据类型(map[string]interface{} or map[string]string or struct)
// 如果fieldsKv为struct类型，只保留非零值字段
func NewFieldsKvCondition(fieldsKv interface{}) QueryConditionResult {
	var (
		result QueryConditionResult
		queryC QueryCondition
		fields map[string]interface{}
	)
	err := utils.NewDecoder(fieldsKv).Decode(&fields)
	if err != nil {
		result.Error = err
		return result
	}
	queryC.CType = COND_KV
	queryC.FieldsKv = fields
	result.Condition = queryC
	return result
}

// NewCondition 获取查询条件
// condition 查询条件
// values 格式化参数
func NewCondition(condition string, values ...interface{}) QueryConditionResult {
	var result QueryConditionResult
	result.Condition = QueryCondition{
		CType:     COND_CV,
		Condition: condition,
		Values:    values,
	}
	return result
}

// QueryCondition 查询条件
type QueryCondition struct {
	CType     CondType
	FieldsKv  map[string]interface{}
	Condition string
	Values    []interface{}
}
