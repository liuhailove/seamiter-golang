package base

import (
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"github.com/pkg/errors"
	"sync"
)

type ExitHandler func(entry *SeaEntry, ctx *EntryContext) error

type SeaEntry struct {
	res *ResourceWrapper
	// one entry bounds with one context
	ctx *EntryContext

	exitHandlers []ExitHandler

	// each entry holds a slot chain.
	// it means this entry will go through the sc
	sc *SlotChain

	exitCtl sync.Once
}

func NewSeaEntry(ctx *EntryContext, rw *ResourceWrapper, sc *SlotChain) *SeaEntry {
	var exitHandlers = make([]ExitHandler, 0)
	exitHandlers = append(exitHandlers, MockRspExitHandler)
	return &SeaEntry{
		res:          rw,
		ctx:          ctx,
		exitHandlers: exitHandlers,
		sc:           sc,
	}
}

func (e *SeaEntry) WhenExit(exitHandler ExitHandler) {
	var exitHandlers = make([]ExitHandler, 0)
	exitHandlers = append(exitHandlers, exitHandler)
	for _, handler := range e.exitHandlers {
		exitHandlers = append(exitHandlers, handler)
	}
	e.exitHandlers = exitHandlers
}

func (e *SeaEntry) SetError(err error) {
	if e.ctx != nil {
		e.ctx.SetError(err)
	}
}
func (e *SeaEntry) Context() *EntryContext {
	return e.ctx
}

func (e *SeaEntry) Resource() *ResourceWrapper {
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

func (e *SeaEntry) Exit(exitOps ...ExitOption) {
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
				logging.Error(errors.Errorf("%+v", err), "sea internal panic in SeaEntry.Exit()")
			}
			if e.sc != nil {
				e.sc.RefurbishContext(ctx)
			}
		}()
		for _, handler := range e.exitHandlers {
			if err := handler(e, ctx); err != nil {
				logging.Error(err, "Fail to execute exitHandler in SeaEntry.Exit()", "resource", e.Resource().Name())
			}
		}
		if e.sc != nil {
			e.sc.exit(ctx)
		}
	})
}
