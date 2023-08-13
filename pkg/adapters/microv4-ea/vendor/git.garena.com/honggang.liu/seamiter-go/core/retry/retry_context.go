package retry

const (
	Name      = "context.name"
	StateKey  = "state.key"
	Closed    = "context.closed"
	Recovered = "context.recovered"
	Exhausted = "context.exhausted"
)

// RtyContext 重试上下文
type RtyContext interface {
	// GetRetryCount 返回重试的计数，在重试之前，这个计数应该为0，
	// 重试后依次递增
	GetRetryCount() int32

	// GetLastError 引起重试的错误。
	GetLastError() error

	// SetExhaustedOnly 发出信号，用于表用不应在重试，或者重试当前的
	SetExhaustedOnly()

	// IsExhaustedOnly 获取设置的SetExhaustedOnly的标记
	IsExhaustedOnly() bool

	// GetParent 如果重试块有嵌套，则获取parent ctx
	GetParent() RtyContext

	AttributeAccessorSupport
}
