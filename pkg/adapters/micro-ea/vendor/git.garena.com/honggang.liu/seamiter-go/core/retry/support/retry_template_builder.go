package support

import (
	"errors"
	"git.garena.com/honggang.liu/seamiter-go/core/retry"
	"git.garena.com/honggang.liu/seamiter-go/core/retry/backoff"
	"git.garena.com/honggang.liu/seamiter-go/core/retry/classify"
	"git.garena.com/honggang.liu/seamiter-go/core/retry/policy"
)

// * Fluent API to configure new instance of RetryTemplate. For detailed description of each
// * builder method - see it's doc.
// *
// * <p>
// * Examples: <pre>{@code
// * RetryTemplate.builder()
// *      .maxAttempts(10)
// *      .exponentialBackoff(100, 2, 10000)
// *      .retryOn(IOException.class)
// *      .traversingCauses()
// *      .build();
// *
// * RetryTemplate.builder()
// *      .fixedBackoff(10)
// *      .withinMillis(3000)
// *      .build();
// *
// * RetryTemplate.builder()
// *      .infiniteRetry()
// *      .retryOn(IOException.class)
// *      .uniformRandomBackoff(1000, 3000)
// *      .build();
// * }</pre>
// *
// * <p>
// * The builder provides the following defaults:
// * <ul>
// * <li>retry policy: max attempts = 3 (initial + 2 retries)</li>
// * <li>backoff policy: no backoff (retry immediately)</li>
// * <li>exception classification: retry only on {@link Exception} and it's subclasses,
// * without traversing of causes</li>
// * </ul>
// *
// * <p>
// * The builder supports only widely used properties of {@link RetryTemplate}. More
// * specific properties can be configured directly (after building).
// *
// * <p>
// * Not thread safe. Building should be performed in a single thread. Also, there is no
// * guarantee that all constructors of all fields are thread safe in-depth (means employing
// * only volatile and final writes), so, in concurrent environment, it is recommended to
// * ensure presence of happens-before between publication and any usage. (e.g. publication
// * via volatile write, or other safe publication technique)

type RetryTemplateBuilder struct {
	// 重试策略
	BaseRtyPolicy retry.RtyPolicy
	// 回退策略
	BackOffPolicy backoff.BackOffPolicy
	// 重试监听
	Listeners []retry.RtyListener
	// 异常分类构建器
	ClassifierBuilder *classify.ErrorClassifierBuilder

	// 异常匹配模式
	ErrorMatcher classify.PatternMatcher
}

// NeverRtyPolicy 不重试策略
func (e *RetryTemplateBuilder) NeverRtyPolicy() *RetryTemplateBuilder {
	if e.BaseRtyPolicy != nil {
		panic("You have already selected backoff policy")
	}
	e.BaseRtyPolicy = policy.NewNeverRetryPolicy()
	return e
}

// NewSimpleRetryPolicyWithMaxAttemptsAndErrors 简单重试
func (e *RetryTemplateBuilder) NewSimpleRetryPolicyWithMaxAttemptsAndErrors(maxAttempts int32, errs []error) *RetryTemplateBuilder {
	if e.BaseRtyPolicy != nil {
		panic("You have already selected backoff policy")
	}
	if maxAttempts <= 0 {
		e.BaseRtyPolicy = policy.NewSimpleRetryPolicy()
	} else {
		e.BaseRtyPolicy = policy.NewSimpleRetryPolicyWithMaxAttemptsAndErrors(maxAttempts, errs)
	}
	return e
}

// MaxAttemptsRtyPolicy 设置最大重试策略的重试
func (e *RetryTemplateBuilder) MaxAttemptsRtyPolicy(maxAttempts int32) *RetryTemplateBuilder {
	if e.BaseRtyPolicy != nil {
		panic("You have already selected backoff policy")
	}
	if maxAttempts <= 0 {
		e.BaseRtyPolicy = policy.NewMaxAttemptsRetryPolicy()
	} else {
		e.BaseRtyPolicy = policy.NewMaxAttemptsRetryPolicyWithAttempts(maxAttempts)
	}
	return e
}

// WithinMillisRtyPolicy 设置具备timeout的重试策略
func (e *RetryTemplateBuilder) WithinMillisRtyPolicy(timeoutInMs int64) *RetryTemplateBuilder {
	//if timeoutInMs <= 0 {
	//	panic("Timeout should be positive")
	//}
	if e.BaseRtyPolicy != nil {
		panic("You have already selected backoff policy")
	}
	var timeoutRetryPolicy *policy.TimeoutRtyPolicy
	if timeoutInMs <= 0 {
		timeoutRetryPolicy = &policy.TimeoutRtyPolicy{Timeout: policy.DefaultTimeout}
	} else {
		timeoutRetryPolicy = &policy.TimeoutRtyPolicy{Timeout: timeoutInMs}
	}
	e.BaseRtyPolicy = timeoutRetryPolicy
	return e
}

// InfiniteRtyPolicy 构建没有重试限制的重试
func (e *RetryTemplateBuilder) InfiniteRtyPolicy() *RetryTemplateBuilder {
	if e.BaseRtyPolicy != nil {
		panic("You have already selected backoff policy")
	}
	e.BaseRtyPolicy = &policy.AlwaysRetryPolicy{}
	return e
}

// CustomPolicyRtyPolicy 用户自定义策略，如果提供的策略不能满足用户需求，用户可以提供自己的方法来实现
func (e *RetryTemplateBuilder) CustomPolicyRtyPolicy(rtyPolicy retry.RtyPolicy) *RetryTemplateBuilder {
	if e.BaseRtyPolicy != nil {
		panic("You have already selected backoff policy")
	}
	e.BaseRtyPolicy = rtyPolicy
	return e
}

// ExponentialBackoff 构建一个具有指数回退的策略，回退表达式：currentInterval = Math.min(initialInterval * Math.pow(multiplier, retryNum), maxInterval)
func (e *RetryTemplateBuilder) ExponentialBackoff(initialInterval int64, multiplier int32, maxInterval int64) *RetryTemplateBuilder {
	return e.ExponentialBackoffWithRandom(initialInterval, multiplier, maxInterval, false)
}

// ExponentialBackoffWithRandom 构建具备随机回退或者固定回退的指数回退策略
func (e *RetryTemplateBuilder) ExponentialBackoffWithRandom(initialInterval int64, multiplier int32, maxInterval int64, withRandom bool) *RetryTemplateBuilder {
	var backOffPolicy backoff.BackOffPolicy
	if withRandom {
		var exponentialRandomBackOffPolicy = &backoff.ExponentialRandomBackOffPolicy{}
		exponentialRandomBackOffPolicy.ExponentialBackOffPolicy.SetInitialInterval(initialInterval)
		exponentialRandomBackOffPolicy.ExponentialBackOffPolicy.SetMultiplier(multiplier)
		exponentialRandomBackOffPolicy.ExponentialBackOffPolicy.SetMaxInterval(maxInterval)
		exponentialRandomBackOffPolicy.ExponentialBackOffPolicy.SetSleeper(backoff.DefaultWaitSleeper{})
		backOffPolicy = exponentialRandomBackOffPolicy
	} else {
		var exponentialBackOffPolicy = &backoff.ExponentialBackOffPolicy{}
		exponentialBackOffPolicy.SetInitialInterval(initialInterval)
		exponentialBackOffPolicy.SetMultiplier(multiplier)
		exponentialBackOffPolicy.SetMaxInterval(maxInterval)
		exponentialBackOffPolicy.SetSleeper(backoff.DefaultWaitSleeper{})
		backOffPolicy = exponentialBackOffPolicy
	}
	e.BackOffPolicy = backOffPolicy
	return e
}

// FixedBackoff 构建固定间隔的回退策略
func (e *RetryTemplateBuilder) FixedBackoff(interval int64) *RetryTemplateBuilder {
	var backoffPolicy = &backoff.FixedBackOffPolicy{BackOffPeriodInMs: interval, Sleeper: backoff.DefaultWaitSleeper{}}
	e.BackOffPolicy = backoffPolicy
	return e
}

// UniformRandomBackoff 构建均匀随机策略
func (e *RetryTemplateBuilder) UniformRandomBackoff(minInterval, maxInterval int64) *RetryTemplateBuilder {
	var backoffPolicy = &backoff.UniformRandomBackoffPolicy{
		MinBackoffPeriod: minInterval,
		MaxBackoffPeriod: maxInterval,
		Sleeper:          backoff.DefaultWaitSleeper{},
	}
	e.BackOffPolicy = backoffPolicy
	return e
}

// NoBackoff 构建没有任何回退策略的策略
func (e *RetryTemplateBuilder) NoBackoff() *RetryTemplateBuilder {
	e.BackOffPolicy = &backoff.NoBackOffPolicy{}
	return e
}

// CustomBackOff 客户自定义回退策略
func (e *RetryTemplateBuilder) CustomBackOff(backOffPolicy backoff.BackOffPolicy) *RetryTemplateBuilder {
	e.BackOffPolicy = backOffPolicy
	return e
}

// RetryOn 增加一种可重试的异常
func (e *RetryTemplateBuilder) RetryOn(err error) *RetryTemplateBuilder {
	e.classifierBuilder().RetryOn(err)
	return e
}

// NotRetryOn 增加一种不进行重试的异常
func (e *RetryTemplateBuilder) NotRetryOn(err error) *RetryTemplateBuilder {
	e.classifierBuilder().NotRetryOn(err)
	return e
}

// RetryOnErrors 增加一组可以重试的异常
func (e *RetryTemplateBuilder) RetryOnErrors(errs []error) *RetryTemplateBuilder {
	for _, err := range errs {
		e.classifierBuilder().RetryOn(err)
	}
	return e
}

// NotRetryOnErrors 增加一组不可以重试的异常
func (e *RetryTemplateBuilder) NotRetryOnErrors(errs []error) *RetryTemplateBuilder {
	for _, err := range errs {
		e.classifierBuilder().NotRetryOn(err)
	}
	return e
}

func (e *RetryTemplateBuilder) WithErrorMatchPattern(errorMatcher classify.PatternMatcher) *RetryTemplateBuilder {
	e.ErrorMatcher = errorMatcher
	return e
}

// WithListener 增加监听
func (e *RetryTemplateBuilder) WithListener(listener retry.RtyListener) *RetryTemplateBuilder {
	e.Listeners = append(e.Listeners, listener)
	return e
}

// WithListeners 增加监听
func (e *RetryTemplateBuilder) WithListeners(listeners []retry.RtyListener) *RetryTemplateBuilder {
	for _, listener := range listeners {
		e.Listeners = append(e.Listeners, listener)
	}
	return e
}

// Build 构建RetryTemplate
func (e *RetryTemplateBuilder) Build() *RetryTemplate {
	var retryTemplate = &RetryTemplate{}
	var errClassifier *classify.ErrorClassifier
	if e.ClassifierBuilder != nil {
		errClassifier = e.ClassifierBuilder.Build(e.ErrorMatcher)
	} else {
		errClassifier = &classify.ErrorClassifier{
			DefaultValue: false,
			Classified:   []error{errors.New("any")},
			Matcher:      classify.AnyMatch,
		}
	}
	if e.BaseRtyPolicy == nil {
		e.BaseRtyPolicy = policy.NewMaxAttemptsRetryPolicy()
	}
	var finalPolicy = &policy.CompositeRetryPolicy{}
	finalPolicy.SetPolicies([]retry.RtyPolicy{e.BaseRtyPolicy, policy.ErrorClassifierRetryPolicy{ErrorClassifier: errClassifier}})
	retryTemplate.RetryPolicy = finalPolicy

	// 回退策略
	if e.BackOffPolicy == nil {
		e.BackOffPolicy = &backoff.NoBackOffPolicy{}
	}
	retryTemplate.BackOffPolicy = e.BackOffPolicy

	// listeners
	if e.Listeners != nil {
		retryTemplate.Listeners = e.Listeners
	}
	return retryTemplate
}

func (e *RetryTemplateBuilder) classifierBuilder() *classify.ErrorClassifierBuilder {
	if e.ClassifierBuilder == nil {
		e.ClassifierBuilder = &classify.ErrorClassifierBuilder{}
	}
	return e.ClassifierBuilder
}

func (e *RetryTemplateBuilder) ListenersList() []retry.RtyListener {
	return e.Listeners
}

func NewRetryTemplateBuilder() *RetryTemplateBuilder {
	return &RetryTemplateBuilder{}
}
