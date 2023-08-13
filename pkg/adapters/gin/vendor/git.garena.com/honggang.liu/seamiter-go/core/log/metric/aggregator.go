package metric

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/core/config"
	"git.garena.com/honggang.liu/seamiter-go/core/stat"
	"git.garena.com/honggang.liu/seamiter-go/core/system_metric"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/util"
	"sort"
	"sync"
	"time"
)

type metricTimeMap = map[uint64][]*base.MetricItem

const (
	logFlushQueueSize = 60
)

var (
	// The timestamp of the last fetching. The time unit is ms (= second * 1000).
	lastFetchTime int64 = -1
	writeChan           = make(chan metricTimeMap, logFlushQueueSize)
	stopChan            = make(chan struct{})

	metricWriter MetricLogWriter
	initOnce     sync.Once

	// 当前map
	currentMetricTimeMap metricTimeMap = make(map[uint64][]*base.MetricItem)
)

func InitTask() (err error) {
	initOnce.Do(func() {
		flushInterval := config.MetricLogFlushIntervalSec()
		if flushInterval == 0 {
			return
		}

		metricWriter, err = NewDefaultMetricLogWriter(config.MetricLogSingleFileMaxSize(), config.MetricLogMaxFileAmount())
		if err != nil {
			logging.Error(err, "Failed to initialize the MetricLogWriter in aggregator.InitTask()")
			return
		}

		// Schedule the log flushing task
		go util.RunWithRecover(writeTaskLoop)
		// Schedule the log aggregating task
		ticker := util.NewTicker(time.Duration(flushInterval) * time.Second)
		go util.RunWithRecover(func() {
			for {
				select {
				case <-ticker.C():
					doAggregate()
				case <-stopChan:
					ticker.Stop()
					return
				}
			}
		})
	})
	return err
}

func writeTaskLoop() {
	for {
		select {
		case m := <-writeChan:
			keys := make([]uint64, 0, len(m))
			for t := range m {
				keys = append(keys, t)
			}
			// Sort the time
			sort.Slice(keys, func(i, j int) bool {
				return keys[i] < keys[j]
			})

			for _, t := range keys {
				err := metricWriter.Write(t, m[t])
				if err != nil {
					logging.Error(err, "[MetricAggregatorTask] fail tp write metric in aggregator.writeTaskLoop()")
				}
			}
		}
	}
}

func doAggregate() {
	curTime := util.CurrentTimeMillis()
	curTime = curTime - curTime%1000

	if int64(curTime) <= lastFetchTime {
		return
	}
	maps := make(metricTimeMap)
	cns := stat.ResourceNodeList()
	for _, node := range cns {
		metrics := currentMetricItems(node, curTime)
		aggregateIntoMap(maps, metrics, node)
	}
	// Aggregate for inbound entrance node.
	aggregateIntoMap(maps, currentMetricItems(stat.InboundNode(), curTime), stat.InboundNode())
	// 汇聚CPU和负载
	aggregateIntoMap(maps, currentCpuMetricItems(curTime), stat.CpuNode())
	aggregateIntoMap(maps, currentLoadMetricItems(curTime), stat.LoadNode())
	// Update current last fetch timestamp.
	lastFetchTime = int64(curTime)

	if len(maps) > 0 {
		writeChan <- maps
		currentMetricTimeMap = maps
	}
}

func aggregateIntoMap(mm metricTimeMap, metrics map[uint64]*base.MetricItem, node *stat.ResourceNode) {
	for t, item := range metrics {
		item.Resource = node.ResourceName()
		item.Classification = int32(node.ResourceType())
		items, exists := mm[t]
		if exists {
			mm[t] = append(items, item)
		} else {
			mm[t] = []*base.MetricItem{item}
		}
	}
}

func isActiveMetricItem(item *base.MetricItem) bool {
	return item.PassQps > 0 || item.BlockQps > 0 || item.CompleteQps > 0 || item.ErrorQps > 0 ||
		item.AvgRt > 0 || item.Concurrency > 0
}

func isItemTimestampInTime(ts uint64, currentSecStart uint64) bool {
	// The bucket should satisfy: windowStart between [lastFetchTime, curStart)
	return int64(ts) >= lastFetchTime && ts < currentSecStart
}

func currentMetricItems(retriever base.MetricItemRetriever, currentTime uint64) map[uint64]*base.MetricItem {
	items := retriever.MetricsOnCondition(func(ts uint64) bool {
		return isItemTimestampInTime(ts, currentTime)
	})
	m := make(map[uint64]*base.MetricItem, len(items))
	for _, item := range items {
		if !isActiveMetricItem(item) {
			continue
		}
		m[item.Timestamp] = item
	}
	return m
}

//currentCpuMetricItems 当前CPU
func currentCpuMetricItems(currentTime uint64) map[uint64]*base.MetricItem {
	m := make(map[uint64]*base.MetricItem, 1)

	// load
	load := system_metric.CurrentLoad()
	var loadMetric = toNode(load, currentTime, base.SystemLoadResourceName)
	m[loadMetric.Timestamp] = loadMetric
	return m
}

//currentLoadMetricItems 当前负载
func currentLoadMetricItems(currentTime uint64) map[uint64]*base.MetricItem {
	m := make(map[uint64]*base.MetricItem, 1)

	// cpu
	usage := system_metric.CurrentCpuUsage()
	var usageMetric = toNode(usage, currentTime, base.CpuUsageResourceName)
	m[usageMetric.Timestamp] = usageMetric
	return m
}

/**
 * transfer the value to a MetricNode, the value will multiply 10000 then truncate
 * to long value, and as the {@link MetricNode#passQps}.
 * <p>
 * This is an eclectic scheme before we have a standard metric format.
 * </p>
 *
 * @param value    value to save.
 * @param ts       timestamp
 * @param resource resource name.
 * @return a MetricNode represents the value.
 */
func toNode(value float64, ts uint64, resource string) *base.MetricItem {
	item := new(base.MetricItem)
	item.PassQps = uint64(value * 10000)
	item.Timestamp = ts
	item.Resource = resource
	return item
}

// CurrentMetricItems 获取当前时间的侧拉指标项
func CurrentMetricItems() []*base.MetricItem {
	var items = make([]*base.MetricItem, 0)
	for _, val := range currentMetricTimeMap {
		for _, item := range val {
			items = append(items, item)
		}
	}
	return items
}
