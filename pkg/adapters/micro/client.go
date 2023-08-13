package micro

import (
	"context"
	"encoding/json"
	"fmt"
	sea "github.com/liuhailove/seamiter-golang/api"
	"github.com/liuhailove/seamiter-golang/core/base"
	"github.com/liuhailove/seamiter-golang/core/config"
	"github.com/liuhailove/seamiter-golang/core/retry"
	rule "github.com/liuhailove/seamiter-golang/core/retry/rule"
	"github.com/liuhailove/seamiter-golang/logging"
	"github.com/buger/jsonparser"
	"github.com/fatih/structs"
	jsoniter "github.com/json-iterator/go"
	"github.com/micro/go-micro/client"
	micro_error "github.com/micro/go-micro/errors"
	"github.com/micro/go-micro/metadata"
	"github.com/micro/go-micro/util/wrapper"
	"github.com/pkg/errors"
	"math/rand"
	"runtime/debug"
	"strings"
	"time"
)

const (
	DefaultRetryNum = 3
)

type clientWrapper struct {
	client.Client
	Opts []Option
}

func (c *clientWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, optArr ...client.CallOption) error {
	if !config.CloseAll() {
		resourceName := req.Service() + "." + req.Endpoint()
		opts := evaluateOptions(c.Opts)
		if opts.clientResourceExtract != nil {
			resourceName = opts.clientResourceExtract(ctx, req)
		}
		metaDataMap := make(map[string]string, 0)
		metaData, ok := metadata.FromContext(ctx)
		// 来源服务名称
		fromService, _ := metadata.Get(ctx, wrapper.HeaderPrefix+"From-Service")
		if ok {
			for k, v := range metaData {
				metaDataMap[k] = v
			}
		}
		entry, blockErr := sea.Entry(
			resourceName,
			sea.WithResourceType(base.ResTypeMicro),
			sea.WithTrafficType(base.Outbound),
			sea.WithArgs(req.Body()),
			sea.WithRsps(rsp),
			sea.WithMetaData(metaDataMap),
			sea.WithFromService(fromService))
		if blockErr != nil {
			if blockErr.BlockType() == base.BlockTypeMock {
				if strVal, ok := blockErr.TriggeredValue().(string); ok {
					err := json.Unmarshal([]byte(strVal), rsp)
					if err != nil {
						sea.TraceError(entry, err)
					}
					return err
				}
				return blockErr
			}
			if blockErr.BlockType() == base.BlockTypeMockRequest {
				newRequest := c.Client.NewRequest(req.Service(), req.Endpoint(), blockErr.TriggeredValue())
				err := c.Client.Call(ctx, newRequest, rsp, optArr...)
				if err != nil {
					sea.TraceError(entry, err)
				}
				return err
			}
			if blockErr.BlockType() == base.BlockTypeMockError {
				if strVal, ok := blockErr.TriggeredValue().(string); ok {
					return errors.New(strVal)
				}
				return blockErr
			}
			if blockErr.BlockType() == base.BlockTypeMockCtxTimeout {
				if ctxTimeout, ok := blockErr.TriggeredValue().(int64); ok {
					md, success := metadata.FromContext(ctx)
					if success {
						newMd := metadata.Copy(md)
						newCtx, cancel := context.WithTimeout(context.Background(), time.Duration(ctxTimeout*time.Millisecond.Nanoseconds()))
						defer cancel()
						ctx = metadata.NewContext(newCtx, newMd)
					}
					goto RetryLabel
				}
			}
			if opts.clientBlockFallback != nil {
				return opts.clientBlockFallback(ctx, req, blockErr)
			}
			return blockErr
		}
		defer entry.Exit()
		if entry.GrayResource() != nil {
			var service, endpoint, err = splitServiceAndEndpoint(entry.GrayResource().Name())
			if err == nil {
				req = c.Client.NewRequest(service, endpoint, req.Body(), client.WithContentType(req.ContentType()))
			} else {
				logging.Warn("exist error in gray flow", "err", err)
			}
			if entry.LinkPass() {
				md, success := metadata.FromContext(ctx)
				if success {
					newMd := metadata.Copy(md)
					newMd["grayTag"] = entry.GrayTag()
					ctx = metadata.NewContext(ctx, newMd)
				}
			}
			if len(entry.GrayAddress()) > 0 {
				// 在灰度验证时，如果灰度地址部分失效了，需要进行重试，此处设置重试3次，重试间隔为10ms,100ms,1000ms
				// 重排灰度地址，因为当设置地址后，目前默认一直选择第一个地址，这回导致流量不均匀，因此要重排
				optArr = append(optArr, client.WithAddress(randomSort(entry.GrayAddress())...), client.WithRetries(3))
			}
		}
	RetryLabel:
		var err error
		// 获取重试模板
		var rules = rule.GetRulesOfResource(resourceName)
		var resRetryTemplate = rule.GetRetryTemplateOfResource(resourceName)
		if resRetryTemplate != nil && rules != nil {
			// 模板调用
			_, err = resRetryTemplate.Execute(&GrpcRetryCallback{
				c.Client,
				ctx,
				optArr,
				req,
				rsp,
				rules,
			})
			if err != nil && strings.HasPrefix(err.Error(), "additionalItem match,can retry ,value=") {
				err = nil
			}
		} else {
			err = c.Client.Call(ctx, req, rsp, optArr...)
			if err != nil {
				// 灰度报错，重试3次
				var needBreak = true
				for i := 0; i < DefaultRetryNum; i++ {
					// 断言为micro error，为灰度重试
					if microErr, ok := err.(*micro_error.Error); ok {
						if microErr.Code == 500 && microErr.Detail == "error blocked by gray" {
							err = c.Client.Call(ctx, req, rsp, optArr...)
						}
						if err != nil {
							needBreak = false
						} else {
							needBreak = true
						}
					}
					if needBreak {
						break
					}
				}
			}
		}
		if err != nil {
			sea.TraceError(entry, err)
		}
		return err
	}
	return c.Client.Call(ctx, req, rsp, optArr...)
}

// GrpcRetryCallback grpc回调结构体
type GrpcRetryCallback struct {
	client client.Client
	ctx    context.Context
	optArr []client.CallOption
	req    client.Request
	rsp    interface{}
	rules  []rule.Rule
}

func (g *GrpcRetryCallback) DoWithRetry(content retry.RtyContext) interface{} {
	if logging.InfoEnabled() {
		if content.GetRetryCount() == 0 {
			logging.Info("DoWithRetryFirst", "resource", g.req.Service()+"."+g.req.Endpoint(), "retry count", content.GetRetryCount(), "err", content.GetLastError())
		} else {
			logging.Info("DoWithRetryMore", "resource", g.req.Service()+"."+g.req.Endpoint(), "retry count", content.GetRetryCount(), "err", content.GetLastError())
		}
	}
	if content.GetLastError() != nil {
		logging.Info("in micro err handle", "resource", g.req.Service()+"."+g.req.Endpoint(), "retry count", content.GetRetryCount())
		if microErr, ok := content.GetLastError().(*micro_error.Error); ok {
			// context超时异常，针对这种异常，直接重试是无效的，这是因为context被销毁了，因此需要重新生成context
			if microErr.Code == 408 {
				logging.Info("in micro err 408 handle", "resource", g.req.Service()+"."+g.req.Endpoint(), "retry count", content.GetRetryCount())
				// 判断err类型，如果err为GRPC超时异常，则修改context
				md, success := metadata.FromContext(g.ctx)
				if success {
					defer func() {
						if r := recover(); r != nil {
							debug.PrintStack()
						}
					}()
					newMd := metadata.Copy(md)
					newCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					logging.Info("in reset micro err ctx", "resource", g.req.Service()+"."+g.req.Endpoint(), "retry count", content.GetRetryCount())
					g.ctx = metadata.NewContext(newCtx, newMd)
				}
			}
		}
	}
	err := g.client.Call(g.ctx, g.req, g.rsp, g.optArr...)
	if err != nil {
		// 断言为micro error，为灰度重试
		// 灰度报错，重试3次
		var needBreak = true
		for i := 0; i < DefaultRetryNum; i++ {
			// 断言为micro error，为灰度重试
			if microErr, ok := err.(*micro_error.Error); ok {
				if microErr.Code == 500 && microErr.Detail == "error blocked by gray" {
					err = g.client.Call(g.ctx, g.req, g.rsp, g.optArr...)
				}
				if err != nil {
					needBreak = false
				} else {
					needBreak = true
				}
			}
			if needBreak {
				break
			}
		}
		if err != nil {
			panic(err)
		}
	}
	var rules = g.rules
	if len(rules) > 0 {
		var matchRule = rules[0]
		if len(matchRule.SpecificItems) > 0 && structs.IsStruct(g.rsp) {
			if rspJsonData, err := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(g.rsp); err == nil {
				for _, item := range matchRule.SpecificItems {
					if len(item.AdditionalItemKey) == 0 || len(item.AdditionalItemValues) == 0 {
						return nil
					}
					var propertyArr = strings.Split(item.AdditionalItemKey, ".")
					val, dt, _, err := jsonparser.Get(rspJsonData, propertyArr...)
					if err != nil {
						if logging.InfoEnabled() {
							logging.Info("DoWithRetry", "jsonparser", err)
						}
						return nil
					}
					var valString string
					var valFloatString string
					if dt == jsonparser.Boolean {
						var valBool, _ = jsonparser.GetBoolean(rspJsonData, propertyArr...)
						valString = fmt.Sprintf("%t", valBool)
					} else if dt == jsonparser.String {
						valString, _ = jsonparser.GetString(rspJsonData, propertyArr...)
					} else if dt == jsonparser.Number {
						var valInt, _ = jsonparser.GetInt(rspJsonData, propertyArr...)
						valString = fmt.Sprintf("%d", valInt)
						var valFloat, _ = jsonparser.GetFloat(rspJsonData, propertyArr...)
						valFloatString = fmt.Sprintf("%.6f", valFloat)
					} else if dt == jsonparser.Array {
						valString = fmt.Sprint(``, string(val), ``)
					} else {
						valString = fmt.Sprint(`"`, string(val), `"`)
					}
					for _, value := range item.AdditionalItemValues {
						if value == valString || (len(valFloatString) > 0 && value == valFloatString) {
							panic(errors.New("additionalItem match,can retry ,value=" + value))
						}
					}
				}
			}
		}
	}
	return nil
}

func (c *clientWrapper) Stream(ctx context.Context, req client.Request, opts ...client.CallOption) (client.Stream, error) {
	if !config.CloseAll() {
		resourceName := req.Service() + "." + req.Endpoint()
		options := evaluateOptions(c.Opts)
		if options.serverResourceExtract != nil {
			resourceName = options.streamClientResourceExtract(ctx, req)
		}
		entry, blockErr := sea.Entry(
			resourceName,
			sea.WithResourceType(base.ResTypeRPC),
			sea.WithTrafficType(base.Outbound),
			sea.WithArgs(req.Body()))
		if blockErr != nil {
			if options.streamClientBlockFallback != nil {
				return options.streamClientBlockFallback(ctx, req, blockErr)
			}
			return nil, blockErr
		}
		defer entry.Exit()

		stream, err := c.Client.Stream(ctx, req, opts...)
		if err != nil {
			sea.TraceError(entry, err)
		}
		return stream, err
	}
	return c.Client.Stream(ctx, req, opts...)
}

// NewClientWrapper returns a sea client Wrapper.
func NewClientWrapper(opts ...Option) client.Wrapper {
	return func(c client.Client) client.Client {
		return &clientWrapper{c, opts}
	}
}

// splitServiceAndEndpoint 将资源名称且氛围服务和endpoint
func splitServiceAndEndpoint(resource string) (service, endpoint string, err error) {
	if strings.TrimSpace(resource) == "" {
		err = errors.New("resource is empty")
		return
	}
	var lastIndexDot = strings.LastIndex(resource, ".")
	if lastIndexDot < 0 {
		err = errors.New("last index resource dot noe exist")
		return
	}
	var lastSecondIndexDot = strings.LastIndex(resource[:lastIndexDot-2], ".")
	if lastSecondIndexDot < 0 {
		err = errors.New("last second index resource dot noe exist")
		return
	}
	service = resource[:lastSecondIndexDot]
	endpoint = resource[lastSecondIndexDot+1:]
	return
}

func randomSort(sli []string) []string {
	length := len(sli)
	if length <= 1 {
		return sli
	}
	for i := length - 1; i > 0; i-- {
		randNum := rand.Intn(i)
		sli[i], sli[randNum] = sli[randNum], sli[i]
	}
	return sli
}
