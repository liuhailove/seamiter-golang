package gin

import (
	"bytes"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	sea "github.com/liuhailove/seamiter-golang/api"
	"github.com/liuhailove/seamiter-golang/core/base"
	"github.com/liuhailove/seamiter-golang/core/config"
)

const (
	MockRspMaxBodyLen = 8092
)

// bodyWriter body写入
type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// SeaMiddleware returns new gin.HandlerFunc
// Default resource name is {method}:{path}, such as "GET:/api/users/:id"
// Default block fallback is returning 429 code
// Define your own behavior by setting options
func SeaMiddleware(opts ...Option) gin.HandlerFunc {
	options := evaluateOptions(opts)
	return func(c *gin.Context) {
		if !config.CloseAll() {
			resourceName := c.Request.Method + ":" + c.FullPath()

			if options.resourceExtract != nil {
				resourceName = options.resourceExtract(c)
			}
			var excludedPath = false
			// 排除
			if options.excludePath != nil {
				excludedPath = options.excludePath(c)
			}
			if !excludedPath {
				blw := &bodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
				c.Writer = blw
				var method = c.Request.Method
				var entry *base.SeaEntry
				var err *base.BlockError
				var header = make(map[string][]string, 0)
				for k, v := range c.Request.Header {
					header[k] = v
				}
				if strings.ToUpper(method) == "GET" {
					entry, err = sea.Entry(
						resourceName,
						sea.WithResourceType(base.ResTypeWeb),
						sea.WithTrafficType(base.Inbound),
						sea.WithArgs(GetQueryParams(c)...),
						sea.WithHeaders(header))
				} else if strings.ToUpper(method) == "POST" {
					entry, err = sea.Entry(
						resourceName,
						sea.WithResourceType(base.ResTypeWeb),
						sea.WithTrafficType(base.Inbound),
						sea.WithArgs(GetPostFormParams(c)...),
						sea.WithHeaders(header))
				} else {
					entry, err = sea.Entry(
						resourceName,
						sea.WithResourceType(base.ResTypeWeb),
						sea.WithTrafficType(base.Inbound))
				}
				if err != nil {
					if err.BlockType() == base.BlockTypeMock {
						c.Data(http.StatusOK, "application/json", []byte(err.TriggeredValue().(string)))
						c.Abort()
					} else if err.BlockType() == base.BlockTypeMockError {
						c.Data(http.StatusInternalServerError, "application/json", []byte(err.TriggeredValue().(string)))
						c.Abort()
					} else if options.blockFallback != nil {
						options.blockFallback(c)
					} else {
						c.AbortWithStatus(http.StatusTooManyRequests)
					}
					return
				}
				entry.WhenExit(func(entry *base.SeaEntry, ctx *base.EntryContext) error {
					if c.Writer.Status() != http.StatusOK {
						return nil
					}
					if ctx == nil || ctx.Output == nil || ctx.Output.Rsps == nil || len(ctx.Output.Rsps) != 0 {
						return nil
					}
					if blw.body.Len() > MockRspMaxBodyLen {
						return nil
					}
					ctx.Output.Rsps = append(ctx.Output.Rsps, blw.body.String())
					return nil
				})
				defer entry.Exit()
			}
		}
		c.Next()
	}
}

// GetQueryParams 获取全部的get请求参数
func GetQueryParams(c *gin.Context) []interface{} {
	query := c.Request.URL.Query()
	var queryArgs = make([]interface{}, 0)
	for k := range query {
		queryArgs = append(queryArgs, k+"="+c.Query(k))
	}
	return queryArgs
}

// GetPostFormParams 获取全部的Post请求参数
func GetPostFormParams(c *gin.Context) []interface{} {
	var postArgs = make([]interface{}, 0)
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		if !errors.Is(err, http.ErrNotMultipart) {
			return postArgs
		}
	}
	for k, v := range c.Request.PostForm {
		if len(v) > 1 {
			postArgs = append(postArgs, k+"="+v[0])
		} else if len(v) == 1 {
			postArgs = append(postArgs, k+"="+v[0])
		}
	}

	return postArgs
}
