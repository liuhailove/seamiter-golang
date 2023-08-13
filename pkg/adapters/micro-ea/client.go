package micro

import (
	"context"
	"encoding/json"
	sea "github.com/liuhailove/seamiter-golang/api"
	"github.com/liuhailove/seamiter-golang/core/base"
	"github.com/liuhailove/seamiter-golang/core/config"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/metadata"
	"github.com/pkg/errors"
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
				if strVal, ok := blockErr.TriggeredValue().(string); ok {
					err := json.Unmarshal([]byte(strVal), req.Body())
					if err != nil {
						sea.TraceError(entry, err)
					}
					return err
				}
				err := c.Client.Call(ctx, req, rsp, optArr...)
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
		err := c.Client.Call(ctx, req, rsp, optArr...)
		if err != nil {
			sea.TraceError(entry, err)
		}
		return err
	}
	return c.Client.Call(ctx, req, rsp, optArr...)
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
