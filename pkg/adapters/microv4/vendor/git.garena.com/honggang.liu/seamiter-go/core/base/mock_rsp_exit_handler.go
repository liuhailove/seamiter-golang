package base

import (
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/util"
	"sort"
	"sync"
	"sync/atomic"
)

const (
	// resMaxAliveTimeMs 资源最大存活时间，5min，超过5分钟会更新，并重新上报到server
	resMaxAliveTimeMs uint64 = 5 * 60 * 1000
)

func MockRspExitHandler(entry *SeaEntry, ctx *EntryContext) error {
	outputs := ctx.Output.Rsps
	if len(outputs) == 0 {
		return nil
	}
	if ctx.Err() != nil {
		return nil
	}
	_ = GetOrCreateResourceNode(ctx.Resource.Name(), ctx.Resource.Classification(), outputs[0])
	return nil
}

type ResourceNode struct {
	ResourceName string       `json:"resourceName"`
	ResourceType ResourceType `json:"resourceType"`
	Rsp          interface{}  `json:"rsp"`
}

// NewResourceNode creates a new resource node with given name and classification.
func NewResourceNode(resourceName string, resourceType ResourceType, rsp interface{}) *ResourceNode {
	return &ResourceNode{
		ResourceName: resourceName,
		ResourceType: resourceType,
		Rsp:          rsp,
	}
}

var (
	resNodeMap = sync.Map{}
	// resUpdateMap 资源最近的更新时间
	resUpdateMap = sync.Map{}
)

// ResourceNodeList returns the slice of all existing resource nodes.
func ResourceNodeList() []*ResourceNode {
	// 排序
	var keys []string
	resNodeMap.Range(func(key, value interface{}) bool {
		keys = append(keys, key.(string))
		return true
	})
	sort.Strings(keys)
	list := make([]*ResourceNode, 0)
	for key := range keys {
		node, ok := resNodeMap.Load(key)
		if ok {
			list = append(list, node.(*ResourceNode))
		}
	}
	return list
}

func GetResourceNode(resource string) *ResourceNode {
	val, ok := resNodeMap.Load(resource)
	if ok {
		return val.(*ResourceNode)
	}
	return nil
}

// RemoveResourceNode 移除资源节点
func RemoveResourceNode(resource string) {
	resNodeMap.Delete(resource)
}

func GetResourceNodeLen() uint32 {
	var count uint32 = 0
	resNodeMap.Range(func(key, value interface{}) bool {
		atomic.AddUint32(&count, 1)
		return true
	})
	return count
}

func GetOrCreateResourceNode(resource string, resourceType ResourceType, rsp interface{}) *ResourceNode {
	// 获取资源最近一次更新时间
	var lastUpdateTimeMs, ok = resUpdateMap.Load(resource)
	// 如果资源已经被暂存，并且从上次更新到当前已经超过了最大存活时间，则重新缓存数据
	if ok && lastUpdateTimeMs.(uint64)+resMaxAliveTimeMs < util.CurrentTimeMillis() {
		RemoveResourceNode(resource)
	}
	node := GetResourceNode(resource)
	if node != nil {
		return node
	}
	nodeP, ok := resNodeMap.Load(resource)
	if ok {
		return nodeP.(*ResourceNode)
	}
	if GetResourceNodeLen() >= DefaultMaxResourceAmount {
		logging.Warn("[GetOrCreateResourceNode] Resource amount exceeds the threshold", "maxResourceAmount", DefaultMaxResourceAmount)
	}
	node = NewResourceNode(resource, resourceType, rsp)
	resNodeMap.Store(resource, node)
	// 设置更新时间
	resUpdateMap.Store(resource, util.CurrentTimeMillis())
	return node
}

func ResetResourceNodeMap() {
	resNodeMap = sync.Map{}
}
