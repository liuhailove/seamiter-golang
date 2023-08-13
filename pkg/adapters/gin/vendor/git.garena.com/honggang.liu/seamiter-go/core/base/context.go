package base

import (
	"git.garena.com/honggang.liu/seamiter-go/util"
)

type EntryContext struct {
	entry *SeaEntry
	// internal error when sea Entry or
	// biz error of downstream
	err error
	// Use to calculate RT
	startTime uint64
	// the rt of this transaction
	rt uint64

	Resource *ResourceWrapper
	StatNode StatNode

	// 输入
	Input *seaInput
	// 输出
	Output *seaOutput
	// the result of rule slots check
	RuleCheckResult *TokenResult
	// reserve for storing some intermediate data from the Entry execution process
	Data map[interface{}]interface{}
}

func (ctx *EntryContext) SetEntry(entry *SeaEntry) {
	ctx.entry = entry
}

func (ctx *EntryContext) Entry() *SeaEntry {
	return ctx.entry
}

func (ctx *EntryContext) SetError(err error) {
	ctx.err = err
}

func (ctx *EntryContext) Err() error {
	return ctx.err
}

func (ctx *EntryContext) StartTime() uint64 {
	return ctx.startTime
}

func (ctx *EntryContext) IsBlocked() bool {
	if ctx.RuleCheckResult == nil {
		return false
	}
	return ctx.RuleCheckResult.IsBlocked()
}

func (ctx *EntryContext) PutRt(rt uint64) {
	ctx.rt = rt
}

func (ctx *EntryContext) Rt() uint64 {
	if ctx.rt == 0 {
		rt := util.CurrentTimeMillis() - ctx.StartTime()
		return rt
	}
	return ctx.rt
}

func NewEmptyEntryContext() *EntryContext {
	return &EntryContext{}
}

// seaInput The input data of sea
type seaInput struct {
	BatchCount uint32
	Flag       int32
	Args       []interface{}
	// Headers 请求头，主要为了适配web
	Headers map[string][]string
	// MetaData 主要为了适配micro
	MetaData map[string]string
	// store some values in this context when calling context in slot.
	Attachments map[interface{}]interface{}
}

func (i *seaInput) reset() {
	i.BatchCount = 1
	i.Flag = 0
	if len(i.Args) != 0 {
		i.Args = make([]interface{}, 0)
	}
	if len(i.Headers) != 0 {
		i.Headers = make(map[string][]string, 0)
	}
	if len(i.MetaData) != 0 {
		i.MetaData = make(map[string]string, 0)
	}
	if len(i.Attachments) != 0 {
		i.Attachments = map[interface{}]interface{}{}
	}
}

// seaOutput The output data of sea
type seaOutput struct {
	// 存储方法的输出
	Rsps []interface{}
}

func (i *seaOutput) reset() {
	if len(i.Rsps) != 0 {
		i.Rsps = make([]interface{}, 0)
	}
}

// Reset init EntryContext,
func (ctx *EntryContext) Reset() {
	// reset all fields of ctx
	ctx.entry = nil
	ctx.err = nil
	ctx.startTime = 0
	ctx.rt = 0
	ctx.Resource = nil
	ctx.StatNode = nil
	ctx.Input.reset()
	ctx.Output.reset()
	if ctx.RuleCheckResult == nil {
		ctx.RuleCheckResult = NewTokenResultPass()
	} else {
		ctx.RuleCheckResult.ResetToPass()
	}
	if len(ctx.Data) != 0 {
		ctx.Data = make(map[interface{}]interface{})
	}
}
