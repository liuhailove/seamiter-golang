package support

import (
	"errors"
	"git.garena.com/honggang.liu/seamiter-go/core/retry"
	"git.garena.com/honggang.liu/seamiter-go/core/retry/backoff"
	"git.garena.com/honggang.liu/seamiter-go/core/retry/policy"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/util"
)

const (
	GlobalState = "state.global"
)

// RetryTemplate 重试模板，提供具有重试语意的重试模板
// 可重试操作封装在 {@link RetryCallback} 的实现中
// 默认情况下，如果抛出任何异常，则会重试操作，可以使用
// {@link #setRetryPolicy(RetryPolicy)} 方法修改。
// 默认情况下，每个操作最多重试3次且。可以使用
// {@link #setRetryPolicy(RetryPolicy)} 和 {@link #setBackOffPolicy(BackOffPolicy)}
// 修改这个特性。 {@link BackOffPolicy} 控制如何每次重试之间的暂停时间很长。
// 该类是线程安全的，适合在执行操作时并发访问。
type RetryTemplate struct {

	// BackOffPolicy 回退策略
	BackOffPolicy backoff.BackOffPolicy

	// RetryPolicy 重试策略
	RetryPolicy retry.RtyPolicy

	// 重试Listener
	Listeners []retry.RtyListener

	// cache
	RetryContextCache policy.RtyContextCache

	ThrowLastErrorOnExhausted bool
}

func (r *RetryTemplate) DeepCopy(rawTemplate *RetryTemplate) *RetryTemplate {
	return &RetryTemplate{
		BackOffPolicy:             rawTemplate.BackOffPolicy,
		RetryPolicy:               rawTemplate.RetryPolicy,
		Listeners:                 rawTemplate.Listeners,
		RetryContextCache:         rawTemplate.RetryContextCache,
		ThrowLastErrorOnExhausted: rawTemplate.ThrowLastErrorOnExhausted,
	}
}
func (r *RetryTemplate) Execute(callback retry.RtyCallback) (interface{}, error) {
	return r.ExecuteWithState(callback, nil)
}

func (r *RetryTemplate) ExecuteWithRecover(callback retry.RtyCallback, recoverCallback retry.RecoverCallback) (interface{}, error) {
	return r.ExecuteWithRecoverAndState(callback, recoverCallback, nil)
}

func (r *RetryTemplate) ExecuteWithState(callback retry.RtyCallback, state retry.RtyState) (interface{}, error) {
	return r.ExecuteWithRecoverAndState(callback, nil, state)
}

func (r *RetryTemplate) ExecuteWithRecoverAndState(callback retry.RtyCallback, recoverCallback retry.RecoverCallback, state retry.RtyState) (interface{}, error) {
	var retryPolicy = r.RetryPolicy
	var backOffPolicy = r.BackOffPolicy

	// 重试策略初始化
	var ctx = r.open(retryPolicy, state)
	if logging.DebugEnabled() {
		logging.Debug("RetryContext retrieved ", "ctx", ctx)
	}
	// 注册上下文
	// TODO

	var lastError error
	var exhausted = false
	var handleSuccess = false
	var result interface{}
	var handleError error
	util.Try(func() {
		// context增强
		var running = r.doOpenInterceptors(callback, ctx)
		if !running {
			result = nil
			handleError = errors.New("retry terminated abnormally by interceptor before first attempt")
			return
		}
		// 获取或者开启回退上下文
		var backOffContext backoff.BackoffContext
		var resource = ctx.GetAttribute("backOffContext")
		if bctx, ok := resource.(backoff.BackoffContext); ok {
			backOffContext = bctx
		}
		if backOffContext == nil {
			backOffContext = backOffPolicy.Start(ctx)
			if backOffContext != nil {
				ctx.SetAttribute("backOffContext", backOffContext)
			}
		}
		// 如果策略或上下文已经存在，我们允许跳过整个循环
		// 禁止第一次尝试。这用于外部重试的情况，以允许
		// 在没有回调处理的情况下在 handleRetryExhausted 中恢复（这
		// 会抛出异常）
		for {
			if !r.canRetry(retryPolicy, ctx) || ctx.IsExhaustedOnly() {
				break
			}
			var shouldBreak = false
			util.Try(func() {
				if logging.DebugEnabled() {
					logging.Debug("Retry: count=%d", ctx.GetRetryCount())
				}
				lastError = nil
				var res = callback.DoWithRetry(ctx)
				r.doOnSuccessInterceptors(callback, ctx, res)
				result = res
				handleError = nil
				shouldBreak = true
				handleSuccess = true
				return
			}).CatchAll(func(err error) {
				lastError = err
				var e = r.registerError(retryPolicy, state, ctx, err)
				if e != nil {
					result = nil
					handleError = errors.New("Could not register throwable" + e.Error())
					shouldBreak = true
					return
				}
				r.doOnErrorInterceptors(callback, ctx, lastError)
				if r.canRetry(retryPolicy, ctx) && !ctx.IsExhaustedOnly() {
					backOffPolicy.BackOff(backOffContext)
				}
				if logging.DebugEnabled() {
					logging.Debug("Retry for rethrow", "count", ctx.GetRetryCount())
				}
				if r.shouldRethrow(retryPolicy, ctx, state) {
					if logging.DebugEnabled() {
						logging.Debug("Rethrow in retry for policy", "count", ctx.GetRetryCount())
					}
					result = nil
					handleError = errors.New("exception in retry" + lastError.Error())
					shouldBreak = true
					return
				}
				shouldBreak = false
			})
			if shouldBreak {
				break
			}
			// 有状态的重试，可能重新抛出之前的异常。但是如果走到这一部，这是有原因的，比如熔断或者rollback分类
			if state != nil && ctx.HasAttribute(GlobalState) {
				break
			}
			if state == nil && logging.DebugEnabled() {
				logging.Debug("Retry failed last attempt", "count", ctx.GetRetryCount())
			}
		}
		if !handleSuccess {
			exhausted = true
			result, handleError = r.handleRetryExhausted(recoverCallback, ctx, state)
		}
		return
	}).CatchAll(func(err error) {
		logging.Error(err, "error in retry template")
		handleError = err
	}).Finally(func() {
		r.close(retryPolicy, ctx, state, lastError == nil || exhausted)
		r.doCloseInterceptors(callback, ctx, lastError)
		// 清理上下文 TODO
	})
	return result, handleError
}

func (r *RetryTemplate) doOnSuccessInterceptors(callback retry.RtyCallback, ctx retry.RtyContext, result interface{}) {
	for i := len(r.Listeners); i > 0; i-- {
		r.Listeners[i].OnSuccess(ctx, callback, result)
	}
}

func (r *RetryTemplate) doCloseInterceptors(callback retry.RtyCallback, ctx retry.RtyContext, err error) {
	for i := len(r.Listeners); i > 0; i-- {
		r.Listeners[i].Close(ctx, callback, err)
	}
}

// close 清理cache，关闭context
func (r *RetryTemplate) close(rtyPolicy retry.RtyPolicy, ctx retry.RtyContext, state retry.RtyState, succeeded bool) {
	if state != nil {
		if succeeded {
			if !ctx.HasAttribute(GlobalState) {
				r.RetryContextCache.Remove(state.GetKey())
			}
			rtyPolicy.Close(ctx)
			ctx.SetAttribute(retry.Closed, true)
		}
	} else {
		rtyPolicy.Close(ctx)
		ctx.SetAttribute(retry.Closed, true)
	}
}

func (r *RetryTemplate) handleRetryExhausted(callback retry.RecoverCallback, ctx retry.RtyContext, state retry.RtyState) (interface{}, error) {
	ctx.SetAttribute(retry.Exhausted, true)
	if state != nil && !ctx.HasAttribute(GlobalState) {
		r.RetryContextCache.Remove(state.GetKey())
	}
	if callback != nil {
		var recovered = callback.Recover(ctx)
		ctx.SetAttribute(retry.Recovered, true)
		return recovered, nil
	}
	if state != nil {
		if logging.DebugEnabled() {
			logging.Debug("Retry exhausted after last attempt with no recovery path.")
			return nil, errors.New("retry exhausted after last attempt with no recovery path")
		}
	}
	var err error
	if ctx.GetLastError() == nil {
		err = errors.New("exception in retry")
	} else {
		err = ctx.GetLastError()
	}
	return nil, errors.New(err.Error())
}

// shouldRethrow 是一个扩展点，对于子类捕或到异常后，可以决定是否继续抛出异常。一般来说，无状态的行为不需要继续抛出异常，有状态的需要抛出
func (r *RetryTemplate) shouldRethrow(rtyPolicy retry.RtyPolicy, ctx retry.RtyContext, state retry.RtyState) bool {
	return state != nil && state.RollbackFor(ctx.GetLastError())
}

func (r *RetryTemplate) doOnErrorInterceptors(callback retry.RtyCallback, ctx retry.RtyContext, err error) {
	for i := len(r.Listeners); i > 0; i-- {
		r.Listeners[i].OnError(ctx, callback, err)
	}
}

// canRetry 判断当前流程是否可以继续重试.canRetry方法是在RtyCallback执行钱，在backoff和open 拦截器之后执行
func (r *RetryTemplate) canRetry(rtyPolicy retry.RtyPolicy, ctx retry.RtyContext) bool {
	return rtyPolicy.CanRetry(ctx)
}

func (r *RetryTemplate) registerError(rtyPolicy retry.RtyPolicy, state retry.RtyState, ctx retry.RtyContext, err error) error {
	rtyPolicy.RegisterError(ctx, err)
	return r.RegisterContext(ctx, state)
}

func (r *RetryTemplate) RegisterContext(ctx retry.RtyContext, state retry.RtyState) error {
	if state != nil {
		var key = state.GetKey()
		if key != nil {
			if ctx.GetRetryCount() > 1 && !r.RetryContextCache.ContainsKey(key) {
				return errors.New("inconsistent state for failed item key: " +
					"cache key has changed. Consider whether equals() or hashCode() for the key might be inconsistent," +
					"or if you need to supply a better key")
			}
			r.RetryContextCache.Put(key, ctx)
		}
	}
	return nil
}

func (r *RetryTemplate) doOpenInterceptors(callback retry.RtyCallback, ctx retry.RtyContext) bool {
	var result = true
	for _, listener := range r.Listeners {
		result = result && listener.Open(ctx, callback)
	}
	return result
}
func (r *RetryTemplate) open(rtyPolicy retry.RtyPolicy, state retry.RtyState) retry.RtyContext {
	if state == nil {
		return r.doOpenInternalWithoutState(rtyPolicy)
	}
	// TODO 待实现
	return nil
}

func (r *RetryTemplate) doOpenInternalWithoutState(rtyPolicy retry.RtyPolicy) retry.RtyContext {
	return r.doOpenInternal(rtyPolicy, nil)
}

func (r *RetryTemplate) doOpenInternal(rtyPolicy retry.RtyPolicy, state retry.RtyState) retry.RtyContext {
	// TODO 未完全上线
	var ctx = rtyPolicy.Open(nil)
	return ctx
}
