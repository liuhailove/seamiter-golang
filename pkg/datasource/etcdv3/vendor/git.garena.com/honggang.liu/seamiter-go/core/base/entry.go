package base

import (
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"github.com/pkg/errors"
	"sync"
)

type ExitHandler func(entry *SentinelEntry, ctx *EntryContext) error

type SentinelEntry struct {
	res *ResourceWrapper
	// one entry bounds with one context
	ctx *EntryContext

	exitHandlers []ExitHandler

	// each entry holds a slot chain.
	// it means this entry will go through the sc
	sc *SlotChain

	exitCtl sync.Once
}

func NewSentinelEntry(ctx *EntryContext, rw *ResourceWrapper, sc *SlotChain) *SentinelEntry {
	return &SentinelEntry{
		res:          rw,
		ctx:          ctx,
		exitHandlers: make([]ExitHandler, 0),
		sc:           sc,
	}
}

func (e *SentinelEntry) WhenExit(exitHandler ExitHandler) {
	e.exitHandlers = append(e.exitHandlers, exitHandler)
}

func (e *SentinelEntry) SetError(err error) {
	if e.ctx != nil {
		e.ctx.SetError(err)
	}
}
func (e *SentinelEntry) Context() *EntryContext {
	return e.ctx
}

func (e *SentinelEntry) Resource() *ResourceWrapper {
	return e.res
}

type ExitOptions struct {
	err error
}

type ExitOption func(*ExitOptions)

func WithError(err error) ExitOption {
	return func(opts *ExitOptions) {
		opts.err = err
	}
}

func (e *SentinelEntry) Exit(exitOps ...ExitOption) {
	var options = ExitOptions{
		err: nil,
	}
	for _, opt := range exitOps {
		opt(&options)
	}
	ctx := e.ctx
	if ctx == nil {
		return
	}
	if options.err != nil {
		ctx.SetError(options.err)
	}
	e.exitCtl.Do(func() {
		defer func() {
			if err := recover(); err != nil {
				logging.Error(errors.Errorf("%+v", err), "Sentinel internal panic in SentinelEntry.Exit()")
			}
			if e.sc != nil {
				e.sc.RefurbishContext(ctx)
			}
		}()
		for _, handler := range e.exitHandlers {
			if err := handler(e, ctx); err != nil {
				logging.Error(err, "Fail to execute exitHandler in SentinelEntry.Exit()", "resource", e.Resource().Name())
			}
		}
		if e.sc != nil {
			e.sc.exit(ctx)
		}
	})
}
