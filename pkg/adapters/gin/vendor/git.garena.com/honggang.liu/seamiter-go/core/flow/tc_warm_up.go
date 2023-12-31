package flow

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/core/config"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/util"
	"math"
	"sync/atomic"
)

type WarmUpTrafficShapingCalculator struct {
	owner             *TrafficShapingController
	threshold         float64
	warmUpPeriodInSec uint32
	coldFactor        uint32
	warningToken      uint64
	maxToken          uint64
	slope             float64
	storedTokens      int64
	lastFilledTime    uint64
}

func (c *WarmUpTrafficShapingCalculator) BoundOwner() *TrafficShapingController {
	return c.owner
}
func NewWarmUpTrafficShapingCalculator(owner *TrafficShapingController, rule *Rule) TrafficShapingCalculator {
	if rule.WarmUpColdFactor <= 1 {
		rule.WarmUpColdFactor = config.DefaultWarmUpColdFactor
		logging.Warn("[NewWarmUpTrafficShapingCalculator] No set WarmUpColdFactor,use default warm up cold factor value", "defaultWarmUpColdFactor", config.DefaultWarmUpColdFactor)
	}
	warningToken := uint64((float64(rule.WarmUpPeriodSec) * rule.Threshold) / float64(rule.WarmUpColdFactor-1))

	maxToken := warningToken + uint64(2*float64(rule.WarmUpPeriodSec)*rule.Threshold/float64(1.0+rule.WarmUpColdFactor))

	slope := float64(rule.WarmUpColdFactor-1.0) / rule.Threshold / float64(maxToken-warningToken)

	warmUpTrafficShapingCalculator := &WarmUpTrafficShapingCalculator{
		owner:             owner,
		threshold:         rule.Threshold,
		warmUpPeriodInSec: rule.WarmUpPeriodSec,
		coldFactor:        rule.WarmUpColdFactor,
		warningToken:      warningToken,
		maxToken:          maxToken,
		slope:             slope,
		storedTokens:      0,
		lastFilledTime:    0,
	}
	return warmUpTrafficShapingCalculator
}

func (c *WarmUpTrafficShapingCalculator) CalculateAllowedTokens(_ uint32, _ int32) float64 {
	metricReadonlyStat := c.BoundOwner().boundStat.readOnlyMetric
	previousQps := metricReadonlyStat.GetPreviousQPS(base.MetricEventPass)
	c.syncToken(previousQps)

	restToken := atomic.LoadInt64(&c.storedTokens)
	if restToken <= 0 {
		restToken = 0
	}
	if restToken >= int64(c.warningToken) {
		aboveToken := restToken - int64(c.warningToken)
		warningQps := math.Nextafter(1.0/(float64(aboveToken)*c.slope+1.0/c.threshold), math.MaxFloat64)
		return warningQps
	}
	return c.threshold
}

func (c *WarmUpTrafficShapingCalculator) syncToken(passQps float64) {
	currentTime := util.CurrentTimeMillis()
	currentTime = currentTime - currentTime%1000
	oldLastFillTime := atomic.LoadUint64(&c.lastFilledTime)
	if currentTime <= oldLastFillTime {
		return
	}

	oldValue := atomic.LoadInt64(&c.storedTokens)
	newValue := c.coolDownTokens(currentTime, passQps)

	if atomic.CompareAndSwapInt64(&c.storedTokens, oldValue, newValue) {
		if currentValue := atomic.AddInt64(&c.storedTokens, int64(-passQps)); currentValue < 0 {
			atomic.StoreInt64(&c.storedTokens, 0)
		}
		atomic.StoreUint64(&c.lastFilledTime, currentTime)
	}
}

func (c *WarmUpTrafficShapingCalculator) coolDownTokens(currentTime uint64, passQps float64) int64 {
	oldValue := atomic.LoadInt64(&c.storedTokens)
	newValue := oldValue

	// Prerequisites for adding a token:
	// When token consumption is much lower than the warning line
	if oldValue < int64(c.warningToken) {
		newValue = int64(float64(oldValue) + (float64(currentTime)-float64(atomic.LoadUint64(&c.lastFilledTime)))*c.threshold/1000.0)
	} else if oldValue > int64(c.warningToken) {
		if passQps < float64(uint32(c.threshold)/c.coldFactor) {
			newValue = int64(float64(oldValue) + float64(currentTime-atomic.LoadUint64(&c.lastFilledTime))*c.threshold/1000.0)
		}
	}
	if newValue <= int64(c.maxToken) {
		return newValue
	}
	return int64(c.maxToken)
}
