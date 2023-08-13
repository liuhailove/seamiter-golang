package backoff

import (
	"git.garena.com/honggang.liu/seamiter-go/core/retry"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"strconv"
	"sync"
)

const (
	// DefaultInitialInterval 初始化的时间间隔
	DefaultInitialInterval int64 = 100
	// DefaultMaxInterval 最大时间间隔，默认30S
	DefaultMaxInterval int64 = 30000
	// DefaultMultiplier 默认倍率，每次回退增加100%
	DefaultMultiplier int32 = 2
)

// ExponentialBackOffPolicy 指数回退策略
type ExponentialBackOffPolicy struct {
	// InitialInterval 初始化的时间间隔
	InitialInterval int64
	// MaxInterval 最大时间间隔
	MaxInterval int64
	// Multiplier 重试指数倍率
	Multiplier int32
	// Sleeper
	Sleeper Sleeper
}

// NewInstance 实例化
func (e *ExponentialBackOffPolicy) NewInstance() *ExponentialBackOffPolicy {
	var instance = new(ExponentialBackOffPolicy)
	instance.InitialInterval = DefaultInitialInterval
	instance.MaxInterval = DefaultMaxInterval
	instance.Multiplier = DefaultMultiplier
	instance.Sleeper = new(DefaultWaitSleeper)
	return instance
}

func (e *ExponentialBackOffPolicy) WithSleeper(sleeper Sleeper) SleepingBackOffPolicy {
	var res = e.NewInstance()
	e.CloneValues(res)
	res.Sleeper = sleeper
	return res
}

func (e *ExponentialBackOffPolicy) CloneValues(target *ExponentialBackOffPolicy) {
	target.SetMaxInterval(e.GetMaxInterval())
	target.SetInitialInterval(e.GetInitialInterval())
	target.SetMultiplier(e.GetMultiplier())
	target.Sleeper = e.Sleeper
}

func (e *ExponentialBackOffPolicy) SetInitialInterval(initialInterval int64) {
	if initialInterval > 1 {
		e.InitialInterval = initialInterval
	} else {
		e.InitialInterval = 1
	}
}

func (e *ExponentialBackOffPolicy) SetMultiplier(multiplier int32) {
	if multiplier > 1 {
		e.Multiplier = multiplier
	} else {
		e.Multiplier = 1
	}
}

func (e *ExponentialBackOffPolicy) SetMaxInterval(maxInterval int64) {
	if maxInterval > 0 {
		e.MaxInterval = maxInterval
	} else {
		e.MaxInterval = 1
	}
}

func (e *ExponentialBackOffPolicy) SetSleeper(sleeper Sleeper) {
	if sleeper == nil {
		e.Sleeper = DefaultWaitSleeper{}
	} else {
		e.Sleeper = sleeper
	}
}

func (e *ExponentialBackOffPolicy) GetInitialInterval() int64 {
	return e.InitialInterval
}

func (e *ExponentialBackOffPolicy) GetMaxInterval() int64 {
	return e.MaxInterval
}

func (e *ExponentialBackOffPolicy) GetMultiplier() int32 {
	return e.Multiplier
}
func (e *ExponentialBackOffPolicy) Start(_ retry.RtyContext) BackoffContext {
	return &ExponentialBackOffContext{
		Multiplier:  e.Multiplier,
		Interval:    e.InitialInterval,
		MaxInterval: e.MaxInterval,
	}
}

func (e *ExponentialBackOffPolicy) BackOff(ctx BackoffContext) {
	var context = (ctx).(*ExponentialBackOffContext)
	var sleepTime = context.GetSleepAndIncrement()
	if logging.DebugEnabled() {
		logging.Debug("Sleeping for " + strconv.FormatInt(sleepTime, 10))
	}
	e.Sleeper.Sleep(sleepTime)
}

func (e *ExponentialBackOffPolicy) String() string {
	return "ExponentialBackOffPolicy[initialInterval=" + strconv.FormatInt(e.InitialInterval, 10) + ",multiplier=" + strconv.FormatInt(int64(e.Multiplier), 10) + ",maxInterval=" + strconv.FormatInt(e.MaxInterval, 10) + "]"
}

var (
	bcMux = new(sync.RWMutex)
)

// ExponentialBackOffContext 指数回退context
type ExponentialBackOffContext struct {
	Multiplier  int32
	Interval    int64
	MaxInterval int64
}

func (e *ExponentialBackOffContext) GetSleepAndIncrement() int64 {
	bcMux.Lock()
	defer bcMux.Unlock()
	var sleep = e.Interval
	if sleep > e.MaxInterval {
		sleep = e.MaxInterval
	} else {
		e.Interval = e.GetNextInterval()
	}
	return sleep
}

func (e *ExponentialBackOffContext) GetNextInterval() int64 {
	return e.Interval * int64(e.Multiplier)
}

func (e *ExponentialBackOffContext) String() string {
	return "ExponentialBackOffContext" + "[initialInterval=" + strconv.FormatInt(e.Interval, 10) +
		", multiplier=" + strconv.FormatInt(int64(e.Multiplier), 10) + ", maxInterval=" + strconv.FormatInt(e.MaxInterval, 10) + "]"
}
