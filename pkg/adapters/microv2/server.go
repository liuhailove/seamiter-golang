package micro

import (
	"context"
	"github.com/liuhailove/seamiter-golang/core/base"
	"github.com/liuhailove/seamiter-golang/core/config"

	sea "github.com/liuhailove/seamiter-golang/api"
	"github.com/micro/go-micro/v2/server"
)

// NewHandlerWrapper returns a Handler Wrapper with Alibaba sea breaker
func NewHandlerWrapper(seaOpts ...Option) server.HandlerWrapper {
	return func(h server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			if !config.CloseAll() {
				resourceName := req.Service() + "." + req.Endpoint()
				opts := evaluateOptions(seaOpts)
				if opts.serverResourceExtract != nil {
					resourceName = opts.serverResourceExtract(ctx, req)
				}
				entry, blockErr := sea.Entry(
					resourceName,
					sea.WithResourceType(base.ResTypeMicro),
					sea.WithTrafficType(base.Inbound),
				)
				if blockErr != nil {
					if opts.serverBlockFallback != nil {
						return opts.serverBlockFallback(ctx, req, blockErr)
					}
					return blockErr
				}
				defer entry.Exit()
				err := h(ctx, req, rsp)
				if err != nil {
					sea.TraceError(entry, err)
				}
				return err
			}
			return h(ctx, req, rsp)
		}
	}
}

func NewStreamWrapper(seaOpts ...Option) server.StreamWrapper {
	return func(stream server.Stream) server.Stream {
		if !config.CloseAll() {
			resourceName := stream.Request().Service() + "." + stream.Request().Endpoint()
			opts := evaluateOptions(seaOpts)
			if opts.serverResourceExtract != nil {
				resourceName = opts.streamServerResourceExtract(stream)
			}
			entry, blockErr := sea.Entry(
				resourceName,
				sea.WithResourceType(base.ResTypeMicro),
				sea.WithTrafficType(base.Inbound),
			)
			if blockErr != nil {
				if opts.serverBlockFallback != nil {
					return opts.streamServerBlockFallback(stream, blockErr)
				}

				stream.Send(blockErr)
				return stream
			}

			entry.Exit()
		}
		return stream
	}
}
