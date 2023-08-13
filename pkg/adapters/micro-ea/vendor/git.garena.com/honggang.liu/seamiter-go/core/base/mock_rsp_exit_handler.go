package base

import (
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/util"
	"sort"
	"sync"
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

type ResourceNodeMap map[string]*ResourceNode

var (
	resNodeMap = make(ResourceNodeMap)
	// resUpdateMap 资源最近的更新时间
	resUpdateMap = sync.Map{}
	rnsMux       = new(sync.RWMutex)
)

// ResourceNodeList returns the slice of all existing resource nodes.
func ResourceNodeList() []*ResourceNode {
	rnsMux.RLock()
	defer rnsMux.RUnlock()
	// 排序
	var keys []string
	for k := range resNodeMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	list := make([]*ResourceNode, 0, len(resNodeMap))
	for _, k := range keys {
		list = append(list, resNodeMap[k])
	}
	return list
}

func GetResourceNode(resource string) *ResourceNode {
	rnsMux.RLock()
	defer rnsMux.RUnlock()
	return resNodeMap[resource]
}

// RemoveResourceNode 移除资源节点
func RemoveResourceNode(resource string) {
	rnsMux.RLock()
	defer rnsMux.RUnlock()
	delete(resNodeMap, resource)
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
	rnsMux.Lock()
	defer rnsMux.Unlock()
	node = resNodeMap[resource]
	if node != nil {
		return node
	}
	if len(resNodeMap) >= int(DefaultMaxResourceAmount) {
		logging.Warn("[GetOrCreateResourceNode] Resource amount exceeds the threshold", "maxResourceAmount", DefaultMaxResourceAmount)
	}
	node = NewResourceNode(resource, resourceType, rsp)
	resNodeMap[resource] = node
	// 设置更新时间
	resUpdateMap.Store(resource, util.CurrentTimeMillis())
	return node
}

func ResetResourceNodeMap() {
	rnsMux.Lock()
	defer rnsMux.Unlock()
	resNodeMap = make(ResourceNodeMap)
}
