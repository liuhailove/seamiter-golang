package util

import "github.com/liuhailove/seamiter-golang/ext/datasource"

var (
	flowDataSource      datasource.DataSource
	authorityDataSource datasource.DataSource
	degradeDataSource   datasource.DataSource
	systemSource        datasource.DataSource
	hotspotSource       datasource.DataSource
	mockDataSource      datasource.DataSource
	grayDataSource      datasource.DataSource
	isolationDataSource datasource.DataSource

	dsMap = make(map[string]datasource.DataSource)
)

func RegisterDataSource(sourceName string, source datasource.DataSource) {
	dsMap[sourceName] = source
}
func RegisterFlowDataSource(source datasource.DataSource) {
	RegisterDataSource("flowDataSource", source)
}

func RegisterAuthorityDataSource(source datasource.DataSource) {
	RegisterDataSource("authorityDataSource", source)
}

func RegisterDegradeDataSource(source datasource.DataSource) {
	RegisterDataSource("degradeDataSource", source)
}

func RegisterSystemDataSource(source datasource.DataSource) {
	RegisterDataSource("systemSource", source)
}

func RegisterHotspotSource(source datasource.DataSource) {
	RegisterDataSource("hotspotSource", source)
}

func RegisterMockDataSource(source datasource.DataSource) {
	RegisterDataSource("mockDataSource", source)
}

func RegisterRetryDataSource(source datasource.DataSource) {
	RegisterDataSource("retryDataSource", source)
}

func RegisterGrayDataSource(source datasource.DataSource) {
	RegisterDataSource("grayDataSource", source)
}

func RegisterIsolationDataSource(source datasource.DataSource) {
	RegisterDataSource("isolationDataSource", source)
}

func GetFlowDataSource() datasource.DataSource {
	return dsMap["flowDataSource"]
}

func GetAuthorityDataSource() datasource.DataSource {
	return dsMap["authorityDataSource"]
}
func GetDegradeDataSource() datasource.DataSource {
	return dsMap["degradeDataSource"]
}
func GetSystemSource() datasource.DataSource {
	return dsMap["systemSource"]
}

func GetHotspotSource() datasource.DataSource {
	return dsMap["hotspotSource"]

}

func GetMockSource() datasource.DataSource {
	return dsMap["mockDataSource"]
}

func GetRetrySource() datasource.DataSource {
	return dsMap["retryDataSource"]
}

func GetGraySource() datasource.DataSource {
	return dsMap["grayDataSource"]
}

func GetIsolationSource() datasource.DataSource {
	return dsMap["isolationDataSource"]
}
