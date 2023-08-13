package backoff

import (
	"git.garena.com/honggang.liu/seamiter-go/core/retry"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"math/rand"
	"strconv"
	"time"
)

// ExponentialRandomBackOffPolicy 使用随机倍率的回退方法.
// 随机倍率在[1,multiplier]之间.对于指数回退，假设interval=50，multiplier=2，
// maxInterval=30000,numRetries=5,则每次的sleep[50,100,200,400,800],
// 而ExponentialRandomBackOffPolicy的sleep为[53,190,267,451,815]
type ExponentialRandomBackOffPolicy struct {
	ExponentialBackOffPolicy
}

func (e *ExponentialRandomBackOffPolicy) Start(_ retry.RtyContext) BackoffContext {
	return NewExponentialBackOffContext(e.GetInitialInterval(), e.GetMultiplier(), e.GetMaxInterval())
}

func (e *ExponentialRandomBackOffPolicy) NewInstance() *ExponentialBackOffPolicy {
	var instance = &ExponentialRandomBackOffPolicy{}
	return instance.ExponentialBackOffPolicy.NewInstance()
}

func (e *ExponentialRandomBackOffPolicy) BackOff(ctx BackoffContext) {
	var context = (ctx).(*ExponentialRandomBackOffContext)
	var sleepTime = context.GetSleepAndIncrement()
	if logging.DebugEnabled() {
		logging.Debug("Sleeping for " + strconv.FormatInt(sleepTime, 10))
	}
	e.ExponentialBackOffPolicy.Sleeper.Sleep(sleepTime)
}

type ExponentialRandomBackOffContext struct {
	ExponentialBackOffContext
}

func (e *ExponentialRandomBackOffContext) GetSleepAndIncrement() int64 {
	var next = e.ExponentialBackOffContext.GetSleepAndIncrement()
	rand.Seed(time.Now().UnixNano())
	next = next * int64(1+rand.Int31n(e.Multiplier))
	return next
}

func NewExponentialBackOffContext(expSeed int64, multiplier int32, maxInterval int64) *ExponentialRandomBackOffContext {
	var instance = new(ExponentialRandomBackOffContext)
	instance.Interval = expSeed
	instance.Multiplier = multiplier
	instance.MaxInterval = maxInterval
	return instance
}
