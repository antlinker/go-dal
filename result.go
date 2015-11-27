package dal

// QueryPagerResult 分页查询结果类型
type QueryPagerResult struct {
	Rows  []map[string]string `json:"rows"`
	Total int64               `json:"total"`
}

// TranResult 事务执行结果
type TranResult struct {
	Result int64
	Error  error
}
