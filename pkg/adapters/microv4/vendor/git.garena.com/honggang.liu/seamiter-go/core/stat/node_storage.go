package stat

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"sync"
	"sync/atomic"
)

var (
	inboundNode = NewResourceNode(base.TotalInBoundResourceName, base.ResTypeCommon)
	cpuNode     = NewResourceNode(base.SystemLoadResourceName, base.ResTypeCommon)
	loadNode    = NewResourceNode(base.CpuUsageResourceName, base.ResTypeCommon)
	resNodeMap  = sync.Map{}
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
	list := make([]*ResourceNode, 0)
	resNodeMap.Range(func(key, value interface{}) bool {
		list = append(list, value.(*ResourceNode))
		return true
	})
	return list
}

func GetResourceNode(resource string) *ResourceNode {
	val, ok := resNodeMap.Load(resource)
	if ok {
		return val.(*ResourceNode)
	}
	return nil
}

func GetResourceNodeLen() uint32 {
	var count uint32 = 0
	resNodeMap.Range(func(key, value interface{}) bool {
		atomic.AddUint32(&count, 1)
		return true
	})
	return count
}

func GetOrCreateResourceNode(resource string, resourceType base.ResourceType) *ResourceNode {
	node := GetResourceNode(resource)
	if node != nil {
		return node
	}
	if GetResourceNodeLen() >= base.DefaultMaxResourceAmount {
		logging.Warn("[GetOrCreateResourceNode] Resource amount exceeds the threshold", "maxResourceAmount", base.DefaultMaxResourceAmount)
	}
	node = NewResourceNode(resource, resourceType)
	resNodeMap.Store(resource, node)
	return node
}

func ResetResourceNodeMap() {
	resNodeMap = sync.Map{}
}
