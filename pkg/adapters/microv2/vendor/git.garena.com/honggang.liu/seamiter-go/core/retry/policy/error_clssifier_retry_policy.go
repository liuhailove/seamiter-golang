package policy

import (
	"git.garena.com/honggang.liu/seamiter-go/core/retry"
	"git.garena.com/honggang.liu/seamiter-go/core/retry/classify"
	"git.garena.com/honggang.liu/seamiter-go/core/retry/context"
)

type ErrorClassifierRetryPolicy struct {
	ErrorClassifier *classify.ErrorClassifier
}

func (e ErrorClassifierRetryPolicy) CanRetry(ctx retry.RtyContext) bool {
	var err = ctx.GetLastError()
	return err == nil || e.ErrorClassifier.Classify(err)
}

func (e ErrorClassifierRetryPolicy) Open(parent retry.RtyContext) retry.RtyContext {
	return &context.RtyContextSupport{Parent: parent}
}

func (e ErrorClassifierRetryPolicy) Close(ctx retry.RtyContext) {
	// no-op
}

func (e ErrorClassifierRetryPolicy) RegisterError(ctx retry.RtyContext, err error) {
	var simpleContext = ctx.(*context.RtyContextSupport)
	simpleContext.RegisterError(err)
}
