package http

import (
	"github.com/liuhailove/seamiter-golang/core/config"
	"github.com/liuhailove/seamiter-golang/ext/datasource/file"
	"github.com/liuhailove/seamiter-golang/logging"
	"github.com/liuhailove/seamiter-golang/util"
)

var (
	defaultDatasourceInitFuncInst = new(defaultDatasourceInitFunc)
)

type defaultDatasourceInitFunc struct {
	isInitialized util.AtomicBool
}

func (d defaultDatasourceInitFunc) Initial() error {
	if !d.isInitialized.CompareAndSet(false, true) {
		return nil
	}
	if config.CloseAll() {
		logging.Warn("[defaultDatasourceInitFunc] WARN: Sdk closeAll is set true")
		return nil
	}
	// 默认持久化加载
	file.Initialize()
	return nil
}

func (d defaultDatasourceInitFunc) Order() int {
	return 10
}

func (d defaultDatasourceInitFunc) ImmediatelyLoadOnce() error {
	return nil
}

func GetDefaultDatasourceInitFuncInst() *defaultDatasourceInitFunc {
	return defaultDatasourceInitFuncInst
}
