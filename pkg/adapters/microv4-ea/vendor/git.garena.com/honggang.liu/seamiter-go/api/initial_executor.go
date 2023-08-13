package api

import (
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/transport/http"
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
func doInit() {
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
		} else {
			logging.Info("[InitExecutor] Executing {} with order {}", "funName", fun, "order", fun.Order())
		}
	}
}
