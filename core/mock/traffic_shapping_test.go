package mock

import (
	"encoding/json"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/fatih/structs"
	"strings"
	"testing"
)

//func TestName(t *testing.T) {
//	b := &baseTrafficShapingController{}
//	var specificItems = []RuleItem{}
//	var input=
//	var ctx = &base.EntryContext{specificItems:specificItems}
//	b.ArgsCheck(ctx)
//}
type Req1 struct {
	OrderId []string `json:"orderId"`

	Req2 Req2 `json:"req2"`
}

type Req2 struct {
	UserName string `json:"userName"`
}

func TestClearRules2(t *testing.T) {
	req := &Req1{
		OrderId: []string{"123456", "56778"},
		Req2:    Req2{UserName: "liuhg"},
	}
	//structs.Name("OrderId")
	//var names = structs.New(req)
	//var fields = structs.Fields(req)
	//for _, f := range fields {
	//	val := f.Tag("json")
	//	if val == "req2" {
	//		filedsNew := f.Fields()
	//		for _, fI := range filedsNew {
	//			if fI.Tag("json") == "userName" {
	//				fI.Set("hello")
	//			}
	//		}
	//	}
	//}
	var replaceArr = []string{"ABCD"}
	WalkAndSetTest(structs.Fields(req), []string{"orderId"}, 0, replaceArr)
	fmt.Println(req)
	//fmt.Println(names)
	//fmt.Println(fields)
}

func WalkAndSetTest(fields []*structs.Field, properties []string, pos int, val interface{}) {
	if len(fields) == 0 {
		return
	}
	if pos > len(properties) {
		return
	}
	var property = properties[pos]
	for _, field := range fields {
		if field.Tag("json") == property {
			if pos == len(properties)-1 {
				err := field.Set(val)
				if err != nil {
					fmt.Println("err")
				}
				return
			}
			WalkAndSetTest(field.Fields(), properties, pos+1, val)
		}
	}
}
func TestClearRules(t *testing.T) {
	item := &RuleItem{
		ReplaceAttribute:   "res.[*].order_id",
		ThenReturnMockData: "{\"res\":[{\"is_order_set\":false,\"set_time\":0,\"error_code\":0,\"order_id\":\"1666664880\"},{\"is_order_set\":false,\"set_time\":0,\"error_code\":0,\"order_id\":\"1666664888\"},{\"is_order_set\":false,\"set_time\":0,\"error_code\":0,\"order_id\":\"1666664889\"}]}",
	}
	req := &Req1{
		OrderId: []string{"123456", "56778"},
	}
	jsonData, err := json.Marshal(req)
	var key = "orderId"
	var propertyArr = strings.Split(key, ".")
	valBytes, dt, _, _ := jsonparser.Get(jsonData, propertyArr...)
	var valS string
	if dt == jsonparser.Array || dt == jsonparser.Boolean {
		valS = fmt.Sprint(``, string(valBytes), ``)
	} else {
		valS = fmt.Sprint(`"`, string(valBytes), `"`)
	}
	var rArr = strings.Split(item.ReplaceAttribute, ".")

	if dt == jsonparser.Array && strings.Contains(item.ReplaceAttribute, "[*].") {
		var keysPreBreak = false
		var keysPre []string
		var keysPost []string
		for _, r := range rArr {
			if r == "[*]" {
				keysPreBreak = true
				continue
			}
			if !keysPreBreak {
				keysPre = append(keysPre, r)
			} else {
				keysPost = append(keysPost, r)
			}
		}
		var index = 0
		_, err := jsonparser.ArrayEach(jsonData, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			var arrIndex = fmt.Sprintf("[%d]", index)
			index++
			var finalKeys = keysPre
			finalKeys = append(finalKeys, arrIndex)
			finalKeys = append(finalKeys, keysPost...)
			_, _, _, err = jsonparser.Get([]byte(item.ThenReturnMockData), finalKeys...)
			if err == nil {

				changeData, err := jsonparser.Set([]byte(item.ThenReturnMockData), value, finalKeys...)
				fmt.Println(err)
				item.ThenReturnMockData = string(changeData)
				fmt.Println(string(changeData))

			}
		}, propertyArr...)
		if err != nil {
			return
		}
		fmt.Println("----------------")
		fmt.Println(item.ThenReturnMockData)

		//_, err := jsonparser.ArrayEach([]byte(item.ThenReturnMockData), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		//	var arrIndex = fmt.Sprintf("[%d]", index)
		//	var finalKeys = keysPre
		//	finalKeys = append(finalKeys, arrIndex)
		//	finalKeys = append(finalKeys, keysPost...)
		//	rData, _, _, _ := jsonparser.Get(value, keysPost...)
		//	changeData, err := jsonparser.Set([]byte(item.ThenReturnMockData), rData, finalKeys...)
		//	fmt.Println(changeData)
		//
		//}, keysPre...)
		//if err != nil {
		//	return
		//}

	}
	changeData, err := jsonparser.Set([]byte(item.ThenReturnMockData), []byte(valS), strings.Split(item.ReplaceAttribute, ".")...)
	if err != nil {
		fmt.Println(err)
		fmt.Println(string(changeData))
	}
	//fmt.Println(string(changeData))

}

//func TestClearRulesOfResource(t *testing.T) {
//	var ThenReturnMockData = "{\"userId\":1234,\"timeStamp\":1234}"
//	fmt.Println(time.Now().Unix())
//	ThenReturnMockData = strings.ReplaceAll(ThenReturnMockData, TimestampFunc, strconv.FormatInt(time.Now().UnixNano(), 10))
//	fmt.Println(ThenReturnMockData)
//}

func TestBaseTrafficShapingController_ArgsCheck(t *testing.T) {
	//for i := 0; i < 100; i++ {
	//	put(fmt.Sprintf("i_%d", i))
	//}
	//
	//var myMap = sync.Map{}
	//myMap.Store("A", nil)
	//_, exist := myMap.Load("B")
	//fmt.Println(exist)

	var myMayp = make(map[string][]string, 0)
	//myMayp["A"] = []string{"1", "2", "3"}
	//myMayp["B"] = []string{"4", "5", "6"}
	data, _ := json.Marshal(myMayp)
	fmt.Println(string(data))
}

//func put(result string) bool {
//	var requests, exist = holdRequestMap.LoadOrStore("a", []string{})
//	fmt.Println(exist)
//	if reqs, ok := requests.([]string); ok {
//		reqs = append(reqs, result)
//		holdRequestMap.Store("a", reqs)
//	}
//	return true
//}

func Test2k(t *testing.T) {
	var str = `{"RiskReqHeader":{"Version":1,"AppID":"kredit","SceneID":10011,"FlowNo":"0a957926ebc211eda5c616b8ff17ee7f-10011","ReqTime":1683345433,"ReqIp":"100.95.12.227","ReqServer":"credit_usercenter_apply","ReqMemo":"","ReqNo":"1765975759848996865|1765975759874162689|10011|1683345433","WorkflowSceneID":0,"ReqType":0,"ReqAsync":1,"WorkflowFlowNo":"","Env":""}}`
	var findObj, _, _, err = jsonparser.Get([]byte(str), "RiskReqHeader")
	var result = string(findObj)
	fmt.Println(result)
	//var str2 = fmt.Sprint(`"`, findStr, `"`)
	fmt.Println(err)
	//fmt.Println(findStr)
	//fmt.Println(str2)
}

func Test3(t *testing.T) {
	//var s = []string{"123456"}
	//var sc = s[0:1]
	//fmt.Println(sc)
	var val = []byte("123")
	fmt.Println(fmt.Sprint(``, string(val), ``))

	fmt.Println(fmt.Sprint(`"`, string(val), `"`))

}
