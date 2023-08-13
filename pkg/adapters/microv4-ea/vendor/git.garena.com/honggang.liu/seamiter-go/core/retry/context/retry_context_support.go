package context

import (
	"fmt"
	"git.garena.com/honggang.liu/seamiter-go/core/retry"
)

type RtyContextSupport struct {
	Parent retry.RtyContext
	retry.SimpleAttributeAccessorSupport
	Terminate bool
	Count     int32
	LastError error
}

func (r *RtyContextSupport) GetRetryCount() int32 {
	return r.Count
}

func (r *RtyContextSupport) GetLastError() error {
	return r.LastError
}

func (r *RtyContextSupport) SetExhaustedOnly() {
	r.Terminate = true
}

func (r *RtyContextSupport) IsExhaustedOnly() bool {
	return r.Terminate
}

func (r *RtyContextSupport) GetParent() retry.RtyContext {
	return r.Parent
}

func (r *RtyContextSupport) RegisterError(err error) {
	r.LastError = err
	if err != nil {
		r.Count++
	}
}

func (r *RtyContextSupport) String() string {
	return fmt.Sprintf("[RetryContext: count=%d, lastException=%s, exhausted=%v]", r.Count, r.LastError, r.Terminate)
}
