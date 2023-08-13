package reflection

import (
	"context"
	rpb "github.com/liuhailove/seamiter-golang/pkg/adapters/microv2/reflection/grpc_reflection_v1alpha"
	"github.com/micro/go-micro/v2/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"io"
	"log"
	"reflect"
	"sort"
	"strings"
	"unsafe"
)

const (
	FileDescsByNameField   = "descsByName"
	RpcServerHandlersField = "handlers"
)

// MicroServer micro服务
type MicroServer interface {
	server.Server
}

// Register registers the server reflection service on the given Micro server.
func Register(s MicroServer) {
	svr := NewServer(ServerOptions{Services: s})
	err := rpb.RegisterServerReflectionHandler(s, svr)
	if err != nil {
		log.Println(err)
		return
	}
}

// serverReflectionServer 反射服务
type serverReflectionServer struct {
	s            server.Server
	descResolver protodesc.Resolver
	extResolver  ExtensionResolver
}

// fileDescWithDependencies returns a slice of serialized fileDescriptors in
// wire format ([]byte). The fileDescriptors will include fd and all the
// transitive dependencies of fd with names not in sentFileDescriptors.
func (s *serverReflectionServer) fileDescWithDependencies(fd protoreflect.FileDescriptor, sentFileDescriptors map[string]bool) ([][]byte, error) {
	var r [][]byte
	queue := []protoreflect.FileDescriptor{fd}
	for len(queue) > 0 {
		currentfd := queue[0]
		queue = queue[1:]
		if sent := sentFileDescriptors[currentfd.Path()]; len(r) == 0 || !sent {
			sentFileDescriptors[currentfd.Path()] = true
			fdProto := protodesc.ToFileDescriptorProto(currentfd)
			currentfdEncoded, err := proto.Marshal(fdProto)
			if err != nil {
				return nil, err
			}
			r = append(r, currentfdEncoded)
		}
		for i := 0; i < currentfd.Imports().Len(); i++ {
			queue = append(queue, currentfd.Imports().Get(i))
		}
	}
	return r, nil
}

// fileDescEncodingContainingSymbol finds the file descriptor containing the
// given symbol, finds all of its previously unsent transitive dependencies,
// does marshalling on them, and returns the marshalled result. The given symbol
// can be a type, a service or a method.
func (s *serverReflectionServer) fileDescEncodingContainingSymbol(name string, sentFileDescriptors map[string]bool) ([][]byte, error) {
	d, err := s.descResolver.FindDescriptorByName(protoreflect.FullName(name))
	if err != nil {
		// 获取descsByName的非导出对象，这个map保存了注册的proto签名
		v := reflect.ValueOf(s.descResolver).Elem().FieldByName(FileDescsByNameField)
		// 获取非导出字段可寻址反射对象
		v = reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
		if !v.IsNil() && v.CanInterface() {
			if descsByName, ok := v.Interface().(map[protoreflect.FullName]interface{}); ok {
				for fullName := range descsByName {
					if strings.HasSuffix(string(fullName), name) {
						d, err = s.descResolver.FindDescriptorByName(fullName)
						break
					}
				}

			}
		}
		if err != nil {
			return nil, err
		}
	}
	return s.fileDescWithDependencies(d.ParentFile(), sentFileDescriptors)
}

// fileDescEncodingContainingExtension finds the file descriptor containing
// given extension, finds all of its previously unsent transitive dependencies,
// does marshalling on them, and returns the marshalled result.
func (s *serverReflectionServer) fileDescEncodingContainingExtension(typeName string, extNum int32, sentFileDescriptors map[string]bool) ([][]byte, error) {
	xt, err := s.extResolver.FindExtensionByNumber(protoreflect.FullName(typeName), protoreflect.FieldNumber(extNum))
	if err != nil {
		return nil, err
	}
	return s.fileDescWithDependencies(xt.TypeDescriptor().ParentFile(), sentFileDescriptors)
}

// allExtensionNumbersForTypeName returns all extension numbers for the given type.
func (s *serverReflectionServer) allExtensionNumbersForTypeName(name string) ([]int32, error) {
	var numbers []int32
	s.extResolver.RangeExtensionsByMessage(protoreflect.FullName(name), func(xt protoreflect.ExtensionType) bool {
		numbers = append(numbers, int32(xt.TypeDescriptor().Number()))
		return true
	})
	sort.Slice(numbers, func(i, j int) bool {
		return numbers[i] < numbers[j]
	})
	if len(numbers) == 0 {
		// maybe return an error if given type name is not known
		if _, err := s.descResolver.FindDescriptorByName(protoreflect.FullName(name)); err != nil {
			return nil, err
		}
	}
	return numbers, nil
}

// listServices returns the names of services this server exposes.
func (s *serverReflectionServer) listServices() []*rpb.ServiceResponse {
	// 获取server.Server的非导出对象，这个对象handler保存了注册的处理handler
	v := reflect.ValueOf(s.s).Elem().FieldByName(RpcServerHandlersField)
	// 获取非导出字段可寻址反射对象
	v = reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
	resp := make([]*rpb.ServiceResponse, 0)
	if !v.IsNil() && v.CanInterface() {
		if serviceHandlerMap, ok := v.Interface().(map[string]server.Handler); ok {
			for svc := range serviceHandlerMap {
				resp = append(resp, &rpb.ServiceResponse{Name: svc})
			}
			sort.Slice(resp, func(i, j int) bool {
				return resp[i].Name < resp[j].Name
			})
		}
	}
	return resp
}

// ServerOptions represents the options used to construct a reflection server.
//
// Experimental
//
// Notice: This type is EXPERIMENTAL and may be changed or removed in a
// later release.
type ServerOptions struct {
	Services server.Server

	// Optional resolver used to load descriptors. If not specified,
	// protoregistry.GlobalFiles will be used.
	DescriptorResolver protodesc.Resolver
	// Optional resolver used to query for known extensions. If not specified,
	// protoregistry.GlobalTypes will be used.
	ExtensionResolver ExtensionResolver
}

// NewServer returns a reflection server implementation using the given options.
// This can be used to customize behavior of the reflection service. Most usages
// should prefer to use Register instead.
//
// Experimental
//
// Notice: This function is EXPERIMENTAL and may be changed or removed in a
// later release.
func NewServer(opts ServerOptions) rpb.ServerReflectionHandler {
	if opts.DescriptorResolver == nil {
		opts.DescriptorResolver = protoregistry.GlobalFiles
	}
	if opts.ExtensionResolver == nil {
		opts.ExtensionResolver = protoregistry.GlobalTypes
	}
	return &serverReflectionServer{
		s:            opts.Services,
		descResolver: opts.DescriptorResolver,
		extResolver:  opts.ExtensionResolver,
	}
}

// ExtensionResolver is the interface used to query details about extensions.
// This interface is satisfied by protoregistry.GlobalTypes.
//
// Experimental
//
// Notice: This type is EXPERIMENTAL and may be changed or removed in a
// later release.
type ExtensionResolver interface {
	protoregistry.ExtensionTypeResolver
	RangeExtensionsByMessage(message protoreflect.FullName, f func(protoreflect.ExtensionType) bool)
}

func (s serverReflectionServer) ServerReflectionInfo(ctx context.Context, stream rpb.ServerReflection_ServerReflectionInfoStream) error {
	sentFileDescriptors := make(map[string]bool)
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		out := &rpb.ServerReflectionResponse{
			ValidHost:       in.Host,
			OriginalRequest: in,
		}
		switch req := in.MessageRequest.(type) {
		case *rpb.ServerReflectionRequest_FileByFilename:
			var b [][]byte
			fd, err := s.descResolver.FindFileByPath(req.FileByFilename)
			if err == nil {
				b, err = s.fileDescWithDependencies(fd, sentFileDescriptors)
			}
			if err != nil {
				out.MessageResponse = &rpb.ServerReflectionResponse_ErrorResponse{
					ErrorResponse: &rpb.ErrorResponse{
						ErrorCode:    int32(codes.NotFound),
						ErrorMessage: err.Error(),
					},
				}
			} else {
				out.MessageResponse = &rpb.ServerReflectionResponse_FileDescriptorResponse{
					FileDescriptorResponse: &rpb.FileDescriptorResponse{FileDescriptorProto: b},
				}
			}
		case *rpb.ServerReflectionRequest_FileContainingSymbol:
			b, err := s.fileDescEncodingContainingSymbol(req.FileContainingSymbol, sentFileDescriptors)
			if err != nil {
				out.MessageResponse = &rpb.ServerReflectionResponse_ErrorResponse{
					ErrorResponse: &rpb.ErrorResponse{
						ErrorCode:    int32(codes.NotFound),
						ErrorMessage: err.Error(),
					},
				}
			} else {
				out.MessageResponse = &rpb.ServerReflectionResponse_FileDescriptorResponse{
					FileDescriptorResponse: &rpb.FileDescriptorResponse{FileDescriptorProto: b},
				}
			}
		case *rpb.ServerReflectionRequest_FileContainingExtension:
			typeName := req.FileContainingExtension.ContainingType
			extNum := req.FileContainingExtension.ExtensionNumber
			b, err := s.fileDescEncodingContainingExtension(typeName, extNum, sentFileDescriptors)
			if err != nil {
				out.MessageResponse = &rpb.ServerReflectionResponse_ErrorResponse{
					ErrorResponse: &rpb.ErrorResponse{
						ErrorCode:    int32(codes.NotFound),
						ErrorMessage: err.Error(),
					},
				}
			} else {
				out.MessageResponse = &rpb.ServerReflectionResponse_FileDescriptorResponse{
					FileDescriptorResponse: &rpb.FileDescriptorResponse{FileDescriptorProto: b},
				}
			}
		case *rpb.ServerReflectionRequest_AllExtensionNumbersOfType:
			extNums, err := s.allExtensionNumbersForTypeName(req.AllExtensionNumbersOfType)
			if err != nil {
				out.MessageResponse = &rpb.ServerReflectionResponse_ErrorResponse{
					ErrorResponse: &rpb.ErrorResponse{
						ErrorCode:    int32(codes.NotFound),
						ErrorMessage: err.Error(),
					},
				}
			} else {
				out.MessageResponse = &rpb.ServerReflectionResponse_AllExtensionNumbersResponse{
					AllExtensionNumbersResponse: &rpb.ExtensionNumberResponse{
						BaseTypeName:    req.AllExtensionNumbersOfType,
						ExtensionNumber: extNums,
					},
				}
			}
		case *rpb.ServerReflectionRequest_ListServices:
			out.MessageResponse = &rpb.ServerReflectionResponse_ListServicesResponse{
				ListServicesResponse: &rpb.ListServiceResponse{
					Service: s.listServices(),
				},
			}
		default:
			return status.Errorf(codes.InvalidArgument, "invalid MessageRequest: %v", in.MessageRequest)
		}

		if err := stream.Send(out); err != nil {
			return err
		}
	}
}
