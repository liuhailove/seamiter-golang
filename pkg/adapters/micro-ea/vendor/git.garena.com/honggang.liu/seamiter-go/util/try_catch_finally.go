package util

import (
	"errors"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"reflect"
)

type CatchHandler interface {
	Catch(e error, handler func(err error)) CatchHandler
	CatchAll(handler func(err error)) FinalHandler
	FinalHandler
}

type FinalHandler interface {
	Finally(handlers ...func())
}

func Try(f func()) CatchHandler {
	t := &catchHandler{}
	defer func() {
		defer func() {
			r := recover()
			if r != nil {
				if e, ok := r.(error); ok {
					t.err = e
				} else if es, ok := r.(string); ok {
					t.err = errors.New(es)
				} else {
					logging.Warn("CatchHandler cannot trans", "msg", r)
				}
			}
		}()
		f()
	}()
	return t
}

type catchHandler struct {
	err      error
	hasCatch bool
}

func (t *catchHandler) RequireCatch() bool { //<1>判断是否有必要执行catch块，true为需要执行，false为不执行
	if t.hasCatch { //<2>如果已经执行了catch块，就直接判断不执行
		return false
	}
	if t.err == nil { //<3>如果异常为空，则判断不执行
		return false
	}
	return true
}

func (t *catchHandler) Catch(e error, handler func(err error)) CatchHandler {
	if !t.RequireCatch() {
		return t
	}
	//<4>如果传入的error类型和发生异常的类型一致，则执行异常处理器，并将hasCatch修改为true代表已捕捉异常
	if reflect.TypeOf(e) == reflect.TypeOf(t.err) {
		handler(t.err)
		t.hasCatch = true
	}
	return t
}

func (t *catchHandler) CatchAll(handler func(err error)) FinalHandler {
	//<5>CatchAll()函数和Catch()函数都是返回同一个对象，但返回的接口类型却不一样，也就是CatchAll()之后只能调用Finally()
	if !t.RequireCatch() {
		return t
	}
	handler(t.err)
	t.hasCatch = true
	return t
}

func (t *catchHandler) Finally(handlers ...func()) {
	//<6>遍历处理器，并在Finally函数执行完毕之后执行
	for _, handler := range handlers {
		defer handler()
	}
	err := t.err
	//<7>如果异常不为空，且未捕捉异常，则抛出异常
	if err != nil && !t.hasCatch {
		panic(err)
	}
}
