package backoff

// SleepingBackOffPolicy 需要进行暂停的policy
type SleepingBackOffPolicy interface {
	BackOffPolicy
	// WithSleeper 对policy进行clone，返回一个使用sleeper进行暂停的新策略
	WithSleeper(sleeper Sleeper) SleepingBackOffPolicy
}
