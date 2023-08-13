package support

import (
	"errors"
	"fmt"
	"github.com/liuhailove/seamiter-golang/core/retry"
	"github.com/liuhailove/seamiter-golang/core/retry/classify"
	"github.com/liuhailove/seamiter-golang/util"
	"testing"
	"time"
)

func TestRetryTemplateBuilder(t *testing.T) {
	var retryTemplate = NewRetryTemplateBuilder().
		MaxAttemptsRtyPolicy(5).
		//FixedBackoff(1000).
		//ExponentialBackoffWithRandom(1000, 2, 5000, true).
		//WithinMillisRtyPolicy(1000).
		//InfiniteRtyPolicy().
		UniformRandomBackoff(100, 1000).
		//NotRetryOn(errors.New("error")).
		RetryOn(errors.New("hello world")).
		RetryOn(errors.New("error")).
		WithErrorMatchPattern(classify.RegularMatch).
		Build()

	var result, err = retryTemplate.Execute(&MyTestRetryCallback{})
	if err == nil {
		fmt.Println(result)
	} else {
		fmt.Println(err)
	}

}

type MyTestRetryCallback struct {
}

func (m MyTestRetryCallback) DoWithRetry(content retry.RtyContext) interface{} {
	fmt.Println(content.GetRetryCount())
	var result, err = PrintHello()
	if err != nil {
		panic(err.Error())
	}
	return result
}

var i = 0

func PrintHello() (string, error) {
	if i < 2 {
		i++
		if i == 1 {
			panic("error")
		}
		util.Sleep(time.Millisecond * 100)
		fmt.Println("PrintHello error")
		return "", errors.New("error")
	} else {
		//fmt.Println("hello")
		return "hello world", nil
	}
}
