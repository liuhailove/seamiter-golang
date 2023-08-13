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
	}
}

func (r *TokenResult) ResetToPass() {
	r.status = ResultStatusPass
	r.blockErr = nil
	r.nanosToWait = 0
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

func (r *TokenResult) ResetToBlockedWithCause(blockType BlockType, blockMsg string, rule SentinelRule, snapshot interface{}) {
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
	}
	return fmt.Sprintf("TokenResult{status=%s, blockErr=%s, nanosToWait=%d}", r.status.String(), blockMsg, r.nanosToWait)
}

func NewTokenResultPass() *TokenResult {
	return NewTokenResult(ResultStatusPass)
}
func NewTokenResultBlocked(blockType BlockType) *TokenResult {
	return NewTokenResult(ResultStatusBlocked, WithBlockType(blockType))
}
func NewTokenResultBlockedWithCause(blockType BlockType, blockMsg string, rule SentinelRule, snapshot interface{}) *TokenResult {
	return NewTokenResult(ResultStatusBlocked, WithBlockType(blockType), WithBlockMsg(blockMsg), WithRule(rule), WithSnapshotValue(snapshot))
}

func NewTokenResult(status TokenResultStatus, blockErrOpts ...BlockErrorOption) *TokenResult {
	return &TokenResult{
		status:      status,
		blockErr:    NewBlockError(blockErrOpts...),
		nanosToWait: 0,
	}
}
func NewTokenResultShouldWait(waitNs time.Duration) *TokenResult {
	result := NewTokenResult(ResultStatusShouldWait)
	result.nanosToWait = waitNs
	return result
}
