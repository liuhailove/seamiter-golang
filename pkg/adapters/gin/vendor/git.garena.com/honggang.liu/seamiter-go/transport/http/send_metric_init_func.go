package http

import (
	"git.garena.com/honggang.liu/seamiter-go/core/config"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/transport/http/metric"
	"git.garena.com/honggang.liu/seamiter-go/util"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

var (
	sendMetricInitFuncInst = new(sendMetricInitFunc)
)

type sendMetricInitFunc struct {
	isInitialized util.AtomicBool
}

func (f sendMetricInitFunc) Initial() error {
	if !f.isInitialized.CompareAndSet(false, true) {
		return nil
	}
	if config.CloseAll() {
		logging.Warn("[fetchRuleInitFunc] WARN: Sdk closeAll is set true")
		return nil
	}
	if !config.OpenConnectDashboard() {
		logging.Warn("[SendMetricInitFunc] WARN: OpenConnectDashboard is false")
		return nil
	}
	metricSender := metric.NewSimpleHttpMetricSender()
	if metricSender == nil {
		logging.Warn("[SendMetricInitFunc] WARN: No RuleCenter loaded")
		return errors.New("[SendMetricInitFunc] WARN: No RuleCenter loaded")
	}
	metricSender.BeforeStart()
	interval := f.retrieveInterval()
	//延迟10s执行，等待配置文件的初始化
	var metricTimer = time.NewTimer(time.Millisecond * 5)
	go func() {
		for {
			<-metricTimer.C
			metricTimer.Reset(time.Millisecond * time.Duration(interval)) //interval秒心跳防止过期
			_, err := metricSender.SendMetric()
			if err != nil {
				logging.Warn("[SendMetricInitFunc] WARN: SendMetric error", "err", err.Error())
			}
		}
	}()
	return nil
}

func (f sendMetricInitFunc) Order() int {
	return 1
}

func (f sendMetricInitFunc) retrieveInterval() uint64 {
	intervalInConfig := config.SendMetricIntervalMs()
	if intervalInConfig > 0 {
		logging.Info("[SendMetricInitFunc] Using fetch rule interval in sea config property: " + strconv.FormatUint(intervalInConfig, 10))
		return intervalInConfig
	}
	logging.Info("[SendMetricInitFunc] Fetch interval not configured in config property or invalid, using sender default: " + strconv.FormatUint(config.DefaultFetchRuleIntervalMs, 10))
	return config.DefaultSendIntervalMs
}

func GetSendMetricInitFuncInst() *sendMetricInitFunc {
	return sendMetricInitFuncInst
}
