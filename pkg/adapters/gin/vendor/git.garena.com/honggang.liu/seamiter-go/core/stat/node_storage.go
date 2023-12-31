package stat

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"sync"
)

type ResourceNodeMap map[string]*ResourceNode

var (
	inboundNode = NewResourceNode(base.TotalInBoundResourceName, base.ResTypeCommon)
	cpuNode     = NewResourceNode(base.SystemLoadResourceName, base.ResTypeCommon)
	loadNode    = NewResourceNode(base.CpuUsageResourceName, base.ResTypeCommon)
	resNodeMap  = make(ResourceNodeMap)
	rnsMux      = new(sync.RWMutex)
)

// InboundNode returns the global inbound statistic node.
func InboundNode() *ResourceNode {
	return inboundNode
}

// CpuNode returns the global cpu statistic node.
func CpuNode() *ResourceNode {
	return cpuNode
}

// LoadNode returns the global load statistic node.
func LoadNode() *ResourceNode {
	return loadNode
}

// ResourceNodeList returns the slice of all existing resource nodes.
func ResourceNodeList() []*ResourceNode {
	rnsMux.RLock()
	defer rnsMux.RUnlock()

	list := make([]*ResourceNode, 0, len(resNodeMap))
	for _, v := range resNodeMap {
		list = append(list, v)
	}
	return list
}

func GetResourceNode(resource string) *ResourceNode {
	rnsMux.RLock()
	defer rnsMux.RUnlock()
	return resNodeMap[resource]
}

func GetOrCreateResourceNode(resource string, resourceType base.ResourceType) *ResourceNode {
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

	if len(resNodeMap) >= int(base.DefaultMaxResourceAmount) {
		logging.Warn("[GetOrCreateResourceNode] Resource amount exceeds the threshold", "maxResourceAmount", base.DefaultMaxResourceAmount)
	}
	node = NewResourceNode(resource, resourceType)
	resNodeMap[resource] = node
	return node
}

func ResetResourceNodeMap() {
	rnsMux.Lock()
	defer rnsMux.Unlock()
	resNodeMap = make(ResourceNodeMap)
}
