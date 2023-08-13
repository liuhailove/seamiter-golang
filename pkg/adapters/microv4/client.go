package microv4

import (
	"context"
	"encoding/json"
	"fmt"
	sea "github.com/liuhailove/seamiter-golang/api"
	"github.com/liuhailove/seamiter-golang/core/base"
	"github.com/liuhailove/seamiter-golang/core/config"
	"github.com/liuhailove/seamiter-golang/core/retry"
	"github.com/liuhailove/seamiter-golang/core/retry/rule"
	"github.com/liuhailove/seamiter-golang/logging"
	"github.com/buger/jsonparser"
	"github.com/fatih/structs"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/metadata"
	"strings"
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
			sea.WithMetaData(metaDataMap))
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
				var patchMd = metadata.Metadata{}
				patchMd["grayTag"] = entry.GrayTag()
				ctx = metadata.MergeContext(ctx, patchMd, false)
			}
			if len(entry.GrayAddress()) > 0 {
				optArr = append(optArr, client.WithAddress(entry.GrayAddress()...))
			}
		}
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
		} else {
			err = c.Client.Call(ctx, req, rsp, optArr...)
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
		logging.Info("DoWithRetry", "retry count", content.GetRetryCount(), "err", content.GetLastError())
	}
	err := g.client.Call(g.ctx, g.req, g.rsp, g.optArr...)
	if err != nil {
		return err
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
							return errors.New("additionalItem match,can retry ,value=" + value)
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
