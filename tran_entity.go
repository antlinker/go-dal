package dal

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
func NewTranAEntity(table string, fieldsValue interface{}) (TranEntity, error) {
	entity := TranEntity{
		Table:   table,
		Operate: A,
	}
	fields, err := ConvFieldsKv(fieldsValue)
	if err != nil {
		return entity, err
	}
	entity.FieldsValue = fields
	return entity, nil
}

// NewTranUEntity 创建更新实体
func NewTranUEntity(table string, fieldsValue interface{}, cond QueryCondition) (TranEntity, error) {
	entity := TranEntity{
		Table:     table,
		Operate:   U,
		Condition: cond,
	}
	fields, err := ConvFieldsKv(fieldsValue)
	if err != nil {
		return entity, err
	}
	entity.FieldsValue = fields
	return entity, nil
}

// NewTranUEntity 创建删除实体
func NewTranDEntity(table string, cond QueryCondition) TranEntity {
	return TranEntity{
		Table:     table,
		Operate:   D,
		Condition: cond,
	}
}

// TranEntity 提供事务性操作结构体
type TranEntity struct {
	Table       string
	Operate     TranOperate
	FieldsValue map[string]interface{}
	Condition   QueryCondition
}
