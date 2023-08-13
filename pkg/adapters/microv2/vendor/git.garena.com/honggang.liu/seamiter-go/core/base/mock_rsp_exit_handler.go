package base

import (
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"sort"
	"sync"
)

func MockRspExitHandler(entry *SeaEntry, ctx *EntryContext) error {

	outputs := ctx.Output.Rsps
	if len(outputs) == 0 {
		return nil
	}
	// 去除限制
	// 目前仅适配go-micro
	//if ctx.Resource.Classification() != ResTypeMicro {
	//	return nil
	//}
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
	rnsMux     = new(sync.RWMutex)
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

func GetOrCreateResourceNode(resource string, resourceType ResourceType, rsp interface{}) *ResourceNode {
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
	return node
}

func ResetResourceNodeMap() {
	rnsMux.Lock()
	defer rnsMux.Unlock()
	resNodeMap = make(ResourceNodeMap)
}
