package micro

import (
	"context"
	"errors"
	"github.com/liuhailove/seamiter-golang/core/base"
	"github.com/liuhailove/seamiter-golang/core/flow"
	"github.com/liuhailove/seamiter-golang/core/stat"
	"github.com/liuhailove/seamiter-golang/util"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/selector"
	"github.com/micro/go-micro/v2/registry/memory"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"

	proto "github.com/liuhailove/seamiter-golang/pkg/adapters/microv2/test"

	sea "github.com/liuhailove/seamiter-golang/api"
)

func TestClientLimiter(t *testing.T) {
	// setup
	r := memory.NewRegistry()
	s := selector.NewSelector(selector.Registry(r))

	c := client.NewClient(
		// set the selector
		client.Selector(s),
		// add the breaker wrapper
		client.Wrap(NewClientWrapper(
			// add custom fallback function to return a fake error for assertion
			WithClientBlockFallback(
				func(ctx context.Context, request client.Request, blockError *base.BlockError) error {
					return errors.New(FakeErrorMsg)
				}),
		)),
	)

	req := c.NewRequest("sea.test.server", "Test.Ping", &proto.Request{}, client.WithContentType("application/json"))

	err := sea.InitDefault()
	if err != nil {
		log.Fatal(err)
	}

	rsp := &proto.Response{}

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
