package dal

import "strings"

// QueryResultType 查询结果类型
type QueryResultType byte

const (
	// Single 单条数据
	Single QueryResultType = 1 << iota
	// List 列表数据
	List
	// Pager 分页数据
	Pager
)

// NewQueryEntity 创建新的查询实体
func NewQueryEntity(table string, cond QueryCondition, fields ...string) func(resultType ...QueryResultType) QueryEntityResult {
	return func(resultType ...QueryResultType) QueryEntityResult {
		var result QueryEntityResult
		entity := QueryEntity{
			Table:        table,
			FieldsSelect: strings.Join(fields, ","),
			Condition:    cond,
		}
		if len(resultType) > 0 {
			entity.ResultType = resultType[0]
		}
		result.Entity = entity
		return result
	}
}

// NewQueryPagerEntity 创建新的分页查询实体
func NewQueryPagerEntity(table string, cond QueryCondition, pagerParam PagerParam, fields ...string) QueryEntityResult {
	var result QueryEntityResult
	result.Entity = QueryEntity{
		Table:        table,
		FieldsSelect: strings.Join(fields, ","),
		Condition:    cond,
		ResultType:   Pager,
		PagerParam:   pagerParam,
	}
	return result
}

// NewPagerParam 创建新的分页参数
func NewPagerParam(pageIndex, pageSize int) PagerParam {
	if pageIndex <= 0 {
		pageIndex = 1
	}
	if pageSize <= 0 {
		pageSize = 15
	}
	return PagerParam{
		PageIndex: pageIndex,
		PageSize:  pageSize,
	}
}

// PagerParam 分页参数
type PagerParam struct {
	PageIndex int
	PageSize  int
}

// QueryEntity 提供数据查询结构体
type QueryEntity struct {
	Table        string
	FieldsSelect string
	Condition    QueryCondition
	ResultType   QueryResultType
	PagerParam   PagerParam
}
