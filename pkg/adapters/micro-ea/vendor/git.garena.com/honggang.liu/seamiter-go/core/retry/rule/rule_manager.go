package rule

import (
	"git.garena.com/honggang.liu/seamiter-go/core/retry/classify"
	"git.garena.com/honggang.liu/seamiter-go/core/retry/support"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/util"
	"github.com/pkg/errors"
	"reflect"
	"sync"
)

// ResourceRetryTemplateMap 资源及对应的重试模板
type resourceRetryTemplateMap map[string]*support.RetryTemplate

var (
	ruleMap       = make(map[string][]*Rule)
	rwMux         = &sync.RWMutex{}
	currentRules  = make(map[string][]*Rule, 0)
	updateRuleMux = new(sync.Mutex)
	rtMap         = make(resourceRetryTemplateMap)
)

// LoadRules loads the given retry rules to the rule manager, while all previous rules will be replaced.
// the first returned value indicates whether you do real load operation, if the rules is the same with previous rules, return false
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
		logging.Info("[Retry] Load rules is the same with current rules, so ignore load operation.")
		return false, nil
	}
	err := onRuleUpdate(resRulesMap)
	return true, err
}

func onRuleUpdate(rawResRulesMap map[string][]*Rule) (err error) {
	validResRulesMap := make(map[string][]*Rule, len(rawResRulesMap))
	for res, rules := range rawResRulesMap {
		validResRules := make([]*Rule, 0, len(rules))
		for _, rule := range rules {
			if err := IsValidRule(rule); err != nil {
				logging.Warn("[Retry onRuleUpdate] Ignoring invalid isolation rule", "rule", rule, "reason", err.Error())
				continue
			}
			validResRules = append(validResRules, rule)
		}
		if len(validResRules) > 0 {
			validResRulesMap[res] = validResRules
		}
		if err := onResourceUpdate(res, validResRules); err != nil {
			logging.Warn("[Retry onResourceUpdate] Ignoring invalid isolation rule", "rules", validResRules, "reason", err.Error())
			continue
		}
	}
	start := util.CurrentTimeNano()
	rwMux.Lock()
	ruleMap = validResRulesMap
	rwMux.Unlock()
	currentRules = rawResRulesMap
	if logging.DebugEnabled() {
		logging.Debug("[Retry onRuleUpdate] Time statistic(ns) for updating isolation rule", "timeCost", util.CurrentTimeNano()-start)
	}
	logRuleUpdate(validResRulesMap)
	return
}

// LoadRulesOfResource loads the given resource's isolation rules to the rule manager, while all previous resource's rules will be replaced.
// the first returned value indicates whether you do real load operation, if the rules is the same with previous resource's rules, return false
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
		// clear ruleMap
		rwMux.Lock()
		delete(ruleMap, res)
		rwMux.Unlock()
		logging.Info("[Retry] clear resource level rules", "resource", res)
		return true, nil
	}
	// load resource level rules
	isEqual := reflect.DeepEqual(currentRules[res], rules)
	if isEqual {
		logging.Info("[Retry] Load resource level rules is the same with current resource level rules, so ignore load operation.")
		return false, nil
	}
	err := onResourceUpdate(res, rules)
	return true, err
}

func onResourceUpdate(res string, rawResRules []*Rule) (err error) {
	validResRules := make([]*Rule, 0, len(rawResRules))
	for _, rule := range rawResRules {
		if err := IsValidRule(rule); err != nil {
			logging.Warn("[Retry onResourceRuleUpdate] Ignoring invalid isolation rule", "rule", rule, "reason", err.Error())
			continue
		}
		validResRules = append(validResRules, rule)
	}
	start := util.CurrentTimeNano()
	rwMux.Lock()
	if len(validResRules) == 0 {
		delete(ruleMap, res)
		delete(rtMap, res)
	} else {
		ruleMap[res] = validResRules
		rtMap[res] = buildResourceRetryTemplate(res, rawResRules)
	}
	rwMux.Unlock()
	currentRules[res] = rawResRules
	if logging.DebugEnabled() {
		logging.Debug("[Retry onResourceRuleUpdate] Time statistic(ns) for updating isolation rule", "timeCost", util.CurrentTimeNano()-start)
	}
	logging.Info("[Retry] load resource level rules", "resource", res, "validResRules", validResRules)
	return nil
}

// ClearRules clears all the rules in isolation module.
func ClearRules() error {
	_, err := LoadRules(nil)
	return err
}

// GetRules returns all the rules based on copy.
// It doesn't take effect for isolation module if user changes the rule.
func GetRules() []Rule {
	rules := getRules()
	ret := make([]Rule, 0, len(rules))
	for _, rule := range rules {
		ret = append(ret, *rule)
	}
	return ret
}

// GetRulesOfResource returns specific resource's rules based on copy.
// It doesn't take effect for isolation module if user changes the rule.
func GetRulesOfResource(res string) []Rule {
	rules := getRulesOfResource(res)
	ret := make([]Rule, 0, len(rules))
	for _, rule := range rules {
		ret = append(ret, *rule)
	}
	return ret
}

func GetRetryTemplateOfResource(res string) *support.RetryTemplate {
	var resTemplates = getRetryTemplatesOfResource(res)
	if resTemplates == nil {
		return nil
	}
	return resTemplates.DeepCopy(resTemplates)
}

// getRulesOfResource returns specific resource's rules。Any changes of rules take effect for isolation module
// getRulesOfResource is an internal interface.
func getRetryTemplatesOfResource(res string) *support.RetryTemplate {
	rwMux.RLock()
	defer rwMux.RUnlock()

	resTemplates, exist := rtMap[res]
	if !exist {
		return nil
	}
	return resTemplates
}

// getRules returns all the rules。Any changes of rules take effect for isolation module
// getRules is an internal interface.
func getRules() []*Rule {
	rwMux.RLock()
	defer rwMux.RUnlock()
	return rulesFrom(ruleMap)
}

// getRulesOfResource returns specific resource's rules。Any changes of rules take effect for isolation module
// getRulesOfResource is an internal interface.
func getRulesOfResource(res string) []*Rule {
	rwMux.RLock()
	defer rwMux.RUnlock()

	resRules, exist := ruleMap[res]
	if !exist {
		return nil
	}
	ret := make([]*Rule, 0, len(resRules))
	for _, r := range resRules {
		ret = append(ret, r)
	}
	return ret
}

func rulesFrom(m map[string][]*Rule) []*Rule {
	rules := make([]*Rule, 0, 8)
	if len(m) == 0 {
		return rules
	}
	for _, rs := range m {
		for _, r := range rs {
			if r != nil {
				rules = append(rules, r)
			}
		}
	}
	return rules
}

func logRuleUpdate(m map[string][]*Rule) {
	rs := rulesFrom(m)
	if len(rs) == 0 {
		logging.Info("[RetryRuleManager] Retry rules were cleared")
	} else {
		logging.Info("[RetryRuleManager] Retry rules were loaded", "rules", rs)
	}
}

// IsValidRule checks whether the given Rule is valid.
func IsValidRule(r *Rule) error {
	if r == nil {
		return errors.New("nil isolation rule")
	}
	if len(r.Resource) == 0 {
		return errors.New("empty resource of isolation rule")
	}
	return nil
}

func buildResourceRetryTemplate(res string, resRule []*Rule) *support.RetryTemplate {
	if len(resRule) == 0 {
		return nil
	}
	var rule = resRule[0]
	if res != rule.Resource {
		logging.Error(errors.Errorf("unmatched resource name, expect: %s, actual: %s", res, rule.Resource), "Unmatched resource name in retry.buildResourceRetryTemplate()", "rule", rule)
		return nil
	}
	var retryTemplateBuilder = support.NewRetryTemplateBuilder()
	// 设置重试策略
	if rule.RetryPolicy < NeverRetryPolicy || rule.RetryPolicy > CustomPolicyRtyPolicy {
		logging.Error(errors.Errorf("no matched  retry policy,  actual policy: %s", rule.RetryPolicy), "no matched retry policy name in retry.buildResourceRetryTemplate()", "rule", rule)
		return nil
	}
	if rule.RetryPolicy == NeverRetryPolicy {
		retryTemplateBuilder = retryTemplateBuilder.NeverRtyPolicy()
	} else if rule.RetryPolicy == SimpleRetryPolicy {
		var errs []error
		for _, e := range rule.ExcludeExceptions {
			errs = append(errs, errors.New(e))
		}
		retryTemplateBuilder = retryTemplateBuilder.NewSimpleRetryPolicyWithMaxAttemptsAndErrors(rule.RetryMaxAttempts, errs)
	} else if rule.RetryPolicy == TimeoutRtyPolicy {
		retryTemplateBuilder = retryTemplateBuilder.WithinMillisRtyPolicy(rule.RetryTimeout)
	} else if rule.RetryPolicy == MaxAttemptsRetryPolicy {
		retryTemplateBuilder = retryTemplateBuilder.MaxAttemptsRtyPolicy(rule.RetryMaxAttempts)
	} else if rule.RetryPolicy == ErrorClassifierRetryPolicy {
		// TODO 暂未实现
	} else if rule.RetryPolicy == AlwaysRetryPolicy {
		retryTemplateBuilder = retryTemplateBuilder.InfiniteRtyPolicy()
	} else if rule.RetryPolicy == CompositeRetryPolicy {
		// TODO 暂未实现
	} else if rule.RetryPolicy == CustomPolicyRtyPolicy {
		// TODO 暂未实现
	}

	// 设置回退策略
	if rule.BackoffPolicy < NoBackOffPolicy || rule.BackoffPolicy > UniformRandomBackoffPolicy {
		logging.Error(errors.Errorf("no matched  backoff policy,  actual backoff policy: %s", rule.BackoffPolicy), "no matched backoff policy name in retry.buildResourceRetryTemplate()", "rule", rule)
		return nil
	}
	if rule.BackoffPolicy == NoBackOffPolicy {
		retryTemplateBuilder = retryTemplateBuilder.NoBackoff()
	} else if rule.BackoffPolicy == FixedBackOffPolicy {
		retryTemplateBuilder = retryTemplateBuilder.FixedBackoff(rule.FixedBackOffPeriodInMs)
	} else if rule.BackoffPolicy == ExponentialBackOffPolicy {
		retryTemplateBuilder = retryTemplateBuilder.ExponentialBackoff(rule.BackoffDelay, rule.BackoffMultiplier, rule.BackoffMaxDelay)
	} else if rule.BackoffPolicy == ExponentialRandomBackOffPolicy {
		retryTemplateBuilder = retryTemplateBuilder.ExponentialBackoffWithRandom(rule.BackoffDelay, rule.BackoffMultiplier, rule.BackoffMaxDelay, true)
	} else if rule.BackoffPolicy == UniformRandomBackoffPolicy {
		retryTemplateBuilder = retryTemplateBuilder.UniformRandomBackoff(rule.UniformMinBackoffPeriod, rule.UniformMaxBackoffPeriod)
	}

	// 设置异常匹配模式
	retryTemplateBuilder = retryTemplateBuilder.WithErrorMatchPattern(classify.PatternMatcher(rule.ErrorMatcher))
	// 设置需要重试的异常及排除的异常
	if len(rule.IncludeExceptions) != 0 && len(rule.ExcludeExceptions) != 0 {
		// 异常类型只能包含一种，要么包含，要么排除，不可以同时成立
		logging.Error(errors.Errorf("only can set one mode,include or except"), "only can set one mode,include or except in retry.buildResourceRetryTemplate()", "rule", rule)
		return nil
	}
	var retryOnErrors []error
	for _, e := range rule.IncludeExceptions {
		retryOnErrors = append(retryOnErrors, errors.New(e))
	}
	retryTemplateBuilder = retryTemplateBuilder.RetryOnErrors(retryOnErrors)
	var excludeExceptions []error
	for _, e := range rule.ExcludeExceptions {
		excludeExceptions = append(excludeExceptions, errors.New(e))
	}
	retryTemplateBuilder = retryTemplateBuilder.NotRetryOnErrors(excludeExceptions)
	return retryTemplateBuilder.Build()
}
