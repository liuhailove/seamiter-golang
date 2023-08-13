package handler

import (
	"errors"
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/core/config"
	"git.garena.com/honggang.liu/seamiter-go/core/log/metric"
	"git.garena.com/honggang.liu/seamiter-go/core/system_metric"
	"git.garena.com/honggang.liu/seamiter-go/transport/common/command"
	"git.garena.com/honggang.liu/seamiter-go/util"
	"strconv"
	"strings"
	"sync"
)

var (
	sendMetricCommandHandlerInst = new(sendMetricCommandHandler)
)

func init() {
	sendMetricCommandHandlerInst.mux = new(sync.RWMutex)
	command.RegisterHandler(sendMetricCommandHandlerInst.Name(), sendMetricCommandHandlerInst)
}

type sendMetricCommandHandler struct {
	searcher metric.MetricSearcher
	mux      *sync.RWMutex
}

func (s sendMetricCommandHandler) Name() string {
	return "metric"
}

func (s sendMetricCommandHandler) Desc() string {
	return "get and aggregate metrics, accept param:startTime={startTime}&endTime={endTime}&maxLines={maxLines}&identify={resourceName}"
}

func (s sendMetricCommandHandler) Handle(request command.Request) *command.Response {
	if s.searcher == nil {
		s.mux.Lock()
		defer s.mux.Unlock()
		var err error
		if s.searcher == nil {
			s.searcher, err = metric.NewDefaultMetricSearcher(metric.GetLogBaseDir(), metric.FormMetricFileName(config.AppName(), config.LogUsePid()))
			if err != nil {
				return command.OfFailure(errors.New("Error when retrieving metrics" + err.Error()))
			}
		}
	}
	var startTimeStr = request.GetParam("startTime")
	var endTimeStr = request.GetParam("endTime")
	var maxLinesStr = request.GetParam("maxLines")
	var identity = request.GetParam("identity")
	var startTime uint64 = 0
	var maxLines uint32 = 6000
	if startTimeStr != "" {
		startTime, _ = strconv.ParseUint(startTimeStr, 10, 64)
	} else {
		return command.OfSuccess("")
	}
	var metricItemArr []*base.MetricItem
	var err error
	// Find by end time if set.
	if strings.TrimSpace(endTimeStr) != "" {
		var endTime, err = strconv.ParseUint(endTimeStr, 10, 64)
		if err != nil {
			return command.OfFailure(errors.New("Error when retrieving metrics" + err.Error()))
		}
		metricItemArr, err = s.searcher.FindByTimeAndResource(startTime, endTime, identity)
		if err != nil {
			return command.OfFailure(errors.New("Error when retrieving metrics" + err.Error()))
		}
	} else {
		var maxLines64 uint64
		if strings.TrimSpace(maxLinesStr) != "" {
			maxLines64, _ = strconv.ParseUint(maxLinesStr, 10, 32)
		}
		maxLines = uint32(maxLines64)
		maxLines = min(maxLines, 12000)
		metricItemArr, err = s.searcher.FindFromTimeWithMaxLines(startTime, maxLines)
		if err != nil {
			return command.OfFailure(errors.New("Error when retrieving metrics" + err.Error()))
		}
	}
	if metricItemArr == nil || len(metricItemArr) == 0 {
		metricItemArr = make([]*base.MetricItem, 0)
	}
	builder := strings.Builder{}
	for _, item := range metricItemArr {
		thinStr, _ := item.ToThinString()
		builder.WriteString(thinStr)
		builder.WriteString("\n")
	}
	return command.OfSuccess(builder.String())
}

// min 求最小
func min(x, y uint32) uint32 {
	if x < y {
		return x
	}
	return y
}

// addCpuUsageAndLoad 增加CPU使用率和负载
func addCpuUsageAndLoad(metricItemArr []*base.MetricItem) []*base.MetricItem {
	time := util.CurrentTimeMillis() / 1000 * 1000
	load := system_metric.CurrentLoad()
	usage := system_metric.CurrentCpuUsage()
	if load > 0 {
		var loadNode = toNode(load, time, base.SystemLoadResourceName)
		metricItemArr = append(metricItemArr, loadNode)
	}
	if usage > 0 {
		var usageNode = toNode(usage, time, base.CpuUsageResourceName)
		metricItemArr = append(metricItemArr, usageNode)
	}
	return metricItemArr
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
