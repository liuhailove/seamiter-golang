package datasource

import "git.garena.com/honggang.liu/seamiter-go/core/system"

type SystemRule struct {
	// ID represents the unique ID of the rule (optional).
	ID string `json:"id,omitempty"`

	// LimitApp
	// Application name that will be limited by origin.
	// The default limitApp is {@code default}, which means allowing all origin apps.
	// For authority rules, multiple origin name can be separated with comma (',').
	LimitApp string `json:"limitApp"`
	//  HighestSystemLoad negative value means no threshold checking
	//  The load is not same as Linux system load, which is not sensitive enough. To calculate the load, both Linux system load, current global response time and global QPS will be considered, which means that we need to coordinate with setAvgRt(long) and setQps(double)
	//  Note that this parameter is only available on Unix like system.
	HighestSystemLoad float64 `json:"highestSystemLoad"`
	// HighestCpuUsage cpu usage, between [0, 1]
	HighestCpuUsage float64 `json:"highestCpuUsage"`
	// Qps In a high concurrency condition, real passed QPS may be greater than max QPS set. The real passed QPS will nearly satisfy the following formula:
	// real passed QPS = QPS set + concurrent thread number
	Qps float64 `json:"qps"`
	// AvgRt ax average RT(response time) of all passed requests
	AvgRt float64 `json:"avgRt"`
	//  MaxThread When concurrent thread number is greater than maxThread only maxThread will run in parallel.
	MaxThread float64 `json:"maxThread"`
}

// transToSystemRule 把系统规则转换为Java可以识别的系统规则
func transToSystemRule(rules []system.Rule) []*SystemRule {
	if rules == nil || len(rules) == 0 {
		return nil
	}
	var ruleArr = make([]*SystemRule, 0)
	for _, rule := range rules {
		var sysRule = new(SystemRule)
		sysRule.ID = rule.ID
		switch rule.MetricType {
		case system.Load:
			sysRule.HighestSystemLoad = rule.TriggerCount
			sysRule.HighestCpuUsage = -1.0
			sysRule.Qps = -1.0
			sysRule.AvgRt = -1.0
			sysRule.MaxThread = -1.0
			ruleArr = append(ruleArr, sysRule)
		case system.AvgRT:
			sysRule.AvgRt = rule.TriggerCount
			sysRule.HighestCpuUsage = -1.0
			sysRule.Qps = -1.0
			sysRule.HighestSystemLoad = -1.0
			sysRule.MaxThread = -1.0
			ruleArr = append(ruleArr, sysRule)
		case system.Concurrency:
			sysRule.MaxThread = rule.TriggerCount
			sysRule.HighestCpuUsage = -1.0
			sysRule.Qps = -1.0
			sysRule.HighestSystemLoad = -1.0
			sysRule.AvgRt = -1.0
			ruleArr = append(ruleArr, sysRule)
		case system.InboundQPS:
			sysRule.Qps = rule.TriggerCount
			sysRule.HighestCpuUsage = -1.0
			sysRule.HighestSystemLoad = -1.0
			sysRule.AvgRt = -1.0
			sysRule.MaxThread = -1.0
			ruleArr = append(ruleArr, sysRule)
		case system.CpuUsage:
			sysRule.HighestCpuUsage = rule.TriggerCount
			sysRule.Qps = -1.0
			sysRule.HighestSystemLoad = -1.0
			sysRule.AvgRt = -1.0
			sysRule.MaxThread = -1.0
			ruleArr = append(ruleArr, sysRule)
		case system.MetricTypeSize:
		default:
		}
	}
	return ruleArr
}
