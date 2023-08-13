package etcdv3

import (
	"fmt"
	"github.com/liuhailove/seamiter-golang/ext/datasource"
	"github.com/coreos/etcd/clientv3"
	"github.com/stretchr/testify/mock"
	"log"
	"testing"
	"time"
)

//Test_ClientWithOneDatasource  New one datasource based on etcv3 client
func Test_ClientWithOneDatasource(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 10 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
		return
	}
	defer cli.Close()

	h := datasource.MockPropertyHandler{}
	h.On("isPropertyConsistent", mock.Anything).Return(true)
	h.On("Handle", mock.Anything).Return(nil)
	ds, err := NewDataSource(cli, "foo", &h)
	if err != nil {
		log.Fatal(err)
	}
	err = ds.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(20 * time.Second)
	fmt.Println("Prepare to close")
	ds.Close()
	time.Sleep(120 * time.Second)
}

//Test_ClientWithOneDatasource  New one datasource based on etcv3 client
func Test_ClientWithOneDatasource2(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 10 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
		return
	}
	defer cli.Close()
}
