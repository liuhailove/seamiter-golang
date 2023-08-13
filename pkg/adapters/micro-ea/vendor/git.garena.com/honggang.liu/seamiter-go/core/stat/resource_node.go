package stat

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/core/config"
)

type ResourceNode struct {
	BaseStatNode
	//StatNodeInMinute BaseStatNode // 分钟计数
	resourceName string
	resourceType base.ResourceType
}

// NewResourceNode creates a new resource node with given name and classification.
func NewResourceNode(resourceName string, resourceType base.ResourceType) *ResourceNode {
	return &ResourceNode{
		BaseStatNode: *NewBaseStatNode(config.MetricStatisticSampleCount(), config.MetricStatisticIntervalMs()),
		//StatNodeInMinute: *NewBaseStatNode(60, 60*1000),
		resourceName: resourceName,
		resourceType: resourceType,
	}
}

func (n *ResourceNode) ResourceType() base.ResourceType {
	return n.resourceType
}

func (n *ResourceNode) ResourceName() string {
	return n.resourceName
}
