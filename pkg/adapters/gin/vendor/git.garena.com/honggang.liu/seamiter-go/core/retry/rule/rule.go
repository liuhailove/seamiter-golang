package rule

import (
	"fmt"
	"strconv"
)

// RetryPolicyType 重试策略
type RetryPolicyType int32

const (
	// NeverRetryPolicy 仅仅第一次重试，之后不允许重试，默认策略，
	NeverRetryPolicy RetryPolicyType = iota
	// SimpleRetryPolicy 简单重试策略，失败后直接重试，没有休眠等
	SimpleRetryPolicy
	// TimeoutRtyPolicy 超时重试策略， 只有在没有超时的情况下才进行重试，超时后则退出重试
	TimeoutRtyPolicy
	// MaxAttemptsRetryPolicy 设置最大重试次数的重试策略，耗尽后则不再重试
	MaxAttemptsRetryPolicy
	// ErrorClassifierRetryPolicy 自定义错误分类器的重试策略
	ErrorClassifierRetryPolicy
	// AlwaysRetryPolicy 一直重试，直到成功
	AlwaysRetryPolicy
	// CompositeRetryPolicy 组合重试策略，按照一定的组合顺序进行重试
	CompositeRetryPolicy
	// CustomPolicyRtyPolicy 用户自定义重试策略
	CustomPolicyRtyPolicy
)

func (t RetryPolicyType) String() string {
	switch t {
	case NeverRetryPolicy:
		return "NeverRetryPolicy"
	case SimpleRetryPolicy:
		return "SimpleRetryPolicy"
	case TimeoutRtyPolicy:
		return "TimeoutRtyPolicy"
	case MaxAttemptsRetryPolicy:
		return "MaxAttemptsRetryPolicy"
	case ErrorClassifierRetryPolicy:
		return "ErrorClassifierRetryPolicy"
	case AlwaysRetryPolicy:
		return "AlwaysRetryPolicy"
	case CompositeRetryPolicy:
		return "CompositeRetryPolicy"
	case CustomPolicyRtyPolicy:
		return "CustomPolicyRtyPolicy"
	default:
		return strconv.Itoa(int(t))
	}
}

// BackoffPolicyType 退避策略.
type BackoffPolicyType int32

const (
	// NoBackOffPolicy 不回退，立即重试
	NoBackOffPolicy BackoffPolicyType = iota
	// FixedBackOffPolicy 休眠规定时长的回退策略
	FixedBackOffPolicy
	// ExponentialBackOffPolicy 指数退避策略回退，每次回退会在上一次基础上乘N倍
	ExponentialBackOffPolicy
	// ExponentialRandomBackOffPolicy 使用随机倍率的回退
	ExponentialRandomBackOffPolicy
	// UniformRandomBackoffPolicy 均匀随机回退
	UniformRandomBackoffPolicy
)

func (t BackoffPolicyType) String() string {
	switch t {
	case NoBackOffPolicy:
		return "NoBackOffPolicy"
	case FixedBackOffPolicy:
		return "FixedBackOffPolicy"
	case ExponentialBackOffPolicy:
		return "ExponentialBackOffPolicy"
	case ExponentialRandomBackOffPolicy:
		return "ExponentialRandomBackOffPolicy"
	case UniformRandomBackoffPolicy:
		return "UniformRandomBackoffPolicy"
	default:
		return strconv.Itoa(int(t))
	}
}

// ErrorMatcherType 异常匹配模式
type ErrorMatcherType int32

const (
	ExactMatch   ErrorMatcherType = iota // 精确匹配
	PrefixMatch                          // 前缀匹配
	SuffixMatch                          // 后缀匹配
	ContainMatch                         // 包含匹配
	RegularMatch                         // 正则匹配
	AnyMatch                             // 只要不为空，则匹配
)

func (t ErrorMatcherType) String() string {
	switch t {
	case ExactMatch:
		return "ExactMatch"
	case PrefixMatch:
		return "PrefixMatch"
	case SuffixMatch:
		return "SuffixMatch"
	case ContainMatch:
		return "ContainMatch"
	case RegularMatch:
		return "RegularMatch"
	case AnyMatch:
		return "AnyMatch"
	default:
		return strconv.Itoa(int(t))
	}
}

// Rule 重试规则.
type Rule struct {
	// unique id
	Id string `json:"id,omitempty"`
	// resource name
	Resource string `json:"resource"`
	// 重试策略相关参数
	// 重试策略类型,该参数为空时是，失败立即重试，重试的时候阻塞线程
	RetryPolicy RetryPolicyType `json:"retryPolicy"`
	// 最大重试次数，此值只有在RetryPolicyType为MaxAttemptsRetryPolicy/SimpleRetryPolicy时才生效，默认为3,包括第一次失败
	RetryMaxAttempts int32 `json:"retryMaxAttempts"`
	// 超时时间，此值只有在RetryPolicyType为TimeoutRtyPolicy时才生效，默认为1000毫秒
	RetryTimeout int64 `json:"retryTimeout"`

	// 回退策略相关参数
	// 回退策略
	BackoffPolicy BackoffPolicyType `json:"backoffPolicy"`
	// **固定时长间隔的回退策略,在BackoffPolicyType为FixedBackOffPolicy才有值，默认为1000毫秒**
	FixedBackOffPeriodInMs int64 `json:"fixedBackOffPeriodInMs"`

	// **指数回退策略时参数，包含ExponentialBackOffPolicy/ExponentialRandomBackOffPolicy**
	// 延迟时间，单位毫秒，默认值1000，即默认延迟1秒。
	// 当未设置multiplier时，表示每隔backoffDelay的时间重试，
	// 直到重试次数到达maxAttempts设置的最大允许重试次数。
	// 当设置了multiplier参数时，该值作为幂运算的初始值。
	BackoffDelay int64 `json:"backoffDelay"`
	// 两次重试间最大间隔时间。
	// 当设置multiplier参数后，下次延迟时间根据是上次延迟时间乘以multiplier得出的，
	// 这会导致两次重试间的延迟时间越来越长，该参数限制两次重试的最大间隔时间
	// 当间隔时间大于该值时，计算出的间隔时间将会被忽略，使用上次的重试间隔时间。
	BackoffMaxDelay int64 `json:"backoffMaxDelay"`
	// 作为乘数用于计算下次延迟时间。公式：delay = delay * multiplier
	BackoffMultiplier int32 `json:"backoffMultiplier"`
	// **均匀随机策略参数**
	// UniformMaxBackoffPeriod 最小回退间隔，默认500ms
	UniformMinBackoffPeriod int64 `json:"uniformMinBackoffPeriod"`
	// UniformMaxBackoffPeriod 最大回退间隔，默认1500ms
	UniformMaxBackoffPeriod int64 `json:"uniformMaxBackoffPeriod"`

	// **异常分类**
	ErrorMatcher ErrorMatcherType `json:"errorMatcher"`
	// 需要重试的异常,默认为空。当参数exclude也为空时，所有异常都将要求重试
	IncludeExceptions []string `json:"includeExceptions"`
	// 不需要重试的异常。默认为空，当参include也为空时，所有异常都将要求重试
	ExcludeExceptions []string `json:"excludeExceptions"`
}

func (r *Rule) String() string {
	// fallback string
	return fmt.Sprintf("{id=%s, resource=%s, retryPolicy=%s,retryMaxAttempts=%d,retryTimeout=%d,"+
		"backoffPolicy=%s,fixedBackOffPeriodInMs=%d,backoffDelay=%d,backoffMaxDelay=%d,backoffMultiplier=%d,uniformMinBackoffPeriod=%d,uniformMaxBackoffPeriod=%d,"+
		"errorMatcher=%s,includeExceptions=%s,excludeExceptions=%s}",
		r.Id, r.Resource, r.RetryPolicy.String(), r.RetryMaxAttempts, r.RetryTimeout,
		r.BackoffPolicy, r.FixedBackOffPeriodInMs, r.BackoffDelay, r.BackoffMaxDelay, r.BackoffMultiplier, r.UniformMinBackoffPeriod, r.UniformMaxBackoffPeriod,
		r.ErrorMatcher.String(), r.IncludeExceptions, r.ExcludeExceptions)
}

func (r *Rule) isStatReusable(newRule *Rule) bool {
	if newRule == nil {
		return false
	}
	var basic = r.Resource == newRule.Resource && r.RetryPolicy == newRule.RetryPolicy && r.RetryMaxAttempts == newRule.RetryMaxAttempts && r.RetryTimeout == newRule.RetryTimeout &&
		r.BackoffPolicy == newRule.BackoffPolicy && r.FixedBackOffPeriodInMs == newRule.FixedBackOffPeriodInMs && r.BackoffDelay == newRule.BackoffDelay && r.BackoffMultiplier == newRule.BackoffMultiplier && r.UniformMinBackoffPeriod == newRule.UniformMinBackoffPeriod && r.UniformMaxBackoffPeriod == newRule.UniformMaxBackoffPeriod &&
		r.ErrorMatcher == newRule.ErrorMatcher

	if !basic {
		return false
	}
	if len(r.IncludeExceptions) != len(newRule.IncludeExceptions) {
		return false
	}
	for i := 0; i < len(r.IncludeExceptions); i++ {
		if r.IncludeExceptions[i] != newRule.IncludeExceptions[i] {
			return false
		}
	}
	if len(r.ExcludeExceptions) != len(newRule.ExcludeExceptions) {
		return false
	}
	for i := 0; i < len(r.ExcludeExceptions); i++ {
		if r.ExcludeExceptions[i] != newRule.ExcludeExceptions[i] {
			return false
		}
	}
	return true
}

func (r *Rule) ResourceName() string {
	return r.Resource
}

func (r *Rule) isEqualsToBase(newRule *Rule) bool {
	return r.isStatReusable(newRule)
}

func (r *Rule) isEqualTo(newRule *Rule) bool {
	return r.isEqualsToBase(newRule)
}
