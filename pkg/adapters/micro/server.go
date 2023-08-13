package micro

import (
	"context"
	"encoding/json"
	"fmt"
	sea "github.com/liuhailove/seamiter-golang/api"
	"github.com/liuhailove/seamiter-golang/core/base"
	"github.com/liuhailove/seamiter-golang/core/config"
	"github.com/liuhailove/seamiter-golang/ext/micro"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/client/grpc"
	"github.com/micro/go-micro/metadata"
	"github.com/micro/go-micro/server"
	"github.com/micro/go-micro/util/wrapper"
	"github.com/pkg/errors"
)

const (
	DefaultGrpcPort = 0
)

var (
	ErrBlockedByGray = errors.New("error blocked by gray")
)

// NewHandlerWrapper returns a Handler Wrapper with  sea breaker
func NewHandlerWrapper(seaOpts ...Option) server.HandlerWrapper {
	return func(h server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			if !config.CloseAll() {
				resourceName := req.Service() + "." + req.Endpoint()
				opts := evaluateOptions(seaOpts)
				if opts.serverResourceExtract != nil {
					resourceName = opts.serverResourceExtract(ctx, req)
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
					sea.WithTrafficType(base.Inbound),
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
					if blockErr.BlockType() == base.BlockTypeMockError {
						if strVal, ok := blockErr.TriggeredValue().(string); ok {
							return errors.New(strVal)
						}
						return blockErr
					}
					if opts.clientBlockFallback != nil {
						return opts.serverBlockFallback(ctx, req, blockErr)
					}
					return blockErr
				}
				defer entry.Exit()
				// 命中了灰度规则
				if entry.GrayResource() != nil {
					if entry.LinkPass() {
						md, success := metadata.FromContext(ctx)
						if success {
							newMd := metadata.Copy(md)
							newMd["grayTag"] = entry.GrayTag()
							ctx = metadata.NewContext(ctx, newMd)
						}
					}
					if len(entry.GrayAddress()) > 0 {
						// 判断IP地址和当前地址一致，如果不一致，返回被灰度阻塞错误
						var localAddress string
						if micro.GetGrpcPort() > 0 {
							localAddress = fmt.Sprintf("%s:%d", config.HeartbeatClintIp(), micro.GetGrpcPort())
						} else {
							localAddress = fmt.Sprintf("%s:%d", config.HeartbeatClintIp(), DefaultGrpcPort)
						}
						for _, grayAddr := range entry.GrayAddress() {
							// 如果本地地址等于灰度地址，则返回异常，以便上游重试到正确的IP地址上
							if localAddress != grayAddr {
								// 请求转发一次，正常来说应该会路由到正确的节点，
								// 如果不可以则让上游重试
								var err error
								if req.ContentType() == client.DefaultContentType {
									// 默认使用RPC Client
									newRequest := client.NewRequest(req.Service(), req.Endpoint(), req)
									err = client.Call(ctx, newRequest, rsp)
								} else {
									// 此处为GRPC client
									gClient := grpc.NewClient()
									newRequest := gClient.NewRequest(req.Service(), req.Endpoint(), req.Body(), client.WithContentType(req.ContentType()))
									err = gClient.Call(ctx, newRequest, rsp)
								}
								return err
							}

						}
					}
				}
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
			entry, blockErr := sea.Entry(resourceName, sea.WithResourceType(base.ResTypeRPC), sea.WithTrafficType(base.Inbound))
			if blockErr != nil {
				if opts.serverBlockFallback != nil {
					return opts.streamServerBlockFallback(stream, blockErr)
				}
				stream.Send(blockErr)
				return stream
			}
			defer entry.Exit()
		}
		return stream
	}
}
