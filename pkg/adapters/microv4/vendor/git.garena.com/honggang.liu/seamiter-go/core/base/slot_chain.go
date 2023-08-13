package base

import (
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/util"
	"github.com/pkg/errors"
	"sort"
	"sync"
)

type BaseSlot interface {
	// Order returns the sort value of the slot.
	// SlotChain will sort all it's slots by ascending sort value in each bucket
	// (StatPrepareSlot bucket、RuleCheckSlot bucket and StatSlot bucket)
	Order() uint32

	// Initial
	//
	// 初始化，如果有初始化工作放入其中
	Initial()
}

// StatPrepareSlot is responsible for some preparation before statistic
// For example: init structure and so on
type StatPrepareSlot interface {
	BaseSlot
	// Prepare function do some initialization
	// Such as: init statistic structure、node and etc
	// The result of preparing would store in EntryContext
	// All StatPrepareSlots execute in sequence
	// Prepare function should not throw panic.
	Prepare(ctx *EntryContext)
}

// RuleCheckSlot is rule based checking strategy
// All checking rule must implement this interface.
type RuleCheckSlot interface {
	BaseSlot

	// Check function do some validation
	// It can break off the slot pipeline
	// Each TokenResult will return check result
	// The upper logic will control pipeline according to SlotResult.
	Check(ctx *EntryContext) *TokenResult
}

// RouterSlot 流量路由策略
// 所有的路由规则都必须实现这个接口.
type RouterSlot interface {
	BaseSlot

	// Router Check function do some validation
	// It can break off the slot pipeline
	// Each TokenResult will return check result
	// The upper logic will control pipeline according to SlotResult.
	Router(ctx *EntryContext) *TokenResult
}

// StatSlot is responsible for counting all custom biz metrics.
// StatSlot would not handle any panic, and pass up all panic to slot chain
type StatSlot interface {
	BaseSlot
	// OnEntryPassed function will be invoked when StatPrepareSlots and RuleCheckSlots execute pass
	// StatSlots will do some statistic logic, such as QPS、log、etc
	OnEntryPassed(ctx *EntryContext)
	// OnEntryBlocked function will be invoked when StatPrepareSlots and RuleCheckSlots fail to execute
	// It may be inbound flow control or outbound cir
	// StatSlots will do some statistic logic, such as QPS、log、etc
	// blockError introduce the block detail
	OnEntryBlocked(ctx *EntryContext, blockError *BlockError)
	// OnCompleted function will be invoked when chain exits.
	// The semantics of OnCompleted is the entry passed and completed
	// Note: blocked entry will not call this function
	OnCompleted(ctx *EntryContext)
}

// SlotChain hold all system slots and customized slot.
// SlotChain support plug-in slots developed by developer.
type SlotChain struct {
	// statPres is in ascending order by StatPrepareSlot.Order() value.
	statPres []StatPrepareSlot
	// ruleChecks is in ascending order by RuleCheckSlot.Order() value.
	ruleChecks []RuleCheckSlot
	// stats is in ascending order by StatSlot.Order() value.
	stats []StatSlot

	// 流量路由策略
	routers []RouterSlot

	// EntryContext Pool, used for reuse EntryContext object
	ctxPool *sync.Pool
}

var (
	ctxPool = &sync.Pool{
		New: func() interface{} {
			ctx := NewEmptyEntryContext()
			ctx.RuleCheckResult = NewTokenResultPass()
			ctx.Data = make(map[interface{}]interface{})
			ctx.Input = &seaInput{
				BatchCount:  1,
				Flag:        0,
				Args:        make([]interface{}, 0),
				Attachments: make(map[interface{}]interface{}),
				Headers:     make(map[string][]string),
				Cookies:     make(map[string][]string),
				Body:        make(map[string][]string),
				MetaData:    make(map[string]string, 0),
			}
			ctx.Output = &seaOutput{Rsps: make([]interface{}, 0)}
			return ctx
		},
	}
)

func NewSlotChain() *SlotChain {
	return &SlotChain{
		statPres:   make([]StatPrepareSlot, 0, 8),
		ruleChecks: make([]RuleCheckSlot, 0, 8),
		stats:      make([]StatSlot, 0, 8),
		ctxPool:    ctxPool,
	}
}

// GetPooledContext Get a EntryContext from EntryContext ctxPool, if ctxPool doesn't have enough EntryContext then new one.
func (sc *SlotChain) GetPooledContext() *EntryContext {
	ctx := sc.ctxPool.Get().(*EntryContext)
	ctx.startTime = util.CurrentTimeMillis()
	return ctx
}

func (sc *SlotChain) RefurbishContext(c *EntryContext) {
	if c != nil {
		c.Reset()
		sc.ctxPool.Put(c)
	}
}

// AddStatPrepareSlot adds the StatPrepareSlot slot to the StatPrepareSlot list of the SlotChain.
// All StatPrepareSlot in the list will be sorted according to StatPrepareSlot.Order() in ascending order.
// AddStatPrepareSlot is non-thread safe,
// In concurrency scenario, AddStatPrepareSlot must be guarded by SlotChain.RWMutex#Lock
func (sc *SlotChain) AddStatPrepareSlot(s StatPrepareSlot) {
	sc.statPres = append(sc.statPres, s)
	sort.SliceStable(sc.statPres, func(i, j int) bool {
		return sc.statPres[i].Order() < sc.statPres[j].Order()
	})
}

// AddRuleCheckSlot adds the RuleCheckSlot to the RuleCheckSlot list of the SlotChain.
// All RuleCheckSlot in the list will be sorted according to RuleCheckSlot.Order() in ascending order.
// AddRuleCheckSlot is non-thread safe,
// In concurrency scenario, AddRuleCheckSlot must be guarded by SlotChain.RWMutex#Lock
func (sc *SlotChain) AddRuleCheckSlot(s RuleCheckSlot) {

	// 初始化
	s.Initial()
	sc.ruleChecks = append(sc.ruleChecks, s)
	sort.SliceStable(sc.ruleChecks, func(i, j int) bool {
		return sc.ruleChecks[i].Order() < sc.ruleChecks[j].Order()
	})
}

// AddStatSlot adds the StatSlot to the StatSlot list of the SlotChain.
// All StatSlot in the list will be sorted according to StatSlot.Order() in ascending order.
// AddStatSlot is non-thread safe,
// In concurrency scenario, AddStatSlot must be guarded by SlotChain.RWMutex#Lock
func (sc *SlotChain) AddStatSlot(s StatSlot) {
	sc.stats = append(sc.stats, s)
	sort.SliceStable(sc.stats, func(i, j int) bool {
		return sc.stats[i].Order() < sc.stats[j].Order()
	})
}

// AddRouterSlot adds the RouterSlot to the StatSlot list of the SlotChain.
// All RouterSlot in the list will be sorted according to RouterSlot.Order() in ascending order.
// AdRouterSlot is non-thread safe,
// In concurrency scenario, AddStatSlot must be guarded by SlotChain.RWMutex#Lock
func (sc *SlotChain) AddRouterSlot(s RouterSlot) {
	sc.routers = append(sc.routers, s)
	sort.SliceStable(sc.routers, func(i, j int) bool {
		return sc.routers[i].Order() < sc.routers[j].Order()
	})
}

// Entry The entrance of slot chain
// Return the TokenResult and nil if internal panic.
func (sc *SlotChain) Entry(ctx *EntryContext) *TokenResult {
	// This should not happen, unless there are errors existing in sea internal.
	// If happened, need to add TokenResult in EntryContext
	defer func() {
		if err := recover(); err != nil {
			logging.Error(errors.Errorf("%+v", err), "sea internal panic in SlotChain.Entry()")
			ctx.SetError(errors.Errorf("%+v", err))
			return
		}
	}()

	// execute prepare slot

	sps := sc.statPres
	if len(sps) > 0 {
		for _, s := range sps {
			s.Prepare(ctx)
		}
	}

	// execute rule based checking slot
	rcs := sc.ruleChecks
	var ruleCheckRet *TokenResult
	if len(rcs) > 0 {
		for _, s := range rcs {
			sr := s.Check(ctx)
			if sr == nil {
				// nil equals to check pass
				continue
			}
			// check slot result
			if sr.IsBlocked() {
				ruleCheckRet = sr
				break
			}
		}
	}
	if ruleCheckRet == nil {
		ctx.RuleCheckResult.ResetToPass()
	} else {
		ctx.RuleCheckResult = ruleCheckRet
	}

	// 路由Check
	rs := sc.routers
	if len(rs) > 0 {
		// 目前只有一个
		for _, s := range rs {
			var routerResult = s.Router(ctx)
			if routerResult == nil {
				continue
			}
			// 检查是否发生阻塞
			if routerResult.IsBlocked() {
				ctx.RuleCheckResult.blockErr = routerResult.blockErr
				ctx.RuleCheckResult.status = ResultStatusBlocked
				break
			}
			ctx.RuleCheckResult.grayRes = routerResult.grayRes
			ctx.RuleCheckResult.grayTag = routerResult.grayTag
			ctx.RuleCheckResult.linkPass = routerResult.linkPass
			ctx.RuleCheckResult.grayAddress = routerResult.grayAddress
		}
	}

	// execute statistic slot
	ss := sc.stats
	ruleCheckRet = ctx.RuleCheckResult
	if len(ss) > 0 {
		for _, s := range ss {
			// indicate the result of rule based checking slot.
			if !ruleCheckRet.IsBlocked() {
				s.OnEntryPassed(ctx)
			} else {
				// The block error should not be nil.
				s.OnEntryBlocked(ctx, ruleCheckRet.blockErr)
			}
		}
	}
	return ruleCheckRet
}

func (sc *SlotChain) exit(ctx *EntryContext) {
	if ctx == nil || ctx.Entry() == nil {
		logging.Error(errors.New("entryContext or seantry is nil"),
			"EntryContext or SeaEntry is nil in SlotChain.exit()", "ctx", ctx)
		return
	}
	// The OnCompleted is called only when entry passed
	if ctx.IsBlocked() {
		return
	}
	for _, s := range sc.stats {
		s.OnCompleted(ctx)
	}
	// relieve the context here
}
