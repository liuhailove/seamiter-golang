package base

// global variable
const (
	TotalInBoundResourceName = "__total_inbound_traffic__"

	DefaultMaxResourceAmount uint32 = 1000
	DefaultSampleCount       uint32 = 2
	DefaultIntervalMs        uint32 = 1000

	// default 10*1000/500 = 20
	DefaultSampleCountTotal uint32 = 20

	// default 10s (total length)
	DefaultIntervalMsTotal uint32 = 10000

	DefaultStatisticMaxRt = int64(60000)

	MaxContextNameSize     = 2000
	MaxSlotChainSize       = 6000
	RootId                 = "machine-root"
	ContextDefault_Name    = "sea_default_context"
	CpuUsageResourceName   = "__cpu_usage__"
	SystemLoadResourceName = "__system_load__"
)

var (
// Global ROOT statistic node that represents the universal parent node.
//Root, _ = api.Entry(RootId, api.WithTrafficType(Inbound), api.WithResourceType(ResTypeCommon))

//  Global statistic node for inbound traffic. Usually used for {@code SystemRule} checking.
//EntryNode=
)
