package backoff

import (
	"git.garena.com/honggang.liu/seamiter-go/core/retry"
	"math/rand"
	"strconv"
	"time"
)

const (
	// DefaultBackoffMinPeriodInMs 最小回退间隔
	DefaultBackoffMinPeriodInMs int64 = 500

	// DefaultBackoffMaxPeriodInMs 最大回退间隔
	DefaultBackoffMaxPeriodInMs int64 = 1500
)

// UniformRandomBackoffPolicy 均匀随机策略
type UniformRandomBackoffPolicy struct {
	MinBackoffPeriod int64
	MaxBackoffPeriod int64
	Sleeper          Sleeper
}

func (u *UniformRandomBackoffPolicy) Start(content retry.RtyContext) BackoffContext {
	return nil
}

func (u *UniformRandomBackoffPolicy) BackOff(ctx BackoffContext) {
	var min = u.MinBackoffPeriod
	var delta int64
	if u.MaxBackoffPeriod == u.MinBackoffPeriod {
		delta = 0
	} else {
		rand.Seed(time.Now().UnixNano())
		delta = rand.Int63n(u.MaxBackoffPeriod - min)
	}
	u.Sleeper.Sleep(min + delta)
}

func (u *UniformRandomBackoffPolicy) String() string {
	return "RandomBackOffPolicy[backOffPeriod=" + strconv.FormatInt(u.MinBackoffPeriod, 10) + ", " + strconv.FormatInt(u.MaxBackoffPeriod, 10) + "]"
}

func (u *UniformRandomBackoffPolicy) WithSleeper(sleeper Sleeper) *UniformRandomBackoffPolicy {
	var res = &UniformRandomBackoffPolicy{
		MinBackoffPeriod: DefaultBackoffMinPeriodInMs,
		MaxBackoffPeriod: DefaultBackoffMaxPeriodInMs,
		Sleeper:          sleeper,
	}
	return res
}
