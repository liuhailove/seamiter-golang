module github.com/liuhailove/seamiter-golang/pkg/datasource/etcdv3

go 1.14

replace github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

replace github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.0.0

require (
	github.com/liuhailove/seamiter-golang v0.0.117-beta
	github.com/coreos/bbolt v1.3.4 // indirect
	github.com/coreos/etcd v3.3.25+incompatible
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.1
)
