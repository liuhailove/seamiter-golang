package micro

import (
	"context"
	"fmt"
	sea "github.com/liuhailove/seamiter-golang/api"
	"github.com/liuhailove/seamiter-golang/core/base"
	"github.com/liuhailove/seamiter-golang/core/flow"
	"github.com/liuhailove/seamiter-golang/core/stat"
	proto "github.com/liuhailove/seamiter-golang/pkg/adapters/micro/test"
	"github.com/liuhailove/seamiter-golang/util"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/client/selector"
	"github.com/micro/go-micro/registry"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"log"
	"math/rand"
	"sort"
	"strings"
	"testing"
)

func TestClientLimiter(t *testing.T) {
	r := registry.NewRegistry()
	s := selector.NewSelector(selector.Registry(r))

	c := client.NewClient(
		client.Selector(s),
		client.Wrap(NewClientWrapper(
			WithClientBlockFallback(
				func(ctx context.Context, request client.Request, blockError *base.BlockError) error {
					return errors.New(FakeErrorMsg)
				}),
		)),
	)

	req := c.NewRequest("sea.test.server", "Test.Ping", &proto.Request{UserName: "honnggang.liu"}, client.WithContentType("application/json"))

	err := sea.InitDefault()
	if err != nil {
		log.Fatal(err)
	}

	rsp := &proto.Response{Result: "hello"}

	t.Run("success", func(t *testing.T) {
		var _, err = flow.LoadRules([]*flow.Rule{
			{
				Resource:               req.Method(),
				Threshold:              1.0,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
			},
		})
		assert.Nil(t, err)
		//var ri = mock.RuleItem{
		//	WhenParamIdx:            0,
		//	WhenParamKey:            "UserName",
		//	WhenParamValue:          "honnggang.liu",
		//	ControlBehavior:         mock.Mock,
		//	ThenReturnMockData:      `{"result":"Hello234"}`,
		//	ThenReturnWaitingTimeMs: 0,
		//	ThenThrowMsg:            "",
		//}
		//var _, err2 = mock.LoadRules([]*mock.Rule{{
		//	Resource:           "sea.test.server.Test.Ping",
		//	ControlBehavior:    mock.Mock,
		//	Strategy:           mock.Param,
		//	ThenReturnMockData: `{"result":"Hello"}`,
		//	SpecificItems:      []mock.RuleItem{ri},
		//}})
		//assert.Nil(t, err2)
		err = c.Call(context.TODO(), req, rsp)
		// No server started, the return err should not be nil
		assert.NotNil(t, err)
		assert.NotEqual(t, FakeErrorMsg, err.Error())
		assert.True(t, util.Float64Equals(1.0, stat.GetResourceNode(req.Method()).GetQPS(base.MetricEventPass)))

		t.Run("second fail", func(t *testing.T) {
			err := c.Call(context.TODO(), req, rsp)
			assert.EqualError(t, err, FakeErrorMsg)
			assert.True(t, util.Float64Equals(1.0, stat.GetResourceNode(req.Method()).GetQPS(base.MetricEventPass)))
		})
	})
}

func solve(sli []string) []string {
	len := len(sli)
	if len <= 1 {
		return sli
	}

	for i := len - 1; i > 0; i-- {
		randNum := rand.Intn(i)
		sli[i], sli[randNum] = sli[randNum], sli[i]
	}
	return sli
}

func TestName(t *testing.T) {

	//rand.Seed(time.Now().UnixNano())

	//for i := 0; i < 100; i++ {
	//	var grayAddress = []string{"A", "B", "C", "D", "E", "F", "G", "H"}
	//	fmt.Println(solve(grayAddress))
	//}
	//

	//var m map[string]string

	//var strs = []string{"A", "B", "B1", "B1*", "B11*", "B12*"}
	//sort.Slice(strs, func(i, j int) bool {
	//	return strs[i] > strs[j]
	//})
	////sort.Strings(strs)
	//fmt.Println(strs)
	var m = make(map[string]string)
	m["a"] = "ra"
	m["b"] = "rb"
	m["c"] = "rc"
	m["c*"] = "rca"
	m["cc*"] = "rccb"
	m["ccc*"] = "rcccc"
	m["*"] = "abcd"

	fmt.Println(find(m, "a"))
	fmt.Println(find(m, "b"))
	fmt.Println(find(m, "cc"))
	fmt.Println(find(m, "ccccc"))
	fmt.Println(find(m, "abc"))

}

func find(m map[string]string, name string) string {
	var tsc = m[name]
	if tsc != "" {
		return tsc
	}
	var ress []string
	for res := range m {
		ress = append(ress, res)
	}
	sort.Slice(ress, func(i, j int) bool {
		return ress[i] > ress[j]
	})
	for _, res := range ress {
		if res[len(res)-1] == '*' {
			var length = len(res)
			if length == 1 {
				return m[res]
			}
			if length > 1 && strings.HasPrefix(name, res[:length-1]) {
				return m[res]
			}
		}
	}
	return ""
}
