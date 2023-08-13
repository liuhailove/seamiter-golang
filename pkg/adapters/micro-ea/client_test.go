package micro

import (
	"context"
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
