package dal

// QueryPagerResult 分页查询结果类型
type QueryPagerResult struct {
	Rows  []map[string]string `json:"rows"`
	Total int64               `json:"total"`
}

// ResultError 提供统一的错误处理
type ResultError struct {
	Error error
}

// TranEntityResult 提供事务实体
type TranEntityResult struct {
	ResultError
	Entity TranEntity
}

// QueryEntityResult 提供查询实体
type QueryEntityResult struct {
	ResultError
	Entity QueryEntity
}

// ConditionResult 提供查询条件处理
type QueryConditionResult struct {
	ResultError
	Condition QueryCondition
}

// TranResult 提供事务结果处理
type TranResult struct {
	ResultError
	Result int64
}
