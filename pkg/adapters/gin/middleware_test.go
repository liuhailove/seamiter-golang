package gin

import (
	"github.com/gin-gonic/gin"
	sea "github.com/liuhailove/seamiter-golang/api"
	"github.com/liuhailove/seamiter-golang/core/circuitbreaker"
	"github.com/liuhailove/seamiter-golang/core/config"
	"github.com/liuhailove/seamiter-golang/core/flow"
	"github.com/liuhailove/seamiter-golang/core/hotspot"
	"github.com/liuhailove/seamiter-golang/core/system"
	"github.com/liuhailove/seamiter-golang/logging"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func initsea(t *testing.T) {
	// 使用console输出日志
	conf := config.NewDefaultConfig()
	// for testing, logging output to console
	conf.Sea.Log.Logger = logging.NewConsoleLogger()
	err := sea.InitWithConfig(conf)
	//err := sea.InitDefault()
	if err != nil {
		t.Fatalf("Upexpected error: %+v", err)
	}
	_, err = flow.LoadRules([]*flow.Rule{
		{
			Resource:               "GET:/ping",
			Threshold:              1.0,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			StatIntervalInMs:       1000,
		},
		{
			Resource:               "/api/users/:id",
			Threshold:              20.0,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			StatIntervalInMs:       1000,
		},
	})
	specific := make(map[interface{}]int64)
	specific["sss"] = 1
	specific["123"] = 3
	_, err = hotspot.LoadRules([]*hotspot.Rule{{
		ID:                "1",
		Resource:          "/api/jobs/:id",
		MetricType:        hotspot.Concurrency,
		ControlBehavior:   hotspot.Reject,
		ParamIdx:          0,
		Threshold:         100.0,
		MaxQueueingTimeMs: 0,
		BurstCount:        10,
		DurationInSec:     1,
		SpecificItems:     specific,
	}})
	_, err = system.LoadRules([]*system.Rule{{
		ID:           "1",
		MetricType:   system.Load,
		TriggerCount: 100,
		Strategy:     system.BBR,
	}})
	if err != nil {
		t.Fatalf("Unexpected error: %+v", err)
		return
	}
	_, err = circuitbreaker.LoadRules([]*circuitbreaker.Rule{
		{
			Resource:         "abc01",
			Strategy:         circuitbreaker.SlowRequestRatio,
			MinRequestAmount: 5,
			StatIntervalMs:   1000,
			MaxAllowedRtMs:   20,
			Threshold:        0.1,
		},
		{
			Resource:         "abc02",
			Strategy:         circuitbreaker.ErrorRatio,
			MinRequestAmount: 5,
			StatIntervalMs:   1000,
			MaxAllowedRtMs:   20,
			Threshold:        100,
		},
		{
			Resource:         "abc03",
			Strategy:         circuitbreaker.ErrorCount,
			MinRequestAmount: 5,
			StatIntervalMs:   1000,
			MaxAllowedRtMs:   20,
			Threshold:        100,
			ProbeNum:         15,
		},
	})
	if err != nil {
		t.Fatalf("Unexpected error: %+v", err)
		return
	}
}

func TestseaMiddleware(t *testing.T) {
	type args struct {
		opts    []Option
		method  string
		path    string
		reqPath string
		handler func(ctx *gin.Context)
		body    io.Reader
	}

	type want struct {
		code int
	}

	var (
		tests = []struct {
			name string
			args args
			want want
		}{
			{
				name: "default get",
				args: args{
					opts:    []Option{},
					method:  http.MethodGet,
					path:    "/ping",
					reqPath: "/ping",
					handler: func(ctx *gin.Context) {
						ctx.String(http.StatusOK, "ping")
					},
					body: nil,
				},
				want: want{
					code: http.StatusOK,
				},
			},
			{
				name: "customize resource extract",
				args: args{
					opts: []Option{
						WithResourceExtractor(func(ctx *gin.Context) string {
							return ctx.FullPath()
						}),
					},
					method:  http.MethodPost,
					path:    "/api/users/:id",
					reqPath: "/api/users/123",
					handler: func(ctx *gin.Context) {
						ctx.String(http.StatusOK, "ping")
					},
					body: nil,
				},
				want: want{
					code: http.StatusTooManyRequests,
				},
			},
			{
				name: "customize block fallback",
				args: args{
					opts: []Option{
						WithBlockFallback(func(ctx *gin.Context) {
							ctx.String(http.StatusBadRequest, "block")
						}),
					},
					method:  http.MethodGet,
					path:    "/ping",
					reqPath: "/ping",
					handler: func(ctx *gin.Context) {
						ctx.String(http.StatusOK, "ping")
					},
					body: nil,
				},
				want: want{
					code: http.StatusBadRequest,
				},
			},
		}
	)
	initsea(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			//router.Use(seaMiddleware(tt.args.opts...))
			router.Handle(tt.args.method, tt.args.path, tt.args.handler)
			r := httptest.NewRequest(tt.args.method, tt.args.reqPath, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			assert.Equal(t, tt.want.code, w.Code)
		})

	}
	time.Sleep(time.Second * 600)
}
