package file

import (
	"github.com/liuhailove/seamiter-golang/core/config"
	"github.com/liuhailove/seamiter-golang/ext/datasource"
	"github.com/liuhailove/seamiter-golang/ext/datasource/util"
	"github.com/liuhailove/seamiter-golang/logging"
	util2 "github.com/liuhailove/seamiter-golang/util"
)

var (
	isInitialized util2.AtomicBool
)

func Initialize() {
	if !isInitialized.CompareAndSet(false, true) {
		return
	}
	if config.RuleConsistentModeType() == config.FileMode {
		// 流控规则
		flowHandler := datasource.NewFlowRulesHandler(datasource.FlowRuleJsonArrayParser)
		dsFlowRule := NewFileDataSource(config.SourceFilePath(), config.FlowRuleName(), flowHandler)
		err := dsFlowRule.Initialize()
		if err != nil {
			logging.Error(err, "DsFlowRule Fail to Initialize datasource error", err)
			return
		}
		util.RegisterFlowDataSource(dsFlowRule)

		// 授权规则
		dsAuthorityRule := NewFileDataSource(config.SourceFilePath(), config.AuthorityRuleName())
		err = dsAuthorityRule.Initialize()
		if err != nil {
			logging.Error(err, "DsAuthorityRule Fail to Initialize datasource error", err)
			return
		}
		util.RegisterAuthorityDataSource(dsAuthorityRule)

		// 降级规则
		circuitHandler := datasource.NewCircuitBreakerRulesHandler(datasource.CircuitBreakerRuleJsonArrayParser)
		dsDegradeRule := NewFileDataSource(config.SourceFilePath(), config.DegradeRuleName(), circuitHandler)
		err = dsDegradeRule.Initialize()
		if err != nil {
			logging.Error(err, "DsDegradeRule Fail to Initialize datasource error", err)
			return
		}
		util.RegisterDegradeDataSource(dsDegradeRule)

		// 降级规则
		systemHandler := datasource.NewSystemRulesHandler(datasource.SystemRuleJsonArrayParser)
		dsSystemRule := NewFileDataSource(config.SourceFilePath(), config.SystemRuleName(), systemHandler)
		err = dsSystemRule.Initialize()
		if err != nil {
			logging.Error(err, "DsSystemRule Fail to Initialize datasource error", err)
			return
		}
		util.RegisterSystemDataSource(dsSystemRule)

		// 热点规则
		hotspotHandler := datasource.NewHotSpotParamRulesHandler(datasource.HotSpotParamRuleJsonArrayParser)
		dsHotspotRule := NewFileDataSource(config.SourceFilePath(), config.HotspotRuleName(), hotspotHandler)
		err = dsHotspotRule.Initialize()
		if err != nil {
			logging.Error(err, "DsHotspotRule Fail to Initialize datasource error", err)
			return
		}
		util.RegisterHotspotSource(dsHotspotRule)

		// mock规则
		mockHandler := datasource.NewMockRulesHandler(datasource.MockRuleJsonArrayParser)
		dsMockRule := NewFileDataSource(config.SourceFilePath(), config.MockRuleName(), mockHandler)
		err = dsMockRule.Initialize()
		if err != nil {
			logging.Error(err, "DsMockRule Fail to Initialize datasource error", err)
			return
		}
		util.RegisterMockDataSource(dsMockRule)

		// retry规则
		retryHandler := datasource.NewRetryRulesHandler(datasource.RetryRuleJsonArrayParser)
		dsRetryRule := NewFileDataSource(config.SourceFilePath(), config.RetryRuleName(), retryHandler)
		err = dsRetryRule.Initialize()
		if err != nil {
			logging.Error(err, "DsRetryRule Fail to Initialize datasource error", err)
			return
		}
		util.RegisterRetryDataSource(dsRetryRule)

		// gray规则
		grayHandler := datasource.NewGrayRulesHandler(datasource.GrayRuleJsonArrayParser)
		dsGrayRule := NewFileDataSource(config.SourceFilePath(), config.GrayRuleName(), grayHandler)
		err = dsGrayRule.Initialize()
		if err != nil {
			logging.Error(err, "dsGrayRule Fail to Initialize datasource error", err)
			return
		}
		util.RegisterGrayDataSource(dsGrayRule)

		// isolation规则
		isolationHandler := datasource.NewGrayRulesHandler(datasource.IsolationRuleJsonArrayParser)
		dsIsolationRule := NewFileDataSource(config.SourceFilePath(), config.IsolationRuleName(), isolationHandler)
		err = dsIsolationRule.Initialize()
		if err != nil {
			logging.Error(err, "dsIsolationRule Fail to Initialize datasource error", err)
			return
		}
		util.RegisterIsolationDataSource(dsIsolationRule)
	}
}
