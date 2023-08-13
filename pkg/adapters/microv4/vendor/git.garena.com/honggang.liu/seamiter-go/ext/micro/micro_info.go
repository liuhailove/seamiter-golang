package micro

var (
	_grpcPort int64 = 0
)

// SetGrpcPort 设置GRPC接口
func SetGrpcPort(grpcPort int64) {
	_grpcPort = grpcPort
}

// GetGrpcPort 获取Grpc接口
func GetGrpcPort() int64 {
	return _grpcPort
}
