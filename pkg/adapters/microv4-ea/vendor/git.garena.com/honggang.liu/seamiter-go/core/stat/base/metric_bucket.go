package base

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"github.com/pkg/errors"
	"sync/atomic"
)

// MetricBucket represents the entity to record metrics per minimum time unit (i.e. the bucket time span).
// Note that all operations of the MetricBucket are required to be thread-safe.
type MetricBucket struct {
	// Value of statistic
	counter        [base.MetricEventTotal]int64
	minRt          int64
	maxConcurrency int32
}

func NewMetricBucket() *MetricBucket {
	mb := &MetricBucket{
		minRt:          base.DefaultStatisticMaxRt,
		maxConcurrency: 0,
	}
	return mb
}

// Add statistic count for the given metric event.
func (mb *MetricBucket) Add(event base.MetricEvent, count int64) {
	if event >= base.MetricEventTotal || event < 0 {
		logging.Error(errors.Errorf("Unknown metric event: %v", event), "")
		return
	}
	if event == base.MetricEventRt {
		mb.AddRt(count)
		return
	}
	mb.addCount(event, count)
}

func (mb *MetricBucket) addCount(event base.MetricEvent, count int64) {
	atomic.AddInt64(&mb.counter[event], count)
}

// Get current statistic count of the given metric event.
func (mb *MetricBucket) Get(event base.MetricEvent) int64 {
	if event >= base.MetricEventTotal || event < 0 {
		logging.Error(errors.Errorf("Unknown metric event: %v", event), "")
		return 0
	}
	return atomic.LoadInt64(&mb.counter[event])
}

func (mb *MetricBucket) reset() {
	for i := 0; i < int(base.MetricEventTotal); i++ {
		atomic.StoreInt64(&mb.counter[i], 0)
	}
	atomic.StoreInt64(&mb.minRt, base.DefaultStatisticMaxRt)
	atomic.StoreInt32(&mb.maxConcurrency, int32(0))
}

func (mb *MetricBucket) AddRt(rt int64) {
	mb.addCount(base.MetricEventRt, rt)
	if rt < atomic.LoadInt64(&mb.minRt) {
		// Might not be accurate here.
		atomic.StoreInt64(&mb.minRt, rt)
	}
}

func (mb *MetricBucket) MinRt() int64 {
	return atomic.LoadInt64(&mb.minRt)
}

func (mb *MetricBucket) UpdateConcurrency(concurrency int32) {
	cc := concurrency
	if cc > atomic.LoadInt32(&mb.maxConcurrency) {
		// Might not be accurate here.
		atomic.StoreInt32(&mb.maxConcurrency, cc)
	}
}

func (mb *MetricBucket) MaxConcurrency() int32 {
	return atomic.LoadInt32(&mb.maxConcurrency)
}
