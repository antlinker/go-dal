package dal

import (
	"errors"
	"reflect"
)

var (
	ErrInvalidValue = errors.New("Invalid values!")
)

// ConvFieldsKv 键值数据转换
// fieldsKv 数据类型(map[string]interface{} or map[string]string or struct)
// 如果fieldKv为struct类型，只保留非零值字段
func ConvFieldsKv(fieldsKv interface{}) (map[string]interface{}, error) {
	if v, ok := fieldsKv.(map[string]interface{}); ok {
		return v, nil
	} else if v, ok := fieldsKv.(map[string]string); ok {
		fieldsKv := make(map[string]interface{})
		for fk, fv := range v {
			fieldsKv[fk] = fv
		}
		return fieldsKv, nil
	} else {
		fValue := reflect.ValueOf(fieldsKv)
		fValue = reflect.Indirect(fValue)
		if fValue.IsNil() ||
			!fValue.IsValid() ||
			fValue.Kind() != reflect.Struct {
			return nil, ErrInvalidValue
		}
		fieldKv := make(map[string]interface{})
		fType := fValue.Type()
		for i := 0; i < fType.NumField(); i++ {
			field := fType.Field(i)
			val := fValue.FieldByName(field.Name)
			if reflect.DeepEqual(reflect.Zero(field.Type).Interface(), val) {
				continue
			}
			fieldKv[field.Name] = val
		}
		return fieldKv, nil
	}
}

// CondType 查询条件类型标识
type CondType byte

const (
	COND_KV CondType = iota + 1
	COND_CV
)

// NewFieldsKvCondition 获取键值查询条件实例
// fieldsKv 数据类型(map[string]interface{} or map[string]string or struct)
// 如果fieldKv为struct类型，只保留非零值字段
func NewFieldsKvCondition(fieldsKv interface{}) (QueryCondition, error) {
	var queryC QueryCondition
	queryC.CType = COND_KV
	fields, err := ConvFieldsKv(fieldsKv)
	if err != nil {
		return queryC, err
	}
	queryC.FieldsKv = fields
	return queryC, nil
}

// NewCondition 获取查询条件
// condition 查询条件
// values 格式化参数
func NewCondition(condition string, values ...interface{}) QueryCondition {
	return QueryCondition{
		CType:     COND_CV,
		Condition: condition,
		Values:    values,
	}
}

// QueryCondition 查询条件
type QueryCondition struct {
	CType     CondType
	FieldsKv  map[string]interface{}
	Condition string
	Values    []interface{}
}
