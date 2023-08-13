package backoff

import (
	"git.garena.com/honggang.liu/seamiter-go/core/retry"
	"strconv"
)

const (
	// DefaultBackOffPeriod 默认回归时间，1000毫秒
	DefaultBackOffPeriod int64 = 1000
)

// FixedBackOffPolicy 固定时长间隔的回退策略
type FixedBackOffPolicy struct {
	// BackOffPeriodInMs 回退时间间隔
	BackOffPeriodInMs int64
	// 休眠sleeper
	Sleeper Sleeper
}

func (f *FixedBackOffPolicy) Start(_ retry.RtyContext) BackoffContext {
	return nil
}

func (f *FixedBackOffPolicy) BackOff(_ BackoffContext) {
	f.Sleeper.Sleep(f.BackOffPeriodInMs)
}

func (f *FixedBackOffPolicy) String() string {
	return "FixedBackOffPolicy[backOffPeriod=" + strconv.FormatInt(f.BackOffPeriodInMs, 10) + "]"
}

func (f *FixedBackOffPolicy) WithSleeper(sleeper Sleeper) SleepingBackOffPolicy {
	var res = new(FixedBackOffPolicy)
	res.setBackOffPeriod(f.BackOffPeriodInMs)
	res.setSleeper(sleeper)
	return res
}

func (f *FixedBackOffPolicy) setBackOffPeriod(backOffPeriodInMs int64) {
	if backOffPeriodInMs > 0 {
		f.BackOffPeriodInMs = backOffPeriodInMs
	} else {
		f.BackOffPeriodInMs = 1
	}
}

func (f *FixedBackOffPolicy) getBackOffPeriod() int64 {
	return f.BackOffPeriodInMs
}

func (f *FixedBackOffPolicy) setSleeper(sleeper Sleeper) {
	f.Sleeper = sleeper
}
