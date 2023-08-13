package mock

import (
	"fmt"
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/util"
	"github.com/buger/jsonparser"
	"github.com/fatih/structs"
	jsoniter "github.com/json-iterator/go"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	ArrayAnyMatch = "[*]"
	//TimeNanoFunc 时间那秒
	TimeNanoFunc   = "${time.Now().UnixNano()}"
	TimeMillisFunc = "${time.Now().UnixNano()/1e6}"
	// TimeSecFunc 获取秒
	TimeSecFunc = "${time.Now().Unix()}"

	// JsonSTag json tag标识
	JsonSTag        = "json"
	Dot      string = ","
)

var (
	jsonTraffic = jsoniter.ConfigCompatibleWithStandardLibrary
	emptyReg    = regexp.MustCompile(`,\s*{\s*}`)
	emptyReg2   = regexp.MustCompile(`,\s*]`)
)

type TrafficShapingController interface {
	PerformCheckingFunc(ctx *base.EntryContext) *base.TokenResult
	PerformCheckingArgs(ctx *base.EntryContext) *base.TokenResult

	BoundRule() *Rule
	ArgsCheck(ctx *base.EntryContext) *RuleItem
	// MockCheck mock生效检查，要求必须包含请求头
	MockCheck(ctx *base.EntryContext) bool
}

type baseTrafficShapingController struct {
	r               *Rule
	strategy        Strategy
	res             string
	specificItems   []RuleItem
	additionalItems []AdditionalItem
}

func newBaseTrafficShapingController(r *Rule) *baseTrafficShapingController {
	if r.SpecificItems == nil {
		r.SpecificItems = []RuleItem{}
	}
	if r.AdditionalItems == nil {
		r.AdditionalItems = []AdditionalItem{}
	}
	return &baseTrafficShapingController{
		r:               r,
		res:             r.Resource,
		specificItems:   r.SpecificItems,
		strategy:        r.Strategy,
		additionalItems: r.AdditionalItems,
	}
}

func (c *baseTrafficShapingController) BoundRule() *Rule {
	return c.r
}

func (c *baseTrafficShapingController) extractArgs(ctx *base.EntryContext) []interface{} {
	args := ctx.Input.Args
	return args
}

func (c *baseTrafficShapingController) extractAttachmentArgs(ctx *base.EntryContext) map[interface{}]interface{} {
	attachments := ctx.Input.Attachments
	if attachments == nil {
		if logging.DebugEnabled() {
			logging.Debug("[paramKey] The attachments of ctx is nil", "args", attachments)
		}
		return nil
	}
	return attachments
}
func (c *baseTrafficShapingController) HeadersCheck(ctx *base.EntryContext) bool {
	if len(ctx.Input.Headers) == 0 {
		return true
	}
	if ctx.Resource.Classification() != base.ResTypeWeb {
		return true
	}
	additionalItems := c.additionalItems
	if len(additionalItems) == 0 {
		return false
	}
	var op = c.r.Op
	if op == And && len(ctx.Input.Headers) < len(additionalItems) {
		return false
	}
	var containKey = false
	for k, v := range ctx.Input.Headers {
		for _, item := range additionalItems {
			if k == item.Key {
				containKey = true
				if op == Or {
					if strings.Join(v, ",") == item.Value {
						return true
					}
				} else {
					if strings.Join(v, ",") != item.Value {
						return false
					}
				}
			}
		}
	}
	if !containKey {
		return false
	}
	return true
}

func (c *baseTrafficShapingController) ContextCheck(ctx *base.EntryContext) bool {
	if len(ctx.Input.MetaData) == 0 {
		return true
	}
	if ctx.Resource.Classification() != base.ResTypeMicro {
		return true
	}
	additionalItems := c.additionalItems
	if len(additionalItems) == 0 {
		return true
	}
	var op = c.r.Op
	if op == And && len(ctx.Input.MetaData) < len(additionalItems) {
		return false
	}
	var containKey = false
	for k, v := range ctx.Input.MetaData {
		for _, item := range additionalItems {
			if k == item.Key {
				containKey = true
				if op == Or {
					if v == item.Value {
						return true
					}
				} else {
					if v != item.Value {
						return false
					}
				}
			}
		}
	}
	if !containKey {
		return false
	}
	return true
}

func (c *baseTrafficShapingController) mockReplaceCheck(ctx *base.EntryContext) bool {
	if len(ctx.Input.MetaData) == 0 {
		return false
	}
	additionalItems := c.additionalItems
	if len(additionalItems) == 0 {
		return false
	}
	var op = c.r.Op
	if op == And && len(ctx.Input.MetaData) < len(additionalItems) {
		return false
	}
	var containKey = false
	for k, v := range ctx.Input.MetaData {
		for _, item := range additionalItems {
			if k == item.Key {
				containKey = true
				if op == Or {
					if v == item.Value {
						return true
					}
				} else {
					if v != item.Value {
						return false
					}
				}
			}
		}
	}
	return containKey
}

// ruleItemCheck 规则Item check
func (c *baseTrafficShapingController) ruleItemCheck(ctx *base.EntryContext, item RuleItem) bool {
	if len(ctx.Input.MetaData) == 0 {
		return false
	}
	// 不包含附加参数，返回真
	if strings.TrimSpace(item.AdditionalItemKey) == "" {
		return true
	}
	// 如果包含了附加参数，则匹配返回True，否则返回False
	for k, v := range ctx.Input.MetaData {
		if k == item.AdditionalItemKey && v == item.AdditionalItemValue {
			return true
		}
	}
	return false
}

// MockCheck mock生效检查，要求必须包含请求头
func (c *baseTrafficShapingController) MockCheck(ctx *base.EntryContext) bool {
	if ctx.Resource.Classification() == base.ResTypeMicro {
		return c.mockReplaceCheck(ctx)
	}
	return c.HeadersCheck(ctx)
}

// ArgsCheck 参数检查
func (c *baseTrafficShapingController) ArgsCheck(ctx *base.EntryContext) *RuleItem {
	if c == nil {
		return nil
	}
	if len(c.specificItems) == 0 {
		return nil
	}
	attachmentArgs := c.extractAttachmentArgs(ctx)
	if attachmentArgs != nil && len(attachmentArgs) > 0 {
		for _, item := range c.specificItems {
			if item.WhenParamKey == "" {
				if logging.DebugEnabled() {
					logging.Debug("[paramKey] The param key is nil",
						"args", attachmentArgs, "paramKey", item.WhenParamKey)
				}
				continue
			}
			arg, ok := attachmentArgs[item.WhenParamKey]
			if !ok {
				if logging.DebugEnabled() {
					logging.Debug("[paramKey] extracted data does not exist",
						"args", attachmentArgs, "paramKey", item.WhenParamKey)
				}
				continue
			}
			if item.WhenParamValue == arg {
				return &item
			}
		}
	}
	args := c.extractArgs(ctx)
	if args != nil {
		// 资源类型是go-micro并且是结构体时，强制转型为结构体数据
		if ctx.Resource.Classification() == base.ResTypeMicro && structs.IsStruct(args[0]) {
			if requestJsonData, err := jsonTraffic.Marshal(args[0]); err == nil {
				// 模式匹配
				for _, item := range c.specificItems {
					if !c.ruleItemCheck(ctx, item) {
						continue
					}
					var propertyArr = strings.Split(item.WhenParamKey, ".")
					// 先替换，无论是否匹配，都可以先替换
					// nano方法替换
					item.ThenReturnMockData = strings.ReplaceAll(item.ThenReturnMockData, TimeNanoFunc, strconv.FormatInt(time.Now().UnixNano(), 10))
					// 毫秒方法替换
					item.ThenReturnMockData = strings.ReplaceAll(item.ThenReturnMockData, TimeMillisFunc, strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
					// 秒方法替换
					item.ThenReturnMockData = strings.ReplaceAll(item.ThenReturnMockData, TimeSecFunc, strconv.FormatInt(time.Now().Unix(), 10))
					if item.WhenParamKind == KindString {
						var result = c.kindStringHandle(requestJsonData, propertyArr, item, ctx)
						if result != nil {
							return result
						}
					} else {
						var result = c.generalHandle(requestJsonData, propertyArr, item, ctx)
						if result != nil {
							return result
						}
					}
				}
			} else {
				logging.Warn("request cannot transfer to json", "err", err)
			}
		} else if ctx.Resource.Classification() == base.ResTypeWeb {
			// 对于Web，args存储的格式为key=value，所以需要先切割，再对比
			for _, item := range c.specificItems {
				for _, arg := range args {
					kv := strings.SplitN(arg.(string), "=", 2)
					if len(kv) != 2 {
						continue
					}
					if item.WhenParamKey == kv[0] {
						if item.WhenParamValue == kv[1] {
							return &item
						}
						// 不匹配，直接跳出
						break
					}
				}
			}
		} else {
			for _, item := range c.specificItems {
				var idx int
				if item.WhenParamIdx < 0 {
					idx = len(args) + idx
				}
				if idx < 0 {
					if logging.DebugEnabled() {
						logging.Debug("[extractArgs] The param index of mock traffic shaping controller is invalid",
							"args", args, "paramIndex", item.WhenParamIdx)
					}
					continue
				}
				if idx >= len(args) {
					if logging.DebugEnabled() {
						logging.Debug("[extractArgs] The argument in index doesn't exist",
							"args", args, "paramIndex", item.WhenParamIdx)
					}
					continue
				}
				if item.WhenParamValue == args[item.WhenParamIdx] {
					return &item
				}
			}
		}
	}
	return nil
}

// getPropertyValue 获取key对应的value
func (c *baseTrafficShapingController) getPropertyValue(paramKey string, requestJsonData []byte) (string, error) {
	var propertyArr2 = strings.Split(paramKey, ".")
	return jsonparser.GetString(requestJsonData, propertyArr2...)
}

// kindStringHandle string类型数据处理
func (c *baseTrafficShapingController) kindStringHandle(requestJsonData []byte, propertyArr []string, item RuleItem, ctx *base.EntryContext) *RuleItem {
	val, err := jsonparser.GetString(requestJsonData, propertyArr...)
	if err != nil {
		logging.Warn("get property failed", "property", item.WhenParamKey, "request data", requestJsonData, "err", err)
		return nil
	}
	var changeData []byte
	// 响应替换
	if item.MockReplace == Resp {
		// mock替换匹配,把请求属性值替换响应中的属性值
		changeData, err = jsonparser.Set([]byte(item.ThenReturnMockData), []byte(fmt.Sprint(`"`, val, `"`)), strings.Split(item.ReplaceAttribute, ".")...)
		if err != nil {
			logging.Warn("set property failed", "property", item.WhenParamKey, "thenReturnMockData", item.ThenReturnMockData, "request data", requestJsonData, "err", err)
			return nil
		}
		item.ThenReturnMockData = string(changeData)
		return &item
	}
	// 请求替换
	if item.MockReplace == Req {
		// mock替换，把请求体中的属性值替换为mock值，如请求提为userId=1,目标是替换为1000，则替换后userId=1000
		WalkAndSet(structs.Fields(c.extractArgs(ctx)[0]), 0, item.ThenReturnMockData, propertyArr...)
		item.TmpData = c.extractArgs(ctx)[0]
		if err != nil {
			logging.Warn("set property failed", "property", item.WhenParamKey, "thenReturnMockData", item.ThenReturnMockData, "request data", requestJsonData, "err", err)
			return nil
		}
		item.ThenReturnMockData = string(changeData)
		return &item
	}
	// 数据Mock逻辑
	// 首先区分是And还是Or，对于And需要三个都匹配才可以返回，对于Or，只需要一个匹配就可以了
	var isMock bool
	if item.ParamOP == And {
		// ----------判断条件1----------
		isMock = c.isMatch(item.WhenParamValue, val, item.MatchPattern)
		// ----------判断条件2----------
		if isMock && strings.TrimSpace(item.WhenParamKey2) != "" {
			var val2, err = c.getPropertyValue(item.WhenParamKey2, requestJsonData)
			if err != nil {
				logging.Warn("get property failed", "property2", item.WhenParamKey2, "request data", requestJsonData, "err", err)
				return nil
			}
			isMock = isMock && c.isMatch(item.WhenParamValue2, val2, item.MatchPattern)
		}
		// ----------判断条件3----------
		if isMock && strings.TrimSpace(item.WhenParamKey3) != "" {
			var val3, err = c.getPropertyValue(item.WhenParamKey3, requestJsonData)
			if err != nil {
				logging.Warn("get property failed", "property3", item.WhenParamKey3, "request data", requestJsonData, "err", err)
				return nil
			}
			isMock = isMock && c.isMatch(item.WhenParamValue3, val3, item.MatchPattern)
		}
	} else {
		// ----------判断条件1----------
		isMock = c.isMatch(item.WhenParamValue, val, item.MatchPattern)
		// ----------判断条件2----------
		if !isMock && strings.TrimSpace(item.WhenParamKey2) != "" {
			var val2, err = c.getPropertyValue(item.WhenParamKey2, requestJsonData)
			if err != nil {
				logging.Warn("get property failed", "property2", item.WhenParamKey2, "request data", requestJsonData, "err", err)
				return nil
			}
			isMock = c.isMatch(item.WhenParamValue2, val2, item.MatchPattern)
		}
		// ----------判断条件3----------
		if !isMock && strings.TrimSpace(item.WhenParamKey3) != "" {
			var val3, err = c.getPropertyValue(item.WhenParamKey3, requestJsonData)
			if err != nil {
				logging.Warn("get property failed", "property3", item.WhenParamKey3, "request data", requestJsonData, "err", err)
				return nil
			}
			isMock = c.isMatch(item.WhenParamValue3, val3, item.MatchPattern)
		}
	}
	// 返回结果
	if isMock {
		return &item
	}
	return nil
}

// generalHandle 通用类型处理
func (c *baseTrafficShapingController) generalHandle(requestJsonData []byte, propertyArr []string, item RuleItem, ctx *base.EntryContext) *RuleItem {
	// string数组处理
	val, dt, _, err := jsonparser.Get(requestJsonData, propertyArr...)
	if err != nil {
		logging.Warn("get property failed", "property", item.WhenParamKey, "request data", requestJsonData, "err", err)
		return nil
	}
	// 响应替换
	if item.MockReplace == Resp {
		var valS string
		if dt == jsonparser.Array || dt == jsonparser.Boolean {
			valS = fmt.Sprint(``, string(val), ``)
		} else {
			valS = fmt.Sprint(`"`, string(val), `"`)
		}
		var changeData []byte
		var replaceAttributeArr = strings.Split(item.ReplaceAttribute, ".")
		if dt == jsonparser.Array && strings.Contains(item.ReplaceAttribute, ArrayAnyMatch) {
			var keysPreBreak = false
			// [*]前数组
			var keysPre []string
			// [*]后数组
			var keysPost []string
			for _, r := range replaceAttributeArr {
				if r == ArrayAnyMatch {
					keysPreBreak = true
					continue
				}
				if !keysPreBreak {
					keysPre = append(keysPre, r)
				} else {
					keysPost = append(keysPost, r)
				}
			}
			var index = 0
			_, err = jsonparser.ArrayEach(requestJsonData, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				var arrIndex = fmt.Sprintf("[%d]", index)
				// 下标加1
				index++
				var finalKeys = keysPre
				finalKeys = append(finalKeys, arrIndex)
				finalKeys = append(finalKeys, keysPost...)
				// 为了预防下标越界
				_, _, _, err = jsonparser.Get([]byte(item.ThenReturnMockData), finalKeys...)
				if err == nil {
					changeData, err = jsonparser.Set([]byte(item.ThenReturnMockData), value, finalKeys...)
					if err != nil {
						logging.Warn("set property failed in ArrayEach", "property", item.WhenParamKey, "thenReturnMockData", item.ThenReturnMockData, "request data", requestJsonData, "err", err)
					} else {
						item.ThenReturnMockData = string(changeData)
					}
				} else {
					logging.Warn("get property failed in ArrayEach", "property", item.WhenParamKey, "thenReturnMockData", item.ThenReturnMockData, "request data", requestJsonData, "err", err)
				}
			}, propertyArr...)
			// 移除mock中多余的数据
			var deleteIndex = 0
			var originMockData = item.ThenReturnMockData
			// 移除mock中多余的数据
			_, err = jsonparser.ArrayEach([]byte(originMockData), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				if deleteIndex < index {
					deleteIndex++
					return
				}
				var arrIndex = fmt.Sprintf("[%d]", deleteIndex)
				// 下标不需要变更，删除后数据下标会前移
				// deleteIndex++
				var finalKeys = keysPre
				finalKeys = append(finalKeys, arrIndex)
				// 为了预防下标越界
				_, _, _, err = jsonparser.Get([]byte(item.ThenReturnMockData), finalKeys...)
				if err == nil {
					changeData = jsonparser.Delete([]byte(item.ThenReturnMockData), finalKeys...)
					// 如果数组中最后一个为",{}"，删除",{}"
					var replaceStr = emptyReg.ReplaceAllString(string(changeData), "")
					// 如果数组最后一个是"，"，删除最后一个","
					replaceStr = emptyReg2.ReplaceAllString(replaceStr, "]")
					item.ThenReturnMockData = replaceStr
					changeData = []byte(replaceStr)
				} else {
					logging.Warn("get property failed in ArrayEach", "property", item.WhenParamKey, "thenReturnMockData", item.ThenReturnMockData, "request data", requestJsonData, "err", err)
				}
			}, keysPre...)
		} else {
			// mock替换匹配
			changeData, err = jsonparser.Set([]byte(item.ThenReturnMockData), []byte(valS), strings.Split(item.ReplaceAttribute, ".")...)
			if err != nil {
				logging.Warn("set property failed", "property", item.WhenParamKey, "thenReturnMockData", item.ThenReturnMockData, "request data", requestJsonData, "err", err)
				return nil
			}
		}
		item.ThenReturnMockData = string(changeData)
		return &item
	}
	// 请求替换
	if item.MockReplace == Req {
		// mock替换匹配，把请求内容的属性替换为mock中的值
		var replaceValue string
		if dt == jsonparser.Array || dt == jsonparser.Boolean {
			replaceValue = fmt.Sprint(``, item.ThenReturnMockData, ``)
		} else {
			replaceValue = fmt.Sprint(`"`, item.ThenReturnMockData, `"`)
		}
		WalkAndSet(structs.Fields(c.extractArgs(ctx)[0]), 0, replaceValue, propertyArr...)
		item.TmpData = c.extractArgs(ctx)[0]
		return &item
	}
	// 数据Mock逻辑
	// 首先区分是And还是Or，对于And需要三个都匹配才可以返回，对于Or，只需要一个匹配就可以了
	var isMock bool
	if item.ParamOP == And {
		// ----------判断条件1----------
		isMock = item.WhenParamValue == string(val)
		if isMock && strings.TrimSpace(item.WhenParamKey2) != "" {
			var val2, err = c.getPropertyValue(item.WhenParamKey2, requestJsonData)
			if err != nil {
				logging.Warn("get property failed", "property2", item.WhenParamKey2, "request data", requestJsonData, "err", err)
				return nil
			}
			isMock = isMock && item.WhenParamValue2 == val2
		}
		// ----------判断条件3----------
		if isMock && strings.TrimSpace(item.WhenParamKey3) != "" {
			var val3, err = c.getPropertyValue(item.WhenParamKey3, requestJsonData)
			if err != nil {
				logging.Warn("get property failed", "property3", item.WhenParamKey3, "request data", requestJsonData, "err", err)
				return nil
			}
			isMock = isMock && item.WhenParamValue3 == val3
		}
	} else {
		// ----------判断条件1----------
		isMock = item.WhenParamValue == string(val)
		// ----------判断条件2----------
		if !isMock && strings.TrimSpace(item.WhenParamKey2) != "" {
			var val2, err = c.getPropertyValue(item.WhenParamKey2, requestJsonData)
			if err != nil {
				logging.Warn("get property failed", "property2", item.WhenParamKey2, "request data", requestJsonData, "err", err)
				return nil
			}
			isMock = item.WhenParamValue2 == val2
		}
		// ----------判断条件3----------
		if !isMock && strings.TrimSpace(item.WhenParamKey3) != "" {
			var val3, err = c.getPropertyValue(item.WhenParamKey3, requestJsonData)
			if err != nil {
				logging.Warn("get property failed", "property3", item.WhenParamKey3, "request data", requestJsonData, "err", err)
				return nil
			}
			isMock = item.WhenParamValue3 == val3
		}
	}
	// 返回结果
	if isMock {
		return &item
	}
	return nil
}
func (c *baseTrafficShapingController) isMatch(whenParamValue, val string, matchPattern MatchPattern) bool {
	// 根据匹配模式匹配
	if matchPattern == ExactMatch || matchPattern > RegularMatch {
		if val == whenParamValue {
			return true
		}
	} else if matchPattern == PrefixMatch {
		// 首先断言val为字符串，如果不为字符串，则直接跳出，否则进行前缀匹配
		if strings.HasPrefix(val, whenParamValue) {
			return true
		}
	} else if matchPattern == SuffixMatch {
		// 首先断言val为字符串，如果不为字符串，则直接跳出，否则进行后缀匹配
		if strings.HasSuffix(val, whenParamValue) {
			return true
		}
	} else if matchPattern == ContainMatch {
		// 首先断言val为字符串，如果不为字符串，则直接跳出，否则进行包含匹配
		if strings.Contains(val, whenParamValue) {
			return true
		}
	} else if matchPattern == RegularMatch {
		// 首先断言val为字符串，如果不为字符串，则直接跳出，否则进行正则匹配,如果匹配错误，则尝试包含匹配
		if ok2, err := regexp.MatchString(whenParamValue, val); err != nil {
			logging.Warn("mock regular match error,then try contains match", "error", err)
			if strings.Contains(val, whenParamValue) {
				return true
			}
		} else if ok2 {
			return true
		}
	}
	return false
}

// WalkAndSet 遍历并且替换，从field中找到匹配的json属性，如果找到则把属性替换为val
func WalkAndSet(fields []*structs.Field, pos int, val interface{}, properties ...string) {
	if len(fields) == 0 {
		return
	}
	if pos > len(properties) {
		return
	}
	var property = properties[pos]
	for _, field := range fields {
		if strings.Split(field.Tag(JsonSTag), Dot)[0] == property {
			if pos == len(properties)-1 {
				err := field.Set(val)
				if err != nil {
					logging.Warn("WalkAndSet Error", "val", val, "error", err)
				}
				return
			}
			WalkAndSet(field.Fields(), pos+1, val, properties...)
		}
	}
}

func (c *baseTrafficShapingController) DoInnerCheck(ctx *base.EntryContext) *base.TokenResult {
	item := c.ArgsCheck(ctx)
	if item == nil {
		return nil
	}
	if item.MockReplace == Req {
		return base.NewTokenResultBlockedWithCause(base.BlockTypeMockRequest, "", c.BoundRule(), item.TmpData)
	}
	if item.MockReplace == CtxTimeout {
		return base.NewTokenResultBlockedWithCause(base.BlockTypeMockCtxTimeout, "", c.BoundRule(), item.ThenReturnWaitingTimeMs)
	}
	// 内层继续判断
	if DoNothing == item.ControlBehavior {
		return nil
	} else if Panic == item.ControlBehavior {
		return base.NewTokenResultBlockedWithCause(base.BlockTypeMockError, "", c.BoundRule(), item.ThenThrowMsg)
	} else if Mock == item.ControlBehavior {
		return base.NewTokenResultBlockedWithCause(base.BlockTypeMock, "", c.BoundRule(), item.ThenReturnMockData)
	} else if Waiting == item.ControlBehavior {
		if nanosToWait := item.ThenReturnWaitingTimeMs * time.Millisecond.Nanoseconds(); nanosToWait > 0 {
			// Handle waiting action.
			util.Sleep(time.Duration(nanosToWait))
		}
		return nil
	} else if WaitingThenPanic == item.ControlBehavior {
		if nanosToWait := item.ThenReturnWaitingTimeMs * time.Millisecond.Nanoseconds(); nanosToWait > 0 {
			// Handle waiting action.
			util.Sleep(time.Duration(nanosToWait))
		}
		return base.NewTokenResultBlockedWithCause(base.BlockTypeMockError, "", c.BoundRule(), item.ThenThrowMsg)
	} else if WaitingThenMock == item.ControlBehavior {
		if nanosToWait := item.ThenReturnWaitingTimeMs * time.Millisecond.Nanoseconds(); nanosToWait > 0 {
			// Handle waiting action.
			util.Sleep(time.Duration(nanosToWait))
		}
		return base.NewTokenResultBlockedWithCause(base.BlockTypeMock, "", c.BoundRule(), item.ThenReturnMockData)
	}
	return nil
}
