package api

import (
	"github.com/liuhailove/seamiter-golang/core/config"
	"github.com/liuhailove/seamiter-golang/logging"
	"github.com/liuhailove/seamiter-golang/transport/http"
)

var (
	initFuncMap = make(map[InitialFunc]int)
)

func defaultRegister() {
	Register(http.GetCommandCenterInitFuncInst(), http.GetCommandCenterInitFuncInst().Order())
	Register(http.GetHeartBeatSenderInitFuncInst(), http.GetHeartBeatSenderInitFuncInst().Order())
	Register(http.GetFetchRuleInitFuncInst(), http.GetFetchRuleInitFuncInst().Order())
	Register(http.GetSendRspInitFuncInst(), http.GetSendRspInitFuncInst().Order())
	// 发送请求体
	Register(http.GetSendRequestInitFuncInst(), http.GetSendRequestInitFuncInst().Order())
	Register(http.GetSendMetricInitFuncInst(), http.GetSendMetricInitFuncInst().Order())
	// 默认持久化加载
	Register(http.GetDefaultDatasourceInitFuncInst(), http.GetFetchRuleInitFuncInst().Order())
}

func Register(initialFunc InitialFunc, order int) {
	initFuncMap[initialFunc] = order
}

// doInit 初始化
func doInit() error {
	defaultRegister()
	var funcs []InitialFunc
	for k, v := range initFuncMap {
		for _, fun := range funcs {
			if v > fun.Order() {
				break
			}
		}
		funcs = append(funcs, k)
	}
	for _, fun := range funcs {
		err := fun.Initial()
		if err != nil {
			logging.Warn("[InitExecutor] WARN: Initialization failed", "err", err)
			return err
		} else {
			logging.Info("[InitExecutor] Executing {} with order {}", "funName", fun, "order", fun.Order())
		}
		// 如果配置了立即拉取配置文件，则立刻拉取一次，拉取失败将会直接抛出异常
		// 立即加载的原因，是考虑到部分配置在启动前就需要加载，否则会导致不可预期的问题
		if config.ImmediatelyFetch() {
			err = fun.ImmediatelyLoadOnce()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
