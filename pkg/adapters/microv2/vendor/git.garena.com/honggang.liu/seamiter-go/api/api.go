package api

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"sync"
)

var entryOptsPool = sync.Pool{New: func() interface{} {
	return &EntryOptions{
		resourceType: base.ResTypeCommon,
		entryType:    base.Outbound,
		batchCount:   1,
		flag:         0,
		slotChain:    nil,
		args:         nil,
		attachments:  nil,
		headers:      nil,
	}
}}

// EntryOptions represents the options of a sea resource entry.
type EntryOptions struct {
	resourceType base.ResourceType
	entryType    base.TrafficType
	batchCount   uint32
	flag         int32
	slotChain    *base.SlotChain
	args         []interface{}
	attachments  map[interface{}]interface{}
	rsps         []interface{}
	headers      map[string][]string
	metaData     map[string]string
}

func (o *EntryOptions) Reset() {
	o.resourceType = base.ResTypeCommon
	o.entryType = base.Outbound
	o.batchCount = 1
	o.flag = 0
	o.slotChain = nil
	o.args = nil
	o.attachments = nil
	o.rsps = nil
	o.headers = nil
	o.metaData = nil
}

type EntryOption func(options *EntryOptions)

// WithResourceType sets the resource entry with the given resource type.
func WithResourceType(resourceType base.ResourceType) EntryOption {
	return func(options *EntryOptions) {
		options.resourceType = resourceType
	}
}

// WithTrafficType sets the resource entry with the given traffic type.
func WithTrafficType(entryType base.TrafficType) EntryOption {
	return func(options *EntryOptions) {
		options.entryType = entryType
	}
}

// WithBatchCount sets the resource entry with the given batch count (by default 1).
func WithBatchCount(batchCount uint32) EntryOption {
	return func(options *EntryOptions) {
		options.batchCount = batchCount
	}
}

// WithFlag sets the resource entry with the given additional flag.
func WithFlag(flag int32) EntryOption {
	return func(options *EntryOptions) {
		options.flag = flag
	}
}

// WithArgs sets the resource entry with the given additional parameters.
func WithArgs(args ...interface{}) EntryOption {
	return func(options *EntryOptions) {
		options.args = append(options.args, args...)
	}
}

// WithRsps sets the resource entry with the given additional rsps.
func WithRsps(rsps ...interface{}) EntryOption {
	return func(options *EntryOptions) {
		options.rsps = append(options.rsps, rsps...)
	}
}

// WithSlotChain sets the slot chain.
func WithSlotChain(chain *base.SlotChain) EntryOption {
	return func(options *EntryOptions) {
		options.slotChain = chain
	}
}

// WithAttachment set the resource entry with the given k-v pair
func WithAttachment(key interface{}, value interface{}) EntryOption {
	return func(opts *EntryOptions) {
		if opts.attachments == nil {
			opts.attachments = make(map[interface{}]interface{}, 8)
		}
		opts.attachments[key] = value
	}
}

// WithAttachments set the resource entry with the given k-v pairs
func WithAttachments(data map[interface{}]interface{}) EntryOption {
	return func(options *EntryOptions) {
		if options.attachments == nil {
			options.attachments = make(map[interface{}]interface{}, len(data))
		}
		for key, value := range data {
			options.attachments[key] = value
		}
	}
}

// WithHeaders sets the resource entry with the given headers.
// mainly for adapter web mock
func WithHeaders(headers map[string][]string) EntryOption {
	return func(options *EntryOptions) {
		options.headers = headers
	}
}

// WithMetaData sets the resource entry with the given context key value.
// mainly for adapter micro
func WithMetaData(metaData map[string]string) EntryOption {
	return func(options *EntryOptions) {
		options.metaData = metaData
	}
}

// Entry is the basic API of sea.
func Entry(resource string, opts ...EntryOption) (*base.SeaEntry, *base.BlockError) {
	options := entryOptsPool.Get().(*EntryOptions)
	defer func() {
		options.Reset()
		entryOptsPool.Put(options)
	}()

	for _, opt := range opts {
		opt(options)
	}
	if options.slotChain == nil {
		options.slotChain = GlobalSlotChain()
	}
	return entry(resource, options)
}

func entry(resource string, options *EntryOptions) (*base.SeaEntry, *base.BlockError) {
	rw := base.NewResourceWrapper(resource, options.resourceType, options.entryType)
	sc := options.slotChain
	if sc == nil {
		return base.NewSeaEntry(nil, rw, nil), nil
	}

	// Get context from pool.
	ctx := sc.GetPooledContext()
	ctx.Resource = rw
	ctx.Input.BatchCount = options.batchCount
	ctx.Input.Flag = options.flag
	if len(options.args) != 0 {
		ctx.Input.Args = options.args
	}
	if len(options.attachments) != 0 {
		ctx.Input.Attachments = options.attachments
	}
	if len(options.headers) != 0 {
		ctx.Input.Headers = options.headers
	}
	if len(options.metaData) != 0 {
		ctx.Input.MetaData = options.metaData
	}
	if len(options.rsps) != 0 {
		ctx.Output.Rsps = options.rsps
	}
	e := base.NewSeaEntry(ctx, rw, sc)
	ctx.SetEntry(e)
	r := sc.Entry(ctx)
	if r == nil {
		// This indicates internal error in some slots, so just pass
		return e, nil
	}
	if r.Status() == base.ResultStatusBlocked {
		// r will be put to Pool in calling Exit()
		// must finish the lifecycle of r.
		blockErr := base.NewBlockErrorFromDeepCopy(r.BlockError())
		e.Exit()
		return nil, blockErr
	}
	return e, nil

}
