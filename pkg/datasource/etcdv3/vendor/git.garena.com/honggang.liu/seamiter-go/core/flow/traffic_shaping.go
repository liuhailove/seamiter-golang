package flow

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	metric_exporter "git.garena.com/honggang.liu/seamiter-go/exporter/metric"
)

var (
	resourceFlowThresholdGauge = metric_exporter.NewGauge(
		"resource_flow_threshold",
		"Resource flow threshold",
		[]string{"resource"})
)

func init() {
	metric_exporter.Register(resourceFlowThresholdGauge)
}

// TrafficShapingCalculator calculates the actual traffic shaping threshold
// based on the threshold of rule and the traffic shaping strategy.
type TrafficShapingCalculator interface {
	BoundOwner() *TrafficShapingController
	CalculateAllowedTokens(batchCount uint32, flag int32) float64
}

// standaloneStatistic indicates the independent statistic for each TrafficShapingController
type standaloneStatistic struct {
	// reuseResourceStat indicates whether current standaloneStatistic reuse the current resource's global statistic
	reuseResourceStat bool
	// readOnlyMetric is the readonly metric statistic.
	// if reuseResourceStat is true, it would be the reused SlidingWindowMetric
	// if reuseResourceStat is false, it would be the BucketLeapArray
	readOnlyMetric base.ReadStat
	// writeOnlyMetric is to write only metric statistic.
	// if reuseResourceStat is true, it would be nil
	// if reuseResourceStat is false, it would be the BucketLeapArray
	writeOnlyMetric base.WriteStat
}

// TrafficShapingChecker performs checking according to current metrics and the traffic
// shaping strategy, then yield the token result.
type TrafficShapingChecker interface {
	BoundOwner() *TrafficShapingController
	DoCheck(resStat base.StatNode, batchCount uint32, threshold float64) *base.TokenResult
}
type TrafficShapingController struct {
	flowCalculator TrafficShapingCalculator
	flowChecker    TrafficShapingChecker

	rule *Rule

	// boundStat is the statistic of current TrafficShapingController
	boundStat standaloneStatistic
}

func NewTrafficShapingController(rule *Rule, boundStat *standaloneStatistic) (*TrafficShapingController, error) {
	return &TrafficShapingController{
		rule:      rule,
		boundStat: *boundStat,
	}, nil
}

func (t *TrafficShapingController) BoundRule() *Rule {
	return t.rule
}

func (t *TrafficShapingController) FlowChecker() TrafficShapingChecker {
	return t.flowChecker
}

func (t *TrafficShapingController) FlowCalculator() TrafficShapingCalculator {
	return t.flowCalculator
}

func (t *TrafficShapingController) PerformChecking(resStat base.StatNode, batchCount uint32, flag int32) *base.TokenResult {
	allowedToken := t.flowCalculator.CalculateAllowedTokens(batchCount, flag)
	resourceFlowThresholdGauge.Set(allowedToken, t.rule.Resource)
	return t.flowChecker.DoCheck(resStat, batchCount, allowedToken)
}
