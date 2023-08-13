package base

//
//import (
//	"git.garena.com/honggang.liu/seamiter-go/logging"
//	"sync"
//)
//
///**
// * <p>
// * A {@link Node} used to hold statistics for specific resource name in the specific context.
// * Each distinct resource in each distinct {@link Context} will corresponding to a {@link DefaultNode}.
// * </p>
// * <p>
// * This class may have a list of sub {@link DefaultNode}s. Child nodes will be created when
// * calling {@link SphU}#entry() or {@link SphO}@entry() multiple times in the same {@link Context}.
// * </p>
// *
// * @author honggang.liu
// */
//
//type DefaultNode struct {
//	Resource *ResourceWrapper
//	StatNode StatNode
//
//	// The list of all child nodes
//	childList []StatNode
//
//	// 锁
//	rnsMux sync.RWMutex
//}
//
//func NewDefaultNode(resource *ResourceWrapper, clusterNode StatNode) *DefaultNode {
//	node := new(DefaultNode)
//	node.Resource = resource
//	node.StatNode = clusterNode
//	return node
//}
//
//func (d *DefaultNode) GetClusterNode() StatNode {
//	return d.StatNode
//}
//
//func (d *DefaultNode) SetClusterNode(clusterNode StatNode) {
//	d.StatNode = clusterNode
//}
//
//// AddChild
//// Add child node to current node.
//// @param node valid child node
//func (d *DefaultNode) AddChild(node StatNode) {
//	if node == nil {
//		logging.Warn("[AddChild] Trying to add null child to node, ignored", "name", d.Resource.Name())
//		return
//	}
//	if !contains(d.childList, node) {
//		d.rnsMux.Lock()
//		defer d.rnsMux.Unlock()
//		if !contains(d.childList, node) {
//			list := make([]StatNode, 0, len(d.childList)+1)
//			for _, v := range d.childList {
//				list = append(d.childList, v)
//			}
//			list = append(d.childList, node)
//			d.childList = list
//		}
//		logging.Info("[AddChild] Add child to node", "resourceName", node, "nodeName", d.Resource.Name())
//	}
//}
//
//func (d *DefaultNode) RemoveChildList() {
//	d.childList = make([]StatNode, 0)
//}
//
//func (d *DefaultNode) GetChildList() []StatNode {
//	return d.childList
//}
//
//func (d *DefaultNode) IncreaseBlockQps(count int64) {
//	d.StatNode.AddCount(MetricEventBlock, count)
//}
//
//func (d *DefaultNode) IncreaseExceptionQps(count int64) {
//	d.StatNode.AddCount(MetricEventError, count)
//}
//
//func (d *DefaultNode) AddRtAndSuccess(rt int64, successCount int64) {
//	//d.AddRtAndSuccess()
//}
//
//// contains 判断是否包含
//func contains(arr []StatNode, node StatNode) bool {
//	if arr == nil {
//		return false
//	}
//	for _, element := range arr {
//		if node == element {
//			return true
//		}
//	}
//	return false
//}
