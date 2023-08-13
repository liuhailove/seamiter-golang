package mock

import (
	"fmt"
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/util"
	"github.com/fatih/structs"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"reflect"
	"sync"
)

// TrafficControllerGenFunc represents the TrafficShapingController generator function of a specific control behavior.
type TrafficControllerGenFunc func(r *Rule) TrafficShapingController

// trafficControllerMap represents the map storage for TrafficShapingController.
type trafficControllerMap map[string][]TrafficShapingController

var (
	tcGenFuncMap  = make(map[ControlBehavior]TrafficControllerGenFunc, 4)
	tcMap         = make(trafficControllerMap)
	tcMux         = new(sync.RWMutex)
	currentRules  = make(map[string][]*Rule, 0)
	updateRuleMux = new(sync.Mutex)

	// 暂存请求
	cacheRequestMap = sync.Map{}
	// 请求hash值
	requestHashMap = sync.Map{}
	// 已经上报的request
	reportedRequestHashMap = sync.Map{}
	jsonHold               = jsoniter.ConfigCompatibleWithStandardLibrary
)

const (
	// OnceMaxReportLimit 一次最多上报200个请求
	OnceMaxReportLimit int = 200
)

func init() {
	// Initialize the traffic shaping controller generator map for existing control behaviors.
	tcGenFuncMap[DoNothing] = func(r *Rule) TrafficShapingController {
		tsc := newBaseTrafficShapingController(r)
		return &defaultTrafficShapingController{*tsc}
	}
	tcGenFuncMap[Panic] = func(r *Rule) TrafficShapingController {
		tsc := newBaseTrafficShapingController(r)
		return &panicTrafficShapingController{*tsc}
	}
	tcGenFuncMap[Mock] = func(r *Rule) TrafficShapingController {
		tsc := newBaseTrafficShapingController(r)
		return &mockTrafficShapingController{*tsc}
	}
	tcGenFuncMap[Waiting] = func(r *Rule) TrafficShapingController {
		tsc := newBaseTrafficShapingController(r)
		return &waitingTrafficShapingController{*tsc}
	}
	tcGenFuncMap[WaitingThenPanic] = func(r *Rule) TrafficShapingController {
		tsc := newBaseTrafficShapingController(r)
		return &waitingThenPanicTrafficShapingController{*tsc}
	}
	tcGenFuncMap[WaitingThenMock] = func(r *Rule) TrafficShapingController {
		tsc := newBaseTrafficShapingController(r)
		return &waitingThenMockTrafficShapingController{*tsc}
	}
}

func getTrafficControllersFor(res string) []TrafficShapingController {
	tcMux.RLock()
	defer tcMux.RUnlock()
	return tcMap[res]
}

// LoadRules replaces all mock rules with the given rules.
// Return value:
//   bool: indicates whether the internal map has been changed;
//   error: indicates whether occurs the error.
func LoadRules(rules []*Rule) (bool, error) {
	resRulesMap := make(map[string][]*Rule, 16)
	for _, rule := range rules {
		resRules, exists := resRulesMap[rule.Resource]
		if !exists {
			resRules = make([]*Rule, 0, 1)
		}
		resRulesMap[rule.Resource] = append(resRules, rule)
	}
	updateRuleMux.Lock()
	defer updateRuleMux.Unlock()
	isEqual := reflect.DeepEqual(currentRules, resRulesMap)
	if isEqual {
		logging.Info("[Mock] Load rules is the same with current rules, so ignore load operation.")
		return false, nil
	}
	err := onRuleUpdate(resRulesMap)
	return true, err
}

// GetRules returns all the mock rules based on copy.
// It doesn't take effect for mock module if user changes the returned rules.
// GetRules need to compete mock module's global lock and the high performance losses of copy,
// reduce or do not call GetRules if possible.
func GetRules() []Rule {
	tcMux.RLock()
	rules := rulesFrom(tcMap)
	tcMux.RUnlock()

	ret := make([]Rule, 0, len(rules))
	for _, rule := range rules {
		ret = append(ret, *rule)
	}
	return ret
}

// GetRulesOfResource returns specific resource's mock rules based on copy.
// It doesn't take effect for hotspot module if user changes the returned rules.
// GetRulesOfResource need to compete mock module's global lock and the high performance losses of copy,
// reduce or do not call GetRulesOfResource frequently if possible.
func GetRulesOfResource(res string) []Rule {
	tcMux.RLock()
	resTcs := tcMap[res]
	tcMux.RUnlock()

	ret := make([]Rule, 0, len(resTcs))
	for _, tc := range resTcs {
		ret = append(ret, *tc.BoundRule())
	}
	return ret
}

// ClearRules clears all hotspot param flow rules.
func ClearRules() error {
	_, err := LoadRules(nil)
	return err
}

func onRuleUpdate(rawResRulesMap map[string][]*Rule) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%+v", r)
			}
		}
	}()

	// ignore invalid rules
	validResRulesMap := make(map[string][]*Rule, len(rawResRulesMap))
	for res, rules := range rawResRulesMap {
		validResRules := make([]*Rule, 0, len(rules))
		for _, rule := range rules {
			if err := IsValidRule(rule); err != nil {
				logging.Warn("[Mock onRuleUpdate] Ignoring invalid mock rule when loading new rules", "rule", rule, "err", err.Error())
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
	tcMapClone := make(trafficControllerMap, len(tcMap))
	for res, tcs := range tcMap {
		restTcClone := make([]TrafficShapingController, 0, len(tcs))
		restTcClone = append(restTcClone, tcs...)
		tcMapClone[res] = restTcClone
	}
	tcMux.RUnlock()

	m := make(trafficControllerMap, len(validResRulesMap))
	for res, rules := range validResRulesMap {
		m[res] = buildResourceTrafficShapingController(res, rules, tcMapClone[res])
	}

	tcMux.Lock()
	tcMap = m
	tcMux.Unlock()

	currentRules = rawResRulesMap
	if logging.DebugEnabled() {
		logging.Debug("[Mock onRuleUpdate] Time statistic(ns) for updating hotspot param flow rules", "timeCost", util.CurrentTimeNano()-start)
	}
	logRuleUpdate(validResRulesMap)
	return nil
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
			logging.Warn("[Mock onResourceRuleUpdate] Ignoring invalid hotspot param flow rule", "rule", rule, "reason", err.Error())
			continue
		}
		validResRules = append(validResRules, rule)
	}
	start := util.CurrentTimeNano()
	oldResTcs := make([]TrafficShapingController, 0, 8)
	tcMux.RLock()
	oldResTcs = append(oldResTcs, tcMap[res]...)
	tcMux.RUnlock()

	newResTcs := buildResourceTrafficShapingController(res, validResRules, oldResTcs)

	tcMux.Lock()
	if len(newResTcs) == 0 {
		delete(tcMap, res)
	} else {
		tcMap[res] = newResTcs
	}
	tcMux.Unlock()

	currentRules[res] = rawResRules
	logging.Debug("[Mock onResourceRuleUpdate] Time statistic(ns) for updating hotspot param flow rules", "timeCost", util.CurrentTimeNano()-start)
	logging.Info("[Mock] load resource level hotspot param flow rules", "resource", res, "validResRules", validResRules)
	return nil
}

// LoadRulesOfResource loads the given resource's mock rules to the rule manager,
// while all previous resource's rules will be replaced. The first returned value indicates whether you
// do real load operation, if the rules is the same with previous resource's rules, return false.
func LoadRulesOfResource(res string, rules []*Rule) (bool, error) {
	if len(res) == 0 {
		return false, errors.New("empty resource")
	}

	updateRuleMux.Lock()
	defer updateRuleMux.Unlock()

	// clear resource rules
	if len(rules) == 0 {
		// clear resource's currentRules
		delete(currentRules, res)
		// clear tcMap
		tcMux.Lock()
		delete(tcMap, res)
		tcMux.Unlock()
		logging.Info("[Mock] clear resource level hotspot param flow rules", "resource", res)
		return true, nil
	}

	// load resource level rules
	isEqual := reflect.DeepEqual(currentRules[res], rules)
	if isEqual {
		logging.Info("[Mock] Load resource level hotspot param flow rules is the same with current resource level rules, so ignore load operation.")
		return false, nil
	}

	err := onResourceRuleUpdate(res, rules)
	return true, err
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
		logging.Info("[MockRuleManager] Hotspot param flow rules were cleared")
	} else {
		logging.Info("[MockRuleManager] Hotspot param flow rules were loaded", "rules", rules)
	}
}

func rulesFrom(m trafficControllerMap) []*Rule {
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

func calculateReuseIndexFor(r *Rule, oldResTcs []TrafficShapingController) (equalIdx, reuseStatIdx int) {
	// the index of equivalent rule in old traffic shaping controller slice
	equalIdx = -1
	// the index of statistic reusable rule in old traffic shaping controller slice
	reuseStatIdx = -1

	for idx, oldTc := range oldResTcs {
		oldRule := oldTc.BoundRule()
		if oldRule.isEqualTo(r) {
			// break if there is equivalent rule
			equalIdx = idx
			break
		}
		// find the index of first StatReusable rule
		if !oldRule.isStatReusable(r) {
			continue
		}
		if reuseStatIdx >= 0 {
			// had found reuse rule
			continue
		}
		reuseStatIdx = idx
	}
	return equalIdx, reuseStatIdx
}

// buildResourceTrafficShapingController builds TrafficShapingController slice from rules. the resource of rules must be equals to res.
func buildResourceTrafficShapingController(res string, resRules []*Rule, oldResTcs []TrafficShapingController) []TrafficShapingController {
	newTcsOfRes := make([]TrafficShapingController, 0, len(resRules))
	for _, rule := range resRules {
		if res != rule.Resource {
			logging.Error(errors.Errorf("unmatched resource name, expect: %s, actual: %s", res, rule.Resource), "Unmatched resource name in hotspot.buildResourceTrafficShapingController()", "rule", rule)
			continue
		}
		equalIdx, reuseStatIdx := calculateReuseIndexFor(rule, oldResTcs)
		// there is equivalent rule in old traffic shaping controller slice
		if equalIdx >= 0 {
			equalOldTc := oldResTcs[equalIdx]
			newTcsOfRes = append(newTcsOfRes, equalOldTc)
			// remove old tc from old resTcs
			oldResTcs = append(oldResTcs[:equalIdx], oldResTcs[equalIdx+1:]...)
			continue
		}
		// generate new traffic shaping controller
		generator, supported := tcGenFuncMap[rule.ControlBehavior]
		if !supported {
			logging.Warn("[HotSpot buildResourceTrafficShapingController] Ignoring the hotspot param flow rule due to unsupported control behavior", "rule", rule)
			continue
		}
		var tc TrafficShapingController
		if reuseStatIdx >= 0 {
			// generate new traffic shaping controller with reusable statistic metric.
			tc = generator(rule)
			// remove the reused traffic shaping controller old res tcs
			oldResTcs = append(oldResTcs[:reuseStatIdx], oldResTcs[reuseStatIdx+1:]...)
		} else {
			tc = generator(rule)
		}
		if tc == nil {
			logging.Debug("[HotSpot buildResourceTrafficShapingController] Ignoring the hotspot param flow rule due to bad generated traffic controller", "rule", rule)
			continue
		}
		newTcsOfRes = append(newTcsOfRes, tc)
	}
	return newTcsOfRes
}

// ClearRulesOfResource clears resource level hotspot param flow rules.
func ClearRulesOfResource(res string) error {
	_, err := LoadRulesOfResource(res, nil)
	return err
}

func IsValidRule(r *Rule) error {
	if r == nil {
		return errors.New("nil Rule")
	}
	if len(r.Resource) == 0 {
		return errors.New("empty resource name")
	}
	return nil
}

// cacheRequest 请求缓存
func cacheRequest(rule *Rule, ctx *base.EntryContext) {
	if rule == nil || !rule.RequestHold {
		return
	}
	// go-micro处理
	if ctx.Resource.Classification() == base.ResTypeMicro && structs.IsStruct(ctx.Input.Args[0]) {
		if requestJsonData, err := jsonHold.Marshal(ctx.Input.Args[0]); err == nil {
			requestStr := string(requestJsonData)
			// 如果同样地请求已经存储，则不再存储
			_, exist := requestHashMap.Load(util.String(requestStr))
			if exist {
				return
			}
			requestHashMap.Store(util.String(requestStr), nil)
			// 缓存请求
			var requests, _ = cacheRequestMap.LoadOrStore(ctx.Resource.Name(), []string{})
			var reqs = requests.([]string)
			reqs = append(reqs, requestStr)
			if len(reqs) > OneResourceLimit {
				// 剔除最前面数据
				reqs = reqs[1:OneResourceLimit]
			}
			cacheRequestMap.Store(ctx.Resource.Name(), reqs)
		}
	}
	// gin-web
	if ctx.Resource.Classification() == base.ResTypeWeb {
		var args = ctx.Input.Args
		var params string
		for _, arg := range args {
			params += arg.(string) + ";"
		}
		// 如果同样地请求已经存储，则不再存储
		_, exist := requestHashMap.Load(util.String(params))
		if exist {
			return
		}
		requestHashMap.Store(util.String(params), nil)
		// 缓存请求
		var requests, _ = cacheRequestMap.LoadOrStore(ctx.Resource.Name(), []string{})
		var reqs = requests.([]string)
		reqs = append(reqs, params)
		if len(reqs) > OneResourceLimit {
			// 剔除最前面数据
			reqs = reqs[1:OneResourceLimit]
		}
		cacheRequestMap.Store(ctx.Resource.Name(), reqs)
	}
}

// NeedReportRequest 需要上报的请求
func NeedReportRequest() map[string][]string {
	var reportMap = make(map[string][]string, 0)
	// 计数，避免每次上报数据过大
	var count = 0
	cacheRequestMap.Range(func(resourceName, requests interface{}) bool {
		if count < OnceMaxReportLimit {
			var rreqs = reportMap[resourceName.(string)]
			if rreqs == nil {
				rreqs = []string{}
			}
			var reqs = requests.([]string)
			for _, req := range reqs {
				_, exist := reportedRequestHashMap.LoadOrStore(util.String(req), util.String(req))
				if !exist {
					rreqs = append(rreqs, req)
				}
			}
			if len(rreqs) > 0 {
				reportMap[resourceName.(string)] = rreqs
			}
			count++
		}
		return true
	})
	return reportMap
}
