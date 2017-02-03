package dal

import (
	"errors"
)

// QueryProvider 提供数据库查询接口
type QueryProvider interface {
	// 查询单条数据
	Single(entity QueryEntity) (map[string]string, error)
	// 查询单条数据
	SingleWithSQL(sql string, values ...interface{}) (map[string]string, error)
	// 将查询结果解析到对应的指针地址
	// (数据类型包括：map[string]string,map[string]interface{},struct)
	AssignSingle(entity QueryEntity, output interface{}) error
	// 将查询结果解析到对应的指针地址
	// (数据类型包括：map[string]string,map[string]interface{},struct)
	AssignSingleWithSQL(sql string, values []interface{}, output interface{}) error
	// 查询列表数据
	List(entity QueryEntity) ([]map[string]string, error)
	// 使用sql查询数据列表
	ListWithSQL(sql string, values ...interface{}) ([]map[string]string, error)
	// 将查询结果解析到对应的指针地址
	// (数据类型包括：[]map[string]string,[]map[string]interface{},[]struct)
	AssignList(entity QueryEntity, output interface{}) error
	// 使用sql查询数据列表
	AssignListWithSQL(sql string, values []interface{}, output interface{}) error
	// 查询分页数据
	Pager(entity QueryEntity) (QueryPagerResult, error)
	// 查询数据（根据QueryResultType返回数据结果类型）
	Query(entity QueryEntity) (interface{}, error)
}

// TranProvider 提供数据库事务操作
type TranProvider interface {
	// Exec 执行单条事务性操作
	Exec(TranEntity) TranResult
	// ExecTrans 执行多条事务性操作
	ExecTrans([]TranEntity) TranResult
}

// Provider 提供统一的数据库操作
type Provider interface {
	QueryProvider
	TranProvider
}

// DBProvider 提供DB初始化
type DBProvider interface {
	Provider
	// InitDB 数据库初始化
	// config 为配置信息（以json字符串的方式提供）
	InitDB(config string) error
}

// ProvideEngine 数据库操作引擎
type ProvideEngine string

const (
	// MYSQL mysql数据库
	MYSQL ProvideEngine = "mysql"
)

var (
	// GDAL 提供全局的Provider
	GDAL      Provider
	providers map[ProvideEngine]DBProvider
)

func init() {
	providers = make(map[ProvideEngine]DBProvider)
}

// RegisterDBProvider 注册DBProvider
func RegisterDBProvider(provideName ProvideEngine, provider DBProvider) {
	if provider == nil {
		panic("go-dal:DBProvider is nil!")
	}
	if _, ok := providers[provideName]; ok {
		panic("go-dal:DBProvider has been registered!")
	}
	providers[provideName] = provider
}

// RegisterProvider 提供全局的provider
func RegisterProvider(provideName ProvideEngine, config string) error {
	if GDAL != nil {
		return errors.New("Provider has been registered!")
	}
	provide, ok := providers[provideName]
	if !ok {
		return errors.New("Unknown provider!")
	}
	if err := provide.InitDB(config); err != nil {
		return err
	}
	GDAL = provide
	return nil
}

// Single 查询单条数据
func Single(entity QueryEntity) (map[string]string, error) {
	return GDAL.Single(entity)
}

// SingleWithSQL 查询单条数据
func SingleWithSQL(sql string, values ...interface{}) (map[string]string, error) {
	return GDAL.SingleWithSQL(sql, values...)
}

// AssignSingle 将查询结果解析到对应的指针地址
// (数据类型包括：map[string]string,map[string]interface{},struct)
func AssignSingle(entity QueryEntity, output interface{}) error {
	return GDAL.AssignSingle(entity, output)
}

// AssignSingleWithSQL 将查询结果解析到对应的指针地址
// (数据类型包括：map[string]string,map[string]interface{},struct)
func AssignSingleWithSQL(sql string, values []interface{}, output interface{}) error {
	return GDAL.AssignSingleWithSQL(sql, values, output)
}

// List 查询列表数据
func List(entity QueryEntity) ([]map[string]string, error) {
	return GDAL.List(entity)
}

// ListWithSQL 查询列表数据
func ListWithSQL(sql string, values ...interface{}) ([]map[string]string, error) {
	return GDAL.ListWithSQL(sql, values...)
}

// AssignList 将查询结果解析到对应的指针地址
// (数据类型包括：[]map[string]string,[]map[string]interface{},[]struct)
func AssignList(entity QueryEntity, output interface{}) error {
	return GDAL.AssignList(entity, output)
}

// AssignListWithSQL 将查询结果解析到对应的指针地址
// (数据类型包括：[]map[string]string,[]map[string]interface{},[]struct)
func AssignListWithSQL(sql string, values []interface{}, output interface{}) error {
	return GDAL.AssignListWithSQL(sql, values, output)
}

// Pager 查询分页数据
func Pager(entity QueryEntity) (QueryPagerResult, error) {
	return GDAL.Pager(entity)
}

// Query 查询数据
//（根据QueryResultType返回数据结果类型）
func Query(entity QueryEntity) (interface{}, error) {
	return GDAL.Query(entity)
}

// Exec 执行单条事务性操作
func Exec(entity TranEntity) TranResult {
	return GDAL.Exec(entity)
}

// ExecTrans 执行多条事务性操作
func ExecTrans(entities []TranEntity) TranResult {
	return GDAL.ExecTrans(entities)
}
