package base

import (
	"fmt"
	"time"
)

type BlockType uint8

const (
	BlockTypeUnknown BlockType = iota
	BlockTypeFlow
	BlockTypeIsolation
	BlockTypeCircuitBreaking
	BlockTypeSystemFlow
	BlockTypeHotSpotParamFlow
	BlockTypeMock
	BlockTypeMockError
	BlockTypeMockRequest    // mock请求，代表请求替换
	BlockTypeMockCtxTimeout // mock请求，修改ctx超时时间
	BlockTypeGray           // 灰度错误
)

var (
	blockTypeMap = map[BlockType]string{
		BlockTypeUnknown:          "BlockTypeUnknown",
		BlockTypeFlow:             "BlockTypeFlowControl",
		BlockTypeIsolation:        "BlockTypeIsolation",
		BlockTypeCircuitBreaking:  "BlockTypeCircuitBreaking",
		BlockTypeSystemFlow:       "BlockTypeSystem",
		BlockTypeHotSpotParamFlow: "BlockTypeHotSpotParamFlow",
		BlockTypeMock:             "BlockTypeMock",
		BlockTypeMockError:        "BlockTypeMockError",
		BlockTypeMockRequest:      "BlockTypeMockRequest",
	}
	blockTypeExisted = fmt.Errorf("block type existed")
)

// RegistryBlockType adds block type and corresponding description in order.
func RegistryBlockType(blockType BlockType, desc string) error {
	_, exist := blockTypeMap[blockType]
	if exist {
		return blockTypeExisted
	}
	blockTypeMap[blockType] = desc
	return nil
}

func (t BlockType) String() string {
	name, ok := blockTypeMap[t]
	if ok {
		return name
	}
	return fmt.Sprintf("%d", t)
}

type TokenResultStatus uint8

const (
	ResultStatusPass TokenResultStatus = iota
	ResultStatusBlocked
	ResultStatusShouldWait
)

func (s TokenResultStatus) String() string {
	switch s {
	case ResultStatusPass:
		return "ResultStatusPass"
	case ResultStatusBlocked:
		return "ResultStatusBlocked"
	case ResultStatusShouldWait:
		return "ResultStatusShouldWait"
	default:
		return "Undefined"
	}
}

type TokenResult struct {
	status TokenResultStatus

	// 灰度相关
	grayRes     *ResourceWrapper //grayRes 关联资源
	grayAddress []string         // 灰度地址
	linkPass    bool             // 链路传递
	grayTag     string           // 灰度标签

	blockErr    *BlockError
	nanosToWait time.Duration
}

func (r *TokenResult) DeepCopyFrom(newResult *TokenResult) {
	r.status = newResult.status
	r.nanosToWait = newResult.nanosToWait
	if r.blockErr == nil {
		r.blockErr = &BlockError{
			blockType:     newResult.blockErr.blockType,
			blockMsg:      newResult.blockErr.blockMsg,
			rule:          newResult.blockErr.rule,
			snapshotValue: newResult.blockErr.snapshotValue,
		}
	} else {
		r.blockErr.blockType = newResult.blockErr.blockType
		r.blockErr.blockMsg = newResult.blockErr.blockMsg
		r.blockErr.rule = newResult.blockErr.rule
		r.blockErr.snapshotValue = newResult.blockErr.snapshotValue
		r.grayRes = newResult.grayRes
		r.grayTag = newResult.grayTag
		r.linkPass = newResult.linkPass
	}
}

// GrayRes 灰度资源
func (r *TokenResult) GrayRes() *ResourceWrapper {
	return r.grayRes
}

// LinkPass 是否链路传递
func (r *TokenResult) LinkPass() bool {
	return r.linkPass
}

// GrayTag 灰度标签
func (r *TokenResult) GrayTag() string {
	return r.grayTag
}

// GrayAddress 灰度地址
func (r *TokenResult) GrayAddress() []string {
	return r.grayAddress
}

func (r *TokenResult) ResetToPass() {
	r.status = ResultStatusPass
	r.blockErr = nil
	r.nanosToWait = 0
	r.linkPass = false
	r.grayTag = ""
}

func (r *TokenResult) ResetToBlockedWith(opts ...BlockErrorOption) {
	r.status = ResultStatusBlocked
	if r.blockErr == nil {
		r.blockErr = NewBlockError(opts...)
	} else {
		r.blockErr.ResetBlockError(opts...)
	}
	r.nanosToWait = 0
}

func (r *TokenResult) ResetToBlocked(blockType BlockType) {
	r.ResetToBlockedWith(WithBlockType(blockType))
}

func (r *TokenResult) ResetToBlockWithMessage(blockType BlockType, blockMsg string) {
	r.ResetToBlockedWith(WithBlockType(blockType), WithBlockMsg(blockMsg))
}

func (r *TokenResult) ResetToBlockedWithCause(blockType BlockType, blockMsg string, rule SeaRule, snapshot interface{}) {
	r.ResetToBlockedWith(WithBlockType(blockType), WithBlockMsg(blockMsg), WithRule(rule), WithSnapshotValue(snapshot))
}

func (r *TokenResult) IsPass() bool {
	return r.status == ResultStatusPass
}

func (r *TokenResult) IsBlocked() bool {
	return r.status == ResultStatusBlocked
}

func (r *TokenResult) Status() TokenResultStatus {
	return r.status
}

func (r *TokenResult) BlockError() *BlockError {
	return r.blockErr
}

func (r *TokenResult) NanosToWait() time.Duration {
	return r.nanosToWait
}

func (r *TokenResult) String() string {
	var blockMsg string
	if r.blockErr == nil {
		blockMsg = "none"
	} else {
		blockMsg = r.blockErr.Error()
		if r.blockErr.snapshotValue != nil {
			if snap, ok := r.blockErr.snapshotValue.(string); ok {
				blockMsg += ",snapshotValue=" + snap
			}
		}
	}
	return fmt.Sprintf("TokenResult{status=%s, blockErr=%s, nanosToWait=%d}", r.status.String(), blockMsg, r.nanosToWait)
}

func NewTokenResultPass() *TokenResult {
	return NewTokenResult(ResultStatusPass)
}
func NewTokenResultPassWithGrayResource(refResource *ResourceWrapper, linkPass bool, grayTag string, address []string) *TokenResult {
	return NewTokenResultWithResource(ResultStatusPass, refResource, linkPass, grayTag, address)
}
func NewTokenResultBlocked(blockType BlockType) *TokenResult {
	return NewTokenResult(ResultStatusBlocked, WithBlockType(blockType))
}
func NewTokenResultBlockedWithCause(blockType BlockType, blockMsg string, rule SeaRule, snapshot interface{}) *TokenResult {
	return NewTokenResult(ResultStatusBlocked, WithBlockType(blockType), WithBlockMsg(blockMsg), WithRule(rule), WithSnapshotValue(snapshot))
}

func NewTokenResult(status TokenResultStatus, blockErrOpts ...BlockErrorOption) *TokenResult {
	return &TokenResult{
		status:      status,
		blockErr:    NewBlockError(blockErrOpts...),
		nanosToWait: 0,
	}
}

func NewTokenResultWithResource(status TokenResultStatus, refResource *ResourceWrapper, linkPass bool, grayTag string, address []string, blockErrOpts ...BlockErrorOption) *TokenResult {
	return &TokenResult{
		status:      status,
		grayRes:     refResource,
		linkPass:    linkPass,
		grayTag:     grayTag,
		grayAddress: address,
		blockErr:    NewBlockError(blockErrOpts...),
		nanosToWait: 0,
	}
}

func NewTokenResultShouldWait(waitNs time.Duration) *TokenResult {
	result := NewTokenResult(ResultStatusShouldWait)
	result.nanosToWait = waitNs
	return result
}
