package http

import (
	"git.garena.com/honggang.liu/seamiter-go/core/config"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/transport/http/request"
	"git.garena.com/honggang.liu/seamiter-go/util"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

var (
	sendRequestInitFuncInst = new(sendRequestInitFunc)
)

type sendRequestInitFunc struct {
	isInitialized util.AtomicBool
}

func (f sendRequestInitFunc) Initial() error {
	if !f.isInitialized.CompareAndSet(false, true) {
		return nil
	}
	if config.CloseAll() {
		logging.Warn("[sendRequestInitFunc] WARN: Sdk closeAll is set true")
		return nil
	}
	if !config.OpenConnectDashboard() {
		logging.Warn("[sendRequestInitFunc] WARN: OpenConnectDashboard is false")
		return nil
	}
	requestSender := request.NewSimpleHttpRequestSender()
	if requestSender == nil {
		logging.Warn("[SendRspInitFunc] WARN: No RuleCenter loaded")
		return errors.New("[SendRspInitFunc] WARN: No RuleCenter loaded")
	}
	requestSender.BeforeStart()
	interval := f.retrieveInterval()
	//延迟10s执行，等待配置文件的初始化
	var metricTimer = time.NewTimer(time.Millisecond * 10)
	go func() {
		for {
			<-metricTimer.C
			metricTimer.Reset(time.Millisecond * time.Duration(interval)) //interval秒心跳防止过期
			_, err := requestSender.SendRequest()
			if err != nil {
				logging.Warn("[sendRequestInitFunc] WARN: SendMetric error", "err", err.Error())
				continue
			}
		}
	}()
	return nil
}

func (f sendRequestInitFunc) Order() int {
	return 10
}

func (f sendRequestInitFunc) retrieveInterval() uint64 {
	intervalInConfig := config.SendRequestApiPathIntervalMs()
	if intervalInConfig > 0 {
		logging.Info("[sendRequestInitFunc] Using fetch rule interval in sea config property: " + strconv.FormatUint(intervalInConfig, 10))
		return intervalInConfig
	}
	logging.Info("[sendRequestInitFunc] Fetch interval not configured in config property or invalid, using sender default: " + strconv.FormatUint(config.DefaultFetchRuleIntervalMs, 10))
	return config.DefaultSendRequestIntervalMs
}

func GetSendRequestInitFuncInst() *sendRequestInitFunc {
	return sendRequestInitFuncInst
}
