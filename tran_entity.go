package dal

import "gopkg.in/dal.v1/utils"

// TranOperate 操作
type TranOperate byte

const (
	// A 新增
	A TranOperate = 1 << iota
	// U 更新
	U
	// D 删除
	D
)

// NewTranAEntity 创建新增实体
// fieldsValue 数据类型(map[string]interface{} or map[string]string or struct)
// 如果fieldsValue为struct类型，只保留非零值字段
func NewTranAEntity(table string, fieldsValue interface{}) TranEntityResult {
	var result TranEntityResult
	entity := TranEntity{
		Table:   table,
		Operate: A,
	}
	var fields map[string]interface{}
	err := utils.NewDecoder(fieldsValue).Decode(&fields)
	if err != nil {
		result.Error = err
	}
	entity.FieldsValue = fields
	result.Entity = entity
	return result
}

// NewTranUEntity 创建更新实体
// fieldsValue 数据类型(map[string]interface{} or map[string]string or struct)
// 如果fieldsValue为struct类型，只保留非零值字段
func NewTranUEntity(table string, fieldsValue interface{}, cond QueryCondition) TranEntityResult {
	var result TranEntityResult
	entity := TranEntity{
		Table:     table,
		Operate:   U,
		Condition: cond,
	}
	var fields map[string]interface{}
	err := utils.NewDecoder(fieldsValue).Decode(&fields)
	if err != nil {
		result.Error = err
		return result
	}
	entity.FieldsValue = fields
	result.Entity = entity
	return result
}

// NewTranUEntity 创建删除实体
func NewTranDEntity(table string, cond QueryCondition) TranEntityResult {
	var result TranEntityResult
	result.Entity = TranEntity{
		Table:     table,
		Operate:   D,
		Condition: cond,
	}
	return result
}

// TranEntity 提供事务性操作结构体
type TranEntity struct {
	Table       string
	Operate     TranOperate
	FieldsValue map[string]interface{}
	Condition   QueryCondition
}
