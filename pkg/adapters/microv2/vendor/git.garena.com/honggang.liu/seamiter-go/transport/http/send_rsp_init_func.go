package http

import (
	"git.garena.com/honggang.liu/seamiter-go/core/config"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/transport/http/rsp"
	"git.garena.com/honggang.liu/seamiter-go/util"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

var (
	sendRspInitFuncInst = new(sendRspInitFunc)
)

type sendRspInitFunc struct {
	isInitialized util.AtomicBool
}

func (f sendRspInitFunc) Initial() error {
	if !f.isInitialized.CompareAndSet(false, true) {
		return nil
	}
	if config.CloseAll() {
		logging.Warn("[fetchRuleInitFunc] WARN: Sdk closeAll is set true")
		return nil
	}
	if !config.OpenConnectDashboard() {
		logging.Warn("[SendRspInitFunc] WARN: OpenConnectDashboard is false")
		return nil
	}
	rspSender := rsp.NewSimpleHttpRspSender()
	if rspSender == nil {
		logging.Warn("[SendRspInitFunc] WARN: No RuleCenter loaded")
		return errors.New("[SendRspInitFunc] WARN: No RuleCenter loaded")
	}
	rspSender.BeforeStart()
	interval := f.retrieveInterval()
	//延迟10s执行，等待配置文件的初始化
	var metricTimer = time.NewTimer(time.Millisecond * 3)
	go func() {
		for {
			<-metricTimer.C
			metricTimer.Reset(time.Millisecond * time.Duration(interval)) //interval秒心跳防止过期
			_, err := rspSender.SendRsp()
			if err != nil {
				logging.Warn("[SendRspInitFunc] WARN: SendMetric error", "err", err.Error())
				continue
			}
		}
	}()
	return nil
}

func (f sendRspInitFunc) Order() int {
	return 2
}

func (f sendRspInitFunc) retrieveInterval() uint64 {
	intervalInConfig := config.SendRspApiPathIntervalMs()
	if intervalInConfig > 0 {
		logging.Info("[SendRspInitFunc] Using fetch rule interval in sea config property: " + strconv.FormatUint(intervalInConfig, 10))
		return intervalInConfig
	}
	logging.Info("[SendRspInitFunc] Fetch interval not configured in config property or invalid, using sender default: " + strconv.FormatUint(config.DefaultFetchRuleIntervalMs, 10))
	return config.DefaultSendRspIntervalMs
}

func GetSendRspInitFuncInst() *sendRspInitFunc {
	return sendRspInitFuncInst
}
