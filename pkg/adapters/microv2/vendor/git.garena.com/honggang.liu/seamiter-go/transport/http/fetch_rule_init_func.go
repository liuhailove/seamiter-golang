package http

import (
	"git.garena.com/honggang.liu/seamiter-go/core/config"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/transport/http/rule"
	"git.garena.com/honggang.liu/seamiter-go/util"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

var (
	fetchRuleInitFuncInst = new(fetchRuleInitFunc)
)

type fetchRuleInitFunc struct {
	isInitialized util.AtomicBool
}

func (f fetchRuleInitFunc) Initial() error {
	if !f.isInitialized.CompareAndSet(false, true) {
		return nil
	}
	if config.CloseAll() {
		logging.Warn("[fetchRuleInitFunc] WARN: Sdk closeAll is set true")
		return nil
	}
	if !config.OpenConnectDashboard() {
		logging.Warn("[FetchRuleInitFuncInst] WARN: OpenConnectDashboard is false")
		return nil
	}
	ruleSender := rule.NewSimpleHttpRuleSender()
	if ruleSender == nil {
		logging.Warn("[FetchRuleInitFuncInst] WARN: No RuleCenter loaded")
		return errors.New("[FetchRuleInitFuncInst] WARN: No RuleCenter loaded")
	}
	ruleSender.BeforeStart()
	interval := f.retrieveInterval()
	//延迟1s执行，等待配置文件的初始化
	var ruleTimer = time.NewTimer(time.Millisecond * 3000)

	// 全量拉取timer,尽量早执行
	var fetchAllRuleTimer = time.NewTimer(time.Millisecond * 1)

	go func() {
		for {
			<-ruleTimer.C
			ruleTimer.Reset(time.Millisecond * time.Duration(interval)) //interval秒心跳防止过期
			ruleVersionMap, err := ruleSender.FindMaxVersion()
			if err != nil {
				logging.Warn("[FetchRuleInitFuncInst] WARN: FindMaxVersion error", "err", err.Error())
				continue
			}
			if ruleVersionMap == nil {
				logging.Warn("[FetchRuleInitFuncInst] WARN: FindMaxVersion return nil map version")
				continue
			}
			for ruleType, latestVersion := range ruleVersionMap {
				if !ruleSender.Check(ruleType, latestVersion) {
					data, err := ruleSender.FindRule(ruleType)
					if err != nil {
						logging.Warn("[FetchRuleInitFuncInst] WARN: FindRule error", "err", err.Error())
						continue
					}
					err = ruleSender.HandleRule(ruleType, data)
					if err != nil {
						logging.Warn("[FetchRuleInitFuncInst] WARN: HandleRule error", "err", err.Error())
						continue
					}
					ruleSender.SetRuleTypeCurrentVersion(ruleType, latestVersion)
				}
			}
		}
	}()

	// 全量拉取Timer
	go func() {
		for {
			<-fetchAllRuleTimer.C
			fetchAllRuleTimer.Reset(time.Minute * time.Duration(5)) //5min复原，以便全量拉取
		}
	}()
	return nil
}

func (f fetchRuleInitFunc) Order() int {
	return 1
}

func (f fetchRuleInitFunc) retrieveInterval() uint64 {
	intervalInConfig := config.FetchRuleIntervalMs()
	if intervalInConfig > 0 {
		logging.Info("[FetchRuleInitFuncInst] Using fetch rule interval in sea config property: " + strconv.FormatUint(intervalInConfig, 10))
		return intervalInConfig
	}
	logging.Info("[FetchRuleInitFuncInst] Fetch interval not configured in config property or invalid, using sender default: " + strconv.FormatUint(config.DefaultFetchRuleIntervalMs, 10))
	return config.DefaultFetchRuleIntervalMs
}

func GetFetchRuleInitFuncInst() *fetchRuleInitFunc {
	return fetchRuleInitFuncInst
}
