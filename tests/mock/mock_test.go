package mock

import (
	"encoding/json"
	"fmt"
	"github.com/liuhailove/seamiter-golang/api"
	"github.com/liuhailove/seamiter-golang/core/base"
	"github.com/liuhailove/seamiter-golang/core/config"
	"github.com/liuhailove/seamiter-golang/core/mock"
	"github.com/liuhailove/seamiter-golang/logging"
	"github.com/stretchr/testify/assert"
	"log"
	"regexp"
	"testing"
)

func Initsea() {
	// We should initialize sea first.
	conf := config.NewDefaultConfig()
	// for testing, logging output to console
	conf.Sea.Log.Logger = logging.NewConsoleLogger()
	conf.Sea.Log.Metric.FlushIntervalSec = 0
	conf.Sea.Stat.System.CollectIntervalMs = 0
	conf.Sea.Stat.System.CollectMemoryIntervalMs = 0
	conf.Sea.Stat.System.CollectCpuIntervalMs = 0
	conf.Sea.Stat.System.CollectLoadIntervalMs = 0
	err := api.InitWithConfig(conf)
	if err != nil {
		log.Fatal(err)
	}
}
func TestMockRule(t *testing.T) {
	Initsea()
	var rs1 = "TestMock1"
	var rs2 = "TestMock2"
	type Result struct {
		UserId   int64
		UserName string
		Code     int
	}

	var r = &Result{
		UserId:   100,
		UserName: "test-user",
		Code:     200,
	}
	var rD, _ = json.Marshal(r)

	rule1 := &mock.Rule{
		Resource:           rs1,
		ControlBehavior:    mock.Mock,
		ThenReturnMockData: string(rD),
	}
	rule2 := &mock.Rule{
		Resource:           rs2,
		ControlBehavior:    mock.Mock,
		ThenReturnMockData: string(rD),
	}
	_, err := mock.LoadRules([]*mock.Rule{rule1, rule2})
	if err != nil {
		panic(err)
	}

	for i := 0; i < 5; i++ {
		entry, blockError := api.Entry(rs1, api.WithTrafficType(base.Inbound))
		assert.Nil(t, blockError)
		if blockError != nil {
			t.Errorf("entry error:%+v", blockError)
		}
		if blockError == nil {
			entry.Exit()
		}
	}
}

func TestReplace(t *testing.T) {
	var emptyReg = regexp.MustCompile(`,\s*{\s*}`)
	var emptyReg2 = regexp.MustCompile(`,\s*]`)

	var str = `{
		"res": [{
			"is_order_set": true,
			"set_time": 1667905102,
			"error_code": 0,
			"order_id": 1
		}, { ss }, ],
		"common_result": {
			"err_code": 0,
			"err_msg": "success"
		}
	}`
	var replaceStr = emptyReg.ReplaceAllString(str, "")
	//fmt.Println(replaceStr)
	replaceStr = emptyReg2.ReplaceAllString(replaceStr, "]")
	fmt.Println(replaceStr)
	//	var str2 = `{
	//  "res": [
	//    {
	//      "is_order_set": true,
	//      "set_time": 1667905102,
	//      "error_code": 0,
	//      "order_id": 1
	//    },{
	//      "is_order_set": true,
	//      "set_time": 1667905102,
	//      "error_code": 0,
	//      "order_id": 1666664880
	//    }
	//  ],
	//  "common_result": {
	//    "err_code": 0,
	//    "err_msg": "success"
	//  }
	//}`
	//	var finalKeys = []string{"res", "[1]"}
	//	var changeData = jsonparser.Delete([]byte(str2), finalKeys...)
	//
	//	fmt.Println(string(changeData))

}

func TestDelete(t *testing.T) {
	//	var originMockData = `{
	//  "res": [
	//    {
	//      "is_order_set": true,
	//      "set_time": 1667905102,
	//      "error_code": 0,
	//      "order_id": 1
	//    },{
	//      "is_order_set": true,
	//      "set_time": 1667905102,
	//      "error_code": 0,
	//      "order_id": 1666664880
	//    }
	//  ],
	//  "common_result": {
	//    "err_code": 0,
	//    "err_msg": "success"
	//  }
	//}`
	//
	//	// 移除mock中多余的数据
	//	_, err = jsonparser.ArrayEach([]byte(originMockData), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
	//		if deleteIndex < index {
	//			deleteIndex++
	//			return
	//		}
	//		var arrIndex = fmt.Sprintf("[%d]", deleteIndex)
	//		// 下标不需要变更，删除后数据下标会前移
	//		// deleteIndex++
	//		var finalKeys = keysPre
	//		finalKeys = append(finalKeys, arrIndex)
	//		// 为了预防下标越界
	//		_, _, _, err = jsonparser.Get([]byte(item.ThenReturnMockData), finalKeys...)
	//		if err == nil {
	//			changeData = jsonparser.Delete([]byte(item.ThenReturnMockData), finalKeys...)
	//			var replaceStr = emptyReg.ReplaceAllString(string(changeData), "")
	//			item.ThenReturnMockData = replaceStr
	//			changeData = []byte(replaceStr)
	//		} else {
	//			logging.Warn("get property failed in ArrayEach", "property", item.WhenParamKey, "thenReturnMockData", item.ThenReturnMockData, "request data", requestJsonData, "err", err)
	//		}
	//	}, keysPre...)
}
