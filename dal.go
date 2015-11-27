package dal

import (
	"errors"
)

// QueryProvider 提供数据库查询接口
type QueryProvider interface {
	// Single 查询单条数据
	Single(QueryEntity) (map[string]string, error)
	// AssignSingle 将查询结果解析到对应的指针地址(数据类型包括：map[string]string,map[string]interface{},struct)
	AssignSingle(QueryEntity, interface{}) error
	// List 查询列表数据
	List(QueryEntity) ([]map[string]string, error)
	// AssignList 将查询结果解析到对应的指针地址(数据类型包括：[]map[string]string,[]map[string]interface{},[]struct)
	AssignList(QueryEntity, interface{}) error
	// Pager 查询分页数据
	Pager(QueryEntity) (QueryPagerResult, error)
	// Query 查询数据（根据QueryResultType返回数据结果类型）
	Query(QueryEntity) (interface{}, error)
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
