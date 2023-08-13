package vo

import (
	"github.com/go-basic/uuid"
	"github.com/liuhailove/seamiter-golang/core/base"
	"github.com/liuhailove/seamiter-golang/core/stat"
)

// NodeVo This class is view object of DefaultNode or ClusterNode.
type NodeVo struct {
	id       string
	parentId string
	resource string

	threadNum          int32
	passQps            int64
	blockQps           int64
	totalQps           int64
	averageRt          int64
	successQps         int64
	exceptionQps       int64
	oneMinutePass      int64
	oneMinuteBlock     int64
	oneMinuteException int64
	oneMinuteTotal     int64

	timestamp int64
}

func FromDefaultNode(node *stat.ResourceNode, parentId string) *NodeVo {
	if node == nil {
		return nil
	}
	var vo = new(NodeVo)
	vo.id = uuid.New()
	vo.parentId = parentId
	vo.resource = node.ResourceName()
	vo.threadNum = node.CurrentConcurrency()
	// TODO
	//vo.passQps = 1
	//vo.blockQps = (long) node.blockQps();
	//vo.totalQps = (long) node.totalQps();
	//vo.averageRt = (long) node.avgRt();
	//vo.successQps = (long) node.successQps();
	//vo.exceptionQps = (long) node.exceptionQps();
	//vo.oneMinuteException = node.totalException();
	//vo.oneMinutePass = node.totalRequest() - node.blockRequest();
	//vo.oneMinuteBlock = node.blockRequest();
	//vo.oneMinuteTotal = node.totalRequest();
	//vo.timestamp = System.currentTimeMillis();
	return vo
}

func FromClusterNode(name base.ResourceWrapper, node *stat.ResourceNode) *NodeVo {
	return FromClusterNodeBy(name.Name(), node)
}

func FromClusterNodeBy(name string, node *stat.ResourceNode) *NodeVo {
	//if (node == null) {
	//	return null;
	//}
	//NodeVo vo = new NodeVo();
	//vo.resource = name;
	//vo.threadNum = node.curThreadNum();
	//vo.passQps = (long) node.passQps();
	//vo.blockQps = (long) node.blockQps();
	//vo.totalQps = (long) node.totalQps();
	//vo.averageRt = (long) node.avgRt();
	//vo.successQps = (long) node.successQps();
	//vo.exceptionQps = (long) node.exceptionQps();
	//vo.oneMinuteException = node.totalException();
	//vo.oneMinutePass = node.totalRequest() - node.blockRequest();
	//vo.oneMinuteBlock = node.blockRequest();
	//vo.oneMinuteTotal = node.totalRequest();
	//vo.timestamp = System.currentTimeMillis();
	//return vo;
	return nil
}

func (v NodeVo) GetId() string {
	return v.id
}

func (v NodeVo) SetId(id string) {
	v.id = id
}

func (v NodeVo) GetParentId() string {
	return v.parentId
}

func (v NodeVo) SetParentId(parentId string) {
	v.parentId = parentId
}

func (v NodeVo) GetResource() string {
	return v.resource
}

func (v NodeVo) SetResource(resource string) {
	v.resource = resource
}

func (v NodeVo) GetThreadNum() int32 {
	return v.threadNum
}

func (v NodeVo) SetThreadNum(threadNum int32) {
	v.threadNum = threadNum
}

func (v NodeVo) GetPassQps() int64 {
	return v.passQps
}

func (v NodeVo) SetPassQps(passQps int64) {
	v.passQps = passQps
}

func (v NodeVo) GetBlockQps() int64 {
	return v.blockQps
}
func (v NodeVo) SetBlockQps(blockQps int64) {
	v.blockQps = blockQps
}

func (v NodeVo) GetTotalQps() int64 {
	return v.totalQps
}

func (v NodeVo) SetTotalQps(totalQps int64) {
	v.totalQps = totalQps
}

func (v NodeVo) GetAverageRt() int64 {
	return v.averageRt
}

func (v NodeVo) SetAverageRt(averageRt int64) {
	v.averageRt = averageRt
}

func (v NodeVo) GetSuccessQps() int64 {
	return v.successQps
}

func (v NodeVo) SetSuccessQps(successQps int64) {
	v.successQps = successQps
}

func (v NodeVo) GetExceptionQps() int64 {
	return v.exceptionQps
}

func (v NodeVo) SetExceptionQps(exceptionQps int64) {
	v.exceptionQps = exceptionQps
}

func (v NodeVo) GetOneMinuteException() int64 {
	return v.oneMinuteException
}

func (v NodeVo) SetOneMinuteException(oneMinuteException int64) {
	v.oneMinuteException = oneMinuteException
}

func (v NodeVo) GetOneMinutePass() int64 {
	return v.oneMinutePass
}

func (v NodeVo) SetOneMinutePass(oneMinutePass int64) {
	v.oneMinutePass = oneMinutePass
}

func (v NodeVo) GetOneMinuteBlock() int64 {
	return v.oneMinuteBlock
}

func (v NodeVo) SetOneMinuteBlock(oneMinuteBlock int64) {
	v.oneMinuteBlock = oneMinuteBlock
}

func (v NodeVo) GetOneMinuteTotal() int64 {
	return v.oneMinuteTotal
}

func (v NodeVo) SetOneMinuteTotal(oneMinuteTotal int64) {
	v.oneMinuteTotal = oneMinuteTotal
}

func (v NodeVo) getTimestamp() int64 {
	return v.timestamp
}

func (v NodeVo) setTimestamp(timestamp int64) {
	v.timestamp = timestamp
}
