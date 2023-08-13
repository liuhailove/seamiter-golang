package policy

import (
	"git.garena.com/honggang.liu/seamiter-go/core/retry"
	"git.garena.com/honggang.liu/seamiter-go/core/retry/context"
)

// CompositeRetryPolicy 组合了一组策略，并按序代理调用他们
type CompositeRetryPolicy struct {
	Policies   []retry.RtyPolicy
	Optimistic bool
}

func (c *CompositeRetryPolicy) SetOptimistic(optimistic bool) {
	c.Optimistic = optimistic
}

func (c *CompositeRetryPolicy) SetPolicies(rtyPolicy []retry.RtyPolicy) {
	c.Policies = rtyPolicy
}

func (c *CompositeRetryPolicy) CanRetry(ctx retry.RtyContext) bool {
	var ctxs = (ctx.(*CompositeRetryContext)).Contexts
	var polices = (ctx.(*CompositeRetryContext)).Policies
	var retryable = true
	if c.Optimistic {
		retryable = false
		for i := 0; i < len(ctxs); i++ {
			if polices[i].CanRetry(ctxs[i]) {
				retryable = true
			}
		}
	} else {
		for i := 0; i < len(ctxs); i++ {
			if !polices[i].CanRetry(ctxs[i]) {
				retryable = false
			}
		}
	}
	return retryable
}

func (c *CompositeRetryPolicy) Open(parent retry.RtyContext) retry.RtyContext {
	var list []retry.RtyContext
	for _, policy := range c.Policies {
		list = append(list, policy.Open(parent))
	}
	return NewCompositeRetryContext(parent, list, c.Policies)
}

func (c *CompositeRetryPolicy) Close(ctx retry.RtyContext) {
	var ctxs = (ctx.(*CompositeRetryContext)).Contexts
	var polices = (ctx.(*CompositeRetryContext)).Policies
	for i := 0; i < len(ctxs); i++ {
		polices[i].Close(ctxs[i])
	}
}

func (c *CompositeRetryPolicy) RegisterError(ctx retry.RtyContext, err error) {
	var ctxs = (ctx.(*CompositeRetryContext)).Contexts
	var polices = (ctx.(*CompositeRetryContext)).Policies
	for i := 0; i < len(ctxs); i++ {
		polices[i].RegisterError(ctxs[i], err)
	}
	(ctx.(*CompositeRetryContext)).RegisterError(err)
}

type CompositeRetryContext struct {
	Contexts []retry.RtyContext
	Policies []retry.RtyPolicy
	context.RtyContextSupport
	retry.SimpleAttributeAccessorSupport
}

func NewCompositeRetryContext(parent retry.RtyContext, contexts []retry.RtyContext, policies []retry.RtyPolicy) *CompositeRetryContext {
	var compositeRetryContext = new(CompositeRetryContext)
	compositeRetryContext.Parent = parent
	compositeRetryContext.Contexts = contexts
	compositeRetryContext.Policies = policies
	return compositeRetryContext
}
