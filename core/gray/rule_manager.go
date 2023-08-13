package gray

import (
	"fmt"
	"github.com/liuhailove/seamiter-golang/logging"
	"github.com/liuhailove/seamiter-golang/util"
	"github.com/pkg/errors"
	"reflect"
	"sort"
	"strings"
	"sync"
)

// TrafficControllerGenFunc 根据指定的Rule产生对应的controller
type TrafficControllerGenFunc func(*Rule) (*TrafficSelectorController, error)

type trafficControllerGenKey struct {
	routerStrategy RouterStrategy
}

// TrafficControllerMap 存储TrafficSelectorController的map
type TrafficControllerMap map[string][]*TrafficSelectorController

var (
	tcsGenFuncMap = make(map[trafficControllerGenKey]TrafficControllerGenFunc, 4)
	tcsMap        = make(TrafficControllerMap)
	tcMux         = new(sync.RWMutex)
	currentRules  = make(map[string][]*Rule, 0)
	updateRuleMux = new(sync.RWMutex)
)

func init() {
	// 根据现有的流量选择策略生成流量选择器map
	tcsGenFuncMap[trafficControllerGenKey{routerStrategy: ConditionRouter}] = func(rule *Rule) (*TrafficSelectorController, error) {
		tsc, err := NewTrafficSelectorController(rule)
		if err != nil || tsc == nil {
			return nil, err
		}
		tsc.flowCalculator = NewConditionTrafficSelector(tsc, rule)
		return tsc, nil
	}
	tcsGenFuncMap[trafficControllerGenKey{routerStrategy: TagRouter}] = func(rule *Rule) (*TrafficSelectorController, error) {
		tsc, err := NewTrafficSelectorController(rule)
		if err != nil || tsc == nil {
			return nil, err
		}
		tsc.flowCalculator = NewTagTrafficSelector(tsc, rule)
		return tsc, nil
	}
	tcsGenFuncMap[trafficControllerGenKey{routerStrategy: WeightRouter}] = func(rule *Rule) (*TrafficSelectorController, error) {
		tsc, err := NewTrafficSelectorController(rule)
		if err != nil || tsc == nil {
			return nil, err
		}
		tsc.flowCalculator = NewWeightTrafficSelector(tsc, rule)
		return tsc, nil
	}
}

func logRuleUpdate(m map[string][]*Rule) {
	rules := make([]*Rule, 0, 8)
	for _, rs := range m {
		if len(rs) == 0 {
			continue
		}
		rules = append(rules, rs...)
	}
	if len(rules) == 0 {
		logging.Info("[GrayRuleManager] Gray rules were cleared")
	} else {
		logging.Info("[GrayRuleManager] Gray rules were loaded", "rules", rules)
	}
}

func onRuleUpdate(rawResRulesMap map[string][]*Rule) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()
	// 忽略无效规则
	validResRulesMap := make(map[string][]*Rule, len(rawResRulesMap))
	for res, rules := range rawResRulesMap {
		validResRules := make([]*Rule, 0, len(rules))
		for _, rule := range rules {
			if err := IsValidRule(rule); err != nil {
				logging.Warn("[Gray onRuleUpdate] Ignoring invalid gray rule", "rule", rule, "reason", err.Error())
				continue
			}
			validResRules = append(validResRules, rule)
		}
		if len(validResRules) > 0 {
			validResRulesMap[res] = validResRules
		}
	}

	start := util.CurrentTimeNano()
	tcMux.RLock()
	tcsMapClone := make(TrafficControllerMap, len(validResRulesMap))
	for res, tcs := range tcsMap {
		resTcsClone := make([]*TrafficSelectorController, 0, len(tcs))
		resTcsClone = append(resTcsClone, tcs...)
		tcsMapClone[res] = resTcsClone
	}
	tcMux.RUnlock()

	m := make(TrafficControllerMap, len(validResRulesMap))
	for res, rulesOfRes := range validResRulesMap {
		newTscOfRes := buildResourceTrafficSelectingController(res, rulesOfRes, tcsMapClone[res])
		if len(newTscOfRes) > 0 {
			m[res] = newTscOfRes
		}
	}
	tcMux.Lock()
	tcsMap = m
	tcMux.Unlock()
	currentRules = rawResRulesMap
	if logging.DebugEnabled() {
		logging.Debug("[Gray onRuleUpdate] Time statistic(ns) for updating gray rule", "timeCost", util.CurrentTimeNano()-start)
	}
	logRuleUpdate(validResRulesMap)
	return nil
}

// LoadRules 为rule manager 加载灰度规则，所有之前的规则将会被替换.
// 返回值第一个表示是否被真正加载了，如果规则和之前相同，则返回false
func LoadRules(rules []*Rule) (bool, error) {
	resRulesMap := make(map[string][]*Rule, 16)
	for _, rule := range rules {
		resRules, exist := resRulesMap[rule.Resource]
		if !exist {
			resRules = make([]*Rule, 0, 1)
		}
		resRulesMap[rule.Resource] = append(resRules, rule)
	}

	updateRuleMux.Lock()
	defer updateRuleMux.Unlock()
	isEqual := reflect.DeepEqual(currentRules, resRulesMap)
	if isEqual {
		logging.Info("[Gray] Load rules is the same with current rules, so ignore load operation.")
		return false, nil
	}
	err := onRuleUpdate(resRulesMap)
	return true, err
}

// LoadRulesOfResource 为rule manager加载给定resource的灰度规则，同时之前resource的全部规则将会被替换.
// 第一个返回值表示是否加载更高，如果规则相同，则返回false
func LoadRulesOfResource(res string, rules []*Rule) (bool, error) {
	if len(res) == 0 {
		return false, errors.New("empty resource")
	}
	updateRuleMux.Lock()
	defer updateRuleMux.Unlock()
	// clear resource rules
	if len(rules) == 0 {
		// 从currentRules清理resource
		delete(currentRules, res)
		// 清理tcMap
		tcMux.Lock()
		delete(tcsMap, res)
		tcMux.Unlock()
		logging.Info("[Gray] clear resource level rules", "resource", res)
		return true, nil
	}
	// load resource level rules
	isEqual := reflect.DeepEqual(currentRules[res], rules)
	if isEqual {
		logging.Info("[Gray] Load resource level rules is the same with current resource level rules, so ignore load operation.")
		return false, nil
	}
	err := onResourceRuleUpdate(res, rules)
	return true, err
}

func onResourceRuleUpdate(res string, rawResRules []*Rule) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	validResRules := make([]*Rule, 0, len(rawResRules))
	for _, rule := range rawResRules {
		if err := IsValidRule(rule); err != nil {
			logging.Warn("[Gray onResourceUpdate] Ignoring invalid gray rule", "rule", rule, "reason", err.Error())
			continue
		}
		validResRules = append(validResRules, rule)
	}
	start := util.CurrentTimeNano()
	oldResTcs := make([]*TrafficSelectorController, 0)
	tcMux.RLock()
	oldResTcs = append(oldResTcs, tcsMap[res]...)
	tcMux.RUnlock()
	newResTcs := buildResourceTrafficSelectingController(res, validResRules, oldResTcs)

	tcMux.Lock()
	if len(newResTcs) == 0 {
		delete(tcsMap, res)
	} else {
		tcsMap[res] = newResTcs
	}
	tcMux.Unlock()
	currentRules[res] = rawResRules
	if logging.DebugEnabled() {
		logging.Debug("[Gray onResourceRuleUpdate] Time statistic(ns) for updating flow rule", "timeCost", util.CurrentTimeNano()-start)
	}
	logging.Info("[Gray] load resource level rules", "resource", res, "validResRules", validResRules)
	return nil
}

// getRules 返回全部的规则.任何规则的变动都会影响Gray模块
// getRules 是一个内部接口
func getRules() []*Rule {
	tcMux.RLock()
	defer tcMux.Unlock()
	return rulesFrom(tcsMap)
}

// getRuleOfResource 返回特定res对应的规则，任何规则的变动都会影响Gray模块
// getRuleOfResource 是一个内部接口
func getRuleOfResource(res string) []*Rule {
	tcMux.RLock()
	defer tcMux.RUnlock()

	resTsc, exist := tcsMap[res]
	if !exist {
		return nil
	}
	ret := make([]*Rule, 0, len(resTsc))
	for _, tc := range resTsc {
		ret = append(ret, tc.BoundRule())
	}
	return ret
}

// GetRules 返回全部规则的copy.
// 如果用户变更规则，并不会影响gray模块
func GetRules() []Rule {
	rules := getRules()
	ret := make([]Rule, 0, len(rules))
	for _, rule := range rules {
		ret = append(ret, *rule)
	}
	return ret
}

// GetRulesOfResource 返回特定资源规则的copy.
// 如果用户变更规则，并不会影响gray模块
func GetRulesOfResource(res string) []Rule {
	rules := getRuleOfResource(res)
	ret := make([]Rule, 0, len(rules))
	for _, rule := range rules {
		ret = append(ret, *rule)
	}
	return ret
}

// ClearRules 清理Gray模块的全部规则
func ClearRules() error {
	_, err := LoadRules(nil)
	return err
}

// ClearRulesOfResource 清理Gray模块的特定资源规则
func ClearRulesOfResource(res string) error {
	_, err := LoadRulesOfResource(res, nil)
	return err
}

// rulesFrom 从m中获取全部规则
func rulesFrom(m TrafficControllerMap) []*Rule {
	rules := make([]*Rule, 0, 8)
	if len(m) == 0 {
		return rules
	}
	for _, rs := range m {
		if len(rs) == 0 {
			continue
		}
		for _, r := range rs {
			if r != nil && r.BoundRule() != nil {
				rules = append(rules, r.BoundRule())
			}
		}
	}
	return rules
}

// SetTrafficSelectorGenerator 为给定的路由 RouterStrategy 策略设置流量选择器.
// 备注：修改默认的控制策略是部位允许的
func SetTrafficSelectorGenerator(strategy RouterStrategy, generator TrafficControllerGenFunc) error {
	if generator == nil {
		return errors.New("nil generator")
	}
	if strategy >= ConditionRouter && strategy <= WeightRouter {
		return errors.New("not allowed to replace thee generator for default router strategy")
	}
	tcMux.Lock()
	defer tcMux.Unlock()
	tcsGenFuncMap[trafficControllerGenKey{routerStrategy: strategy}] = generator
	return nil
}

// RemoveTrafficSelectorGenerator 移除给定的路由 RouterStrategy 策略设置.
// 备注：修改默认的控制策略是部位允许的
func RemoveTrafficSelectorGenerator(strategy RouterStrategy) error {
	if strategy >= ConditionRouter && strategy <= WeightRouter {
		return errors.New("not allowed to replace thee generator for default router strategy")
	}
	tcMux.Lock()
	defer tcMux.Unlock()
	delete(tcsGenFuncMap, trafficControllerGenKey{
		routerStrategy: strategy,
	})
	return nil
}

// calculateReuseIndexFor 计算规则r的可复用下标
func calculateReuseIndexFor(r *Rule, oldResTsc []*TrafficSelectorController) (equalIdx int) {
	// the index of equivalent rule in old traffic shaping controller slice
	equalIdx = -1

	for idx, oldTc := range oldResTsc {
		oldRule := oldTc.BoundRule()
		if oldRule.isEqualTo(r) {
			// 如果旧规则和r相等，则beak
			equalIdx = idx
			break
		}
	}
	return equalIdx
}

// buildResourceTrafficSelectingController 构建 TrafficSelectorController 规则切片.规则对应的资源必须和res相同
func buildResourceTrafficSelectingController(res string, rulesOfRes []*Rule, oldResTsc []*TrafficSelectorController) []*TrafficSelectorController {
	newTcsOfRes := make([]*TrafficSelectorController, 0, len(rulesOfRes))
	for _, rule := range rulesOfRes {
		if res != rule.Resource {
			logging.Error(errors.Errorf("unmatched resource name expect: %s, actual: %s", res, rule.Resource), "Unmatched resource name in flow.buildResourceTrafficShapingController()", "rule", rule)
			continue
		}
		equalIdx := calculateReuseIndexFor(rule, oldResTsc)
		// 首先检查equals的场景
		if equalIdx >= 0 {
			// 重用旧的tcs
			equalOldTsc := oldResTsc[equalIdx]
			newTcsOfRes = append(newTcsOfRes, equalOldTsc)
			// 从oldResTsc移除旧的tcs
			oldResTsc = append(oldResTsc[:equalIdx], oldResTsc[equalIdx+1:]...)
			continue
		}
		generator, supported := tcsGenFuncMap[trafficControllerGenKey{routerStrategy: rule.RouterStrategy}]
		if !supported || generator == nil {
			logging.Error(errors.New("unsupported gray control strategy"), "Ignoring the rule dule to unsupported control behavior in gray.buildResourceTrafficShapingController()", "rule", rule)
			continue
		}
		tcs, e := generator(rule)
		if tcs == nil || e != nil {
			logging.Error(errors.New("bad generated traffic controller"), "Ignoring the rule due to bad generated traffic controller in gray.buildResourceTrafficShapingController()", "rule", rule)
			continue
		}
		newTcsOfRes = append(newTcsOfRes, tcs)
	}
	return newTcsOfRes
}

// IsValidRule 检查是否为有效的规则
func IsValidRule(rule *Rule) error {
	if rule == nil {
		return errors.New("nil Rule")
	}
	if rule.Resource == "" {
		return errors.New("empty resource")
	}
	if int32(rule.RouterStrategy) < 0 {
		return errors.New("negative RouterStrategy")
	}
	return nil
}

// getTrafficControllerListFor 获取资源关联的流量选择器
func getTrafficControllerListFor(name string) []*TrafficSelectorController {
	tcMux.RLock()
	defer tcMux.RUnlock()
	var tsc = tcsMap[name]
	if tsc != nil {
		return tsc
	}
	var ress []string
	for res := range tcsMap {
		ress = append(ress, res)
	}
	sort.Slice(ress, func(i, j int) bool {
		return ress[i] > ress[j]
	})
	for _, res := range ress {
		if res[len(res)-1] == '*' {
			var length = len(res)
			if length == 1 {
				return tcsMap[res]
			}
			if length > 1 && strings.HasPrefix(name, res[:length-1]) {
				return tcsMap[res]
			}
		}
	}
	return nil
}
