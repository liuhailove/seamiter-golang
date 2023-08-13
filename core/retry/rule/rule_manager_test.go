package rule

import (
	"fmt"
	"github.com/liuhailove/seamiter-golang/core/retry"
	"github.com/liuhailove/seamiter-golang/util"
	"github.com/pkg/errors"
	"testing"
	"time"
)

func TestGetRules(t *testing.T) {
	var rule = new(Rule)
	rule.Resource = "test_resource"
	// simple
	//rule.RetryPolicy = SimpleRetryPolicy
	//rule.RetryMaxAttempts = 3
	//rule.RetryTimeout = 1000
	//
	//rule.BackoffPolicy = FixedBackOffPolicy
	//rule.FixedBackOffPeriodInMs = 2000
	//rule.BackoffDelay = 1000
	//rule.BackoffMaxDelay = 2000
	//rule.BackoffMultiplier = 2
	//
	//rule.UniformMinBackoffPeriod = 1000
	//rule.UniformMaxBackoffPeriod = 2000
	//
	//rule.ErrorMatcher = ExactMatch
	////rule.IncludeExceptions = []string{"error"}
	//rule.ExcludeExceptions = []string{"error1"}

	//// timeout
	//rule.RetryPolicy = TimeoutRtyPolicy
	//rule.RetryMaxAttempts = 3
	//rule.RetryTimeout = 10000
	//
	//rule.BackoffPolicy = FixedBackOffPolicy
	//rule.FixedBackOffPeriodInMs = 2000
	//rule.BackoffDelay = 1000
	//rule.BackoffMaxDelay = 2000
	//rule.BackoffMultiplier = 2
	//
	//rule.UniformMinBackoffPeriod = 1000
	//rule.UniformMaxBackoffPeriod = 2000
	//
	//rule.ErrorMatcher = ExactMatch
	//rule.IncludeExceptions = []string{"error"}
	////rule.ExcludeExceptions = []string{"error"}

	// MaxAttemptsRetryPolicy
	rule.RetryPolicy = MaxAttemptsRetryPolicy
	rule.RetryMaxAttempts = 3
	rule.RetryTimeout = 10000

	rule.BackoffPolicy = ExponentialBackOffPolicy
	rule.FixedBackOffPeriodInMs = 2000
	rule.BackoffDelay = 5000
	rule.BackoffMaxDelay = 20000
	rule.BackoffMultiplier = 2

	rule.UniformMinBackoffPeriod = 1000
	rule.UniformMaxBackoffPeriod = 2000

	rule.ErrorMatcher = ExactMatch
	rule.IncludeExceptions = []string{"error"}
	//rule.ExcludeExceptions = []string{"error"}

	var rules []*Rule
	rules = append(rules, rule)
	LoadRules(rules)

	var resRetryTemplate = getRetryTemplatesOfResource(rule.Resource)
	var result, err = resRetryTemplate.Execute(&MyTestRetryCallback{})
	if err == nil {
		fmt.Println(result)
	} else {
		fmt.Println(err)
	}

}

type MyTestRetryCallback struct {
}

func (m MyTestRetryCallback) DoWithRetry(content retry.RtyContext) interface{} {
	//fmt.Println(content.GetRetryCount())
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
