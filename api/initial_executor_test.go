package api

import (
	"context"
	"fmt"
	metric_exporter "github.com/liuhailove/seamiter-golang/exporter/metric"
	"github.com/liuhailove/seamiter-golang/util"
	"log"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestDoInit(t *testing.T) {
	doInit()
}

func TestInitWithConfig(t *testing.T) {
	InitWithConfigFile("")
}

func TestRegister(t *testing.T) {
	l := &net.ListenConfig{Control: reusePortControl}
	sl, err := l.Listen(context.Background(), "tcp", ":8081")
	if err != nil {
		panic(fmt.Errorf("init metric exporter http server err: %s", err.Error()))
	}
	http.Handle("/metrics", metric_exporter.HTTPHandler())
	defer sl.Close()
	go util.RunWithRecover(func() {
		err = http.Serve(sl, nil)
		if err != nil {
			panic(err)
		}
	})

	l2 := &net.ListenConfig{Control: reusePortControl}
	s, err := l2.Listen(context.Background(), "tcp", "localhost:8081")
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	http.Handle("/metrics", metric_exporter.HTTPHandler())
	go util.RunWithRecover(func() {
		err = http.Serve(s, nil)
		if err != nil {
			panic(err)
		}
	})

	time.Sleep(time.Second * 100)

}

//func reusePortControl(network, address string, c syscall.RawConn) error {
//	var opErr error
//	err := c.Control(func(fd uintptr) {
//		// syscall.SO_REUSEPORT ,在Linux下还可以指定端口重用
//		opErr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
//	})
//	if err != nil {
//		return err
//	}
//	return opErr
//}
