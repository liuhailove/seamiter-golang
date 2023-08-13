package backoff

import (
	"git.garena.com/honggang.liu/seamiter-go/util"
	"time"
)

// Sleeper 用于执行backoff的暂停操作
type Sleeper interface {
	// Sleep 休眠指定间隔,单位毫秒
	Sleep(backOffPeriodInMs int64)
}

// DefaultWaitSleeper 默认休眠
type DefaultWaitSleeper struct {
}

func (d DefaultWaitSleeper) Sleep(backOffPeriodInMs int64) {
	util.Sleep(time.Duration(backOffPeriodInMs) * time.Millisecond)
}
