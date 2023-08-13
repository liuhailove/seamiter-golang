package datasource

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"github.com/go-basic/uuid"
)

const (
	RootId = "machine-root"
)

type Node struct {
	Id       string `json:"id"`
	ParentId string `json:"parentId"`
	Resource string `json:"resource"` // 资源ID

	ThreadNum          uint32 `json:"threadNum"`
	PassQps            uint64 `json:"passQps"`
	BlockQps           uint64 `json:"blockQps"`
	TotalQqs           uint64 `json:"totalQqs"`
	AverageRt          uint64 `json:"averageRt"`
	SuccessQps         uint64 `json:"successQps"`
	ExceptionQps       uint64 `json:"exceptionQps"`
	OneMinutePass      uint64 `json:"oneMinutePass"`
	OneMinuteBlock     uint64 `json:"oneMinuteBlock"`
	OneMinuteException uint64 `json:"oneMinuteException"`
	OneMinuteTotal     uint64 `json:"oneMinuteTotal"`

	Timestamp uint64 `json:"timestamp"`
}

func transToNode(metricItems []*base.MetricItem) []*Node {
	var nodes = make([]*Node, 0)
	if metricItems == nil || len(metricItems) == 0 {
		return nodes
	}
	var rootNode = new(Node)
	rootNode.Id = uuid.New()
	rootNode.Resource = RootId
	nodes = append(nodes, rootNode)
	for _, metricIt := range metricItems {
		var node = new(Node)
		node.Id = uuid.New()
		node.Resource = metricIt.Resource
		node.ThreadNum = metricIt.Concurrency
		node.PassQps = metricIt.PassQps
		node.BlockQps = metricIt.BlockQps
		node.TotalQqs = metricIt.CompleteQps + metricIt.BlockQps
		node.AverageRt = metricIt.AvgRt
		// TODO 一致性确认
		node.SuccessQps = metricIt.PassQps
		node.ExceptionQps = metricIt.ErrorQps
		node.OneMinutePass = 0
		node.OneMinuteBlock = 0
		node.OneMinuteException = 0
		node.OneMinuteTotal = 0
		node.Timestamp = metricIt.Timestamp
		node.ParentId = rootNode.Id
		nodes = append(nodes, node)
	}
	return nodes
}
