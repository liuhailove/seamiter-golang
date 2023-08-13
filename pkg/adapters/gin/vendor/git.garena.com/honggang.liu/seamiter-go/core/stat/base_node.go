package stat

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/core/config"
	base2 "git.garena.com/honggang.liu/seamiter-go/core/stat/base"
	"sync/atomic"
)

type BaseStatNode struct {
	sampleCount uint32
	intervalMs  uint32

	concurrency int32

	arr    *base2.BucketLeapArray
	metric *base2.SlidingWindowMetric

	//arrInMinute    *base2.BucketLeapArray     // 分钟桶
	//metricInMinute *base2.SlidingWindowMetric // 分钟度量
}

func NewBaseStatNode(sampleCount uint32, intervalInMs uint32) *BaseStatNode {
	la := base2.NewBucketLeapArray(config.GlobalStatisticSampleCountTotal(), config.GlobalStatisticIntervalMsTotal())
	metric, _ := base2.NewSlidingWindowMetric(sampleCount, intervalInMs, la)
	//
	//// 分钟统计
	//arrInMinute := base2.NewBucketLeapArray(60, 60*1000)
	//metricInMinute, _ := base2.NewSlidingWindowMetric(60, 60*1000, arrInMinute)
	return &BaseStatNode{
		sampleCount: sampleCount,
		intervalMs:  intervalInMs,
		concurrency: 0,
		arr:         la,
		metric:      metric,
		//arrInMinute:    arrInMinute,
		//metricInMinute: metricInMinute,
	}
}

func (n *BaseStatNode) MetricsOnCondition(predicate base.TimePredicate) []*base.MetricItem {
	return n.metric.SecondMetricsOnCondition(predicate)
}

func (n *BaseStatNode) GetQPS(event base.MetricEvent) float64 {
	return n.metric.GetQPS(event)
}
func (n *BaseStatNode) GetPreviousQPS(event base.MetricEvent) float64 {
	return n.metric.GetPreviousQPS(event)
}

func (n *BaseStatNode) GetSum(event base.MetricEvent) int64 {
	return n.metric.GetSum(event)
}

func (n *BaseStatNode) GetMaxAvg(event base.MetricEvent) float64 {
	return float64(n.metric.GetMaxOfSingleBucket(event)) * float64(n.sampleCount) / float64(n.intervalMs) * 1000.0
}

func (n *BaseStatNode) AddCount(event base.MetricEvent, count int64) {
	n.arr.AddCount(event, count)
	//n.arrInMinute.AddCount(event, count)
}

func (n *BaseStatNode) UpdateConcurrency(concurrency int32) {
	n.arr.UpdateConcurrency(concurrency)
	//n.arrInMinute.UpdateConcurrency(concurrency)
}
func (n *BaseStatNode) AvgRT() float64 {
	complete := n.metric.GetSum(base.MetricEventComplete)
	if complete <= 0 {
		return 0.0
	}
	return float64(n.metric.GetSum(base.MetricEventRt) / complete)
}
func (n *BaseStatNode) MinRT() float64 {
	return n.metric.MinRT()
}

func (n *BaseStatNode) MaxConcurrency() int32 {
	return n.metric.MaxConcurrency()
}
func (n *BaseStatNode) CurrentConcurrency() int32 {
	return atomic.LoadInt32(&(n.concurrency))
}

func (n *BaseStatNode) IncreaseConcurrency() {
	n.UpdateConcurrency(atomic.AddInt32(&(n.concurrency), 1))
}

func (n *BaseStatNode) DecreaseConcurrency() {
	atomic.AddInt32(&(n.concurrency), -1)
}
func (n *BaseStatNode) GenerateReadStat(sampleCount uint32, intervalInMs uint32) (base.ReadStat, error) {
	return base2.NewSlidingWindowMetric(sampleCount, intervalInMs, n.arr)
}

func (n *BaseStatNode) DefaultMetric() base.ReadStat {
	return n.metric
}
