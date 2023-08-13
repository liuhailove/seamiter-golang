package http

import (
	"github.com/liuhailove/seamiter-golang/core/config"
	"github.com/liuhailove/seamiter-golang/logging"
	"github.com/liuhailove/seamiter-golang/transport/http/rsp"
	"github.com/liuhailove/seamiter-golang/util"
	"github.com/pkg/errors"
	"runtime"
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
		defer func() {
			if err := recover(); err != nil {
				var buf [4096]byte
				n := runtime.Stack(buf[:], false)
				logging.Warn("sendRspInitFunc worker exit from panic", "err", string(buf[:n]))
			}
		}()
		for {
			<-metricTimer.C
			metricTimer.Reset(time.Millisecond * time.Duration(interval)) //interval秒心跳防止过期
			util.Try(func() {
				_, err := rspSender.SendRsp()
				if err != nil {
					logging.Warn("[SendRspInitFunc] WARN: SendRspInitFunc error", "err", err.Error())
				}
			}).CatchAll(func(err error) {
				logging.Error(err, "[SendRspInitFunc] WARN: SendRspInitFunc error", "err", err.Error())
			})
		}
	}()
	return nil
}

func (f sendRspInitFunc) Order() int {
	return 2
}

func (f sendRspInitFunc) ImmediatelyLoadOnce() error {
	return nil
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
