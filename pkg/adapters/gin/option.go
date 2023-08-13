package gin

import (
	"github.com/gin-gonic/gin"
)

type (
	Option  func(*options)
	options struct {
		resourceExtract func(*gin.Context) string
		blockFallback   func(*gin.Context)
		excludePath     func(ctx *gin.Context) bool
	}
)

func evaluateOptions(opts []Option) *options {
	optCopy := &options{}
	for _, opt := range opts {
		opt(optCopy)
	}
	return optCopy
}

// WithResourceExtractor sets the resource extractor of the web requests.
func WithResourceExtractor(fn func(ctx *gin.Context) string) Option {
	return func(o *options) {
		o.resourceExtract = fn
	}
}

// WithBlockFallback sets the fallback handler when requests are blocked.
func WithBlockFallback(fn func(ctx *gin.Context)) Option {
	return func(o *options) {
		o.blockFallback = fn
	}
}

// ExcludePath 请求路径排除
func ExcludePath(fn func(ctx *gin.Context) bool) Option {
	return func(o *options) {
		o.excludePath = fn
	}
}
