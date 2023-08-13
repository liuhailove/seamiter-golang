package micro

import (
	"context"
	"encoding/json"
	"github.com/liuhailove/seamiter-golang/core/base"
	"github.com/liuhailove/seamiter-golang/core/config"
	"github.com/liuhailove/seamiter-golang/core/retry"
	"github.com/liuhailove/seamiter-golang/core/retry/rule"
	"github.com/liuhailove/seamiter-golang/logging"

	sea "github.com/liuhailove/seamiter-golang/api"
	"github.com/micro/go-micro/v2/client"
)

type clientWrapper struct {
	client.Client
	Opts []Option
}

func (c *clientWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	if !config.CloseAll() {
		resourceName := req.Service() + "." + req.Endpoint()
		options := evaluateOptions(c.Opts)

		if options.clientResourceExtract != nil {
			resourceName = options.clientResourceExtract(ctx, req)
		}

		entry, blockErr := sea.Entry(
			resourceName,
			sea.WithResourceType(base.ResTypeMicro),
			sea.WithTrafficType(base.Outbound),
		)
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
				if strVal, ok := blockErr.TriggeredValue().(string); ok {
					err := json.Unmarshal([]byte(strVal), req.Body())
					if err != nil {
						sea.TraceError(entry, err)
					}
					return err
				}
				err := c.Client.Call(ctx, req, rsp, opts...)
				if err != nil {
					sea.TraceError(entry, err)
				}
				return err
			}
			if options.clientBlockFallback != nil {
				return options.clientBlockFallback(ctx, req, blockErr)
			}
			return blockErr
		}
		defer entry.Exit()

		var err error
		// 获取重试模板
		var resRetryTemplate = rule.GetRetryTemplateOfResource(resourceName)
		if resRetryTemplate != nil {
			// 模板调用
			_, err = resRetryTemplate.Execute(&GrpcRetryCallback{
				c.Client,
				ctx,
				opts,
				req,
				rsp,
			})
		} else {
			err = c.Client.Call(ctx, req, rsp, opts...)
		}
		if err != nil {
			sea.TraceError(entry, err)
		}
		return err
	}
	return c.Client.Call(ctx, req, rsp, opts...)
}

// GrpcRetryCallback grpc回调结构体
type GrpcRetryCallback struct {
	client client.Client
	ctx    context.Context
	optArr []client.CallOption
	req    client.Request
	rsp    interface{}
}

func (g *GrpcRetryCallback) DoWithRetry(content retry.RtyContext) interface{} {
	if logging.ErrorEnabled() {
		logging.Debug("DoWithRetry", "retry count", content.GetRetryCount(), "err", content.GetLastError())
	}
	err := g.client.Call(g.ctx, g.req, g.rsp, g.optArr...)
	if err != nil {
		panic(err)
	}
	return nil
}
func (c *clientWrapper) Stream(ctx context.Context, req client.Request, opts ...client.CallOption) (client.Stream, error) {
	if !config.CloseAll() {
		resourceName := req.Service() + "." + req.Endpoint()
		options := evaluateOptions(c.Opts)

		if options.streamClientResourceExtract != nil {
			resourceName = options.streamClientResourceExtract(ctx, req)
		}

		entry, blockErr := sea.Entry(
			resourceName,
			sea.WithResourceType(base.ResTypeMicro),
			sea.WithTrafficType(base.Outbound),
			sea.WithArgs(req.Body()),
		)

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
