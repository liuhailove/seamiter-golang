package http

import (
	"github.com/liuhailove/seamiter-golang/core/config"
	"github.com/liuhailove/seamiter-golang/logging"
	"github.com/liuhailove/seamiter-golang/transport/http/rule"
	"github.com/liuhailove/seamiter-golang/util"
	"github.com/pkg/errors"
	"runtime"
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
	// 立刻加载
	var ruleTimer = time.NewTimer(time.Millisecond * 1)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				var buf [4096]byte
				n := runtime.Stack(buf[:], false)
				logging.Warn("fetchRuleInitFunc worker exit from panic", "err", string(buf[:n]))
			}
		}()
		for {
			<-ruleTimer.C
			ruleTimer.Reset(time.Millisecond * time.Duration(interval)) //interval秒心跳防止过期
			util.Try(func() {
				ruleVersionMap, err := ruleSender.FindMaxVersion()
				if err != nil {
					logging.Warn("[FetchRuleInitFuncInst] WARN: FindMaxVersion error", "err", err.Error())
					return
				}
				if ruleVersionMap == nil {
					logging.Warn("[FetchRuleInitFuncInst] WARN: FindMaxVersion return nil map version")
					return
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
			}).CatchAll(func(err error) {
				logging.Error(err, "[FetchRuleInitFuncInst] error", "err", err.Error())
			})

		}
	}()
	return nil
}

func (f fetchRuleInitFunc) ImmediatelyLoadOnce() error {
	if !config.OpenConnectDashboard() {
		logging.Warn("[ImmediatelyLoadOnce] WARN: OpenConnectDashboard is false")
		return nil
	}
	ruleSender := rule.NewSimpleHttpRuleSender()
	if ruleSender == nil {
		logging.Warn("[ImmediatelyLoadOnce] WARN: No RuleCenter loaded")
		return errors.New("[FetchRuleInitFuncInst] WARN: No RuleCenter loaded")
	}
	ruleVersionMap, err := ruleSender.FindMaxVersion()
	if err != nil {
		logging.Warn("[FetchRuleInitFuncInst] WARN: FindMaxVersion error", "err", err.Error())
		return err
	}
	if ruleVersionMap == nil {
		logging.Warn("[FetchRuleInitFuncInst] WARN: FindMaxVersion return nil map version")
		return nil
	}
	for ruleType, latestVersion := range ruleVersionMap {
		if !ruleSender.Check(ruleType, latestVersion) {
			data, err := ruleSender.FindRule(ruleType)
			if err != nil && err == rule.ErrFetchNoData {
				continue
			} else if err != nil {
				logging.Warn("[FetchRuleInitFuncInst] WARN: FindRule error", "err", err.Error())
				return errors.Wrap(err, "FetchRuleInitFuncInst FindRule error")
			}
			err = ruleSender.HandleRule(ruleType, data)
			if err != nil {
				logging.Warn("[FetchRuleInitFuncInst] WARN: HandleRule error", "err", err.Error())
				return errors.Wrap(err, "FetchRuleInitFuncInst HandleRule  error")
			}
			ruleSender.SetRuleTypeCurrentVersion(ruleType, latestVersion)
		}
	}
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
