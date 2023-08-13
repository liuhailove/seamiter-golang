package microv4

//import (
//	"context"
//	"encoding/json"
//	sea "github.com/liuhailove/seamiter-golang/api"
//	"github.com/liuhailove/seamiter-golang/core/base"
//	"github.com/liuhailove/seamiter-golang/core/config"
//	"github.com/pkg/errors"
//	"go-micro.dev/v4/metadata"
//	"go-micro.dev/v4/server"
//)
//
//// NewHandlerWrapper returns a Handler Wrapper with  sea breaker
//func NewHandlerWrapper(seaOpts ...Option) server.HandlerWrapper {
//	return func(h server.HandlerFunc) server.HandlerFunc {
//		return func(ctx context.Context, req server.Request, rsp interface{}) error {
//			if !config.CloseAll() {
//				resourceName := req.Service() + "." + req.Endpoint()
//				opts := evaluateOptions(seaOpts)
//				if opts.serverResourceExtract != nil {
//					resourceName = opts.serverResourceExtract(ctx, req)
//				}
//				metaDataMap := make(map[string]string, 0)
//				metaData, ok := metadata.FromContext(ctx)
//				if ok {
//					for k, v := range metaData {
//						metaDataMap[k] = v
//					}
//				}
//				entry, blockErr := sea.Entry(
//					resourceName,
//					sea.WithResourceType(base.ResTypeMicro),
//					sea.WithTrafficType(base.Inbound),
//					sea.WithArgs(req.Body()),
//					sea.WithRsps(rsp),
//					sea.WithMetaData(metaDataMap))
//				if blockErr != nil {
//					if blockErr.BlockType() == base.BlockTypeMock {
//						if strVal, ok := blockErr.TriggeredValue().(string); ok {
//							err := json.Unmarshal([]byte(strVal), rsp)
//							if err != nil {
//								sea.TraceError(entry, err)
//							}
//							return err
//						}
//						return blockErr
//					}
//					if blockErr.BlockType() == base.BlockTypeMockError {
//						if strVal, ok := blockErr.TriggeredValue().(string); ok {
//							return errors.New(strVal)
//						}
//						return blockErr
//					}
//					if opts.clientBlockFallback != nil {
//						return opts.serverBlockFallback(ctx, req, blockErr)
//					}
//					return blockErr
//				}
//				defer entry.Exit()
//				err := h(ctx, req, rsp)
//				if err != nil {
//					sea.TraceError(entry, err)
//				}
//				return err
//			}
//			return h(ctx, req, rsp)
//		}
//	}
//}
//
//func NewStreamWrapper(seaOpts ...Option) server.StreamWrapper {
//	return func(stream server.Stream) server.Stream {
//		if !config.CloseAll() {
//			resourceName := stream.Request().Service() + "." + stream.Request().Endpoint()
//			opts := evaluateOptions(seaOpts)
//			if opts.serverResourceExtract != nil {
//				resourceName = opts.streamServerResourceExtract(stream)
//			}
//			entry, blockErr := sea.Entry(resourceName, sea.WithResourceType(base.ResTypeRPC), sea.WithTrafficType(base.Inbound))
//			if blockErr != nil {
//				if opts.serverBlockFallback != nil {
//					return opts.streamServerBlockFallback(stream, blockErr)
//				}
//				stream.Send(blockErr)
//				return stream
//			}
//			defer entry.Exit()
//		}
//		return stream
//	}
//}
