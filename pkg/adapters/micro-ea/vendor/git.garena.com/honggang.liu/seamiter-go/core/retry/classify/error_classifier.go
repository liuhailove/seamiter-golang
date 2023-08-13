package classify

import (
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"regexp"
	"strings"
)

type PatternMatcher int32

const (
	ExactMatch   PatternMatcher = iota // 精确匹配
	PrefixMatch                        // 前缀匹配
	SuffixMatch                        // 后缀匹配
	ContainMatch                       // 包含匹配
	RegularMatch                       // 正则匹配
	AnyMatch                           // 只要不为空，则匹配
)

// ErrorClassifier 异常分类
type ErrorClassifier struct {
	DefaultValue bool
	Classified   []error
	// 匹配模式
	Matcher PatternMatcher
}

func (e *ErrorClassifier) SetDefaultValue(defaultValue bool) {
	e.DefaultValue = defaultValue
}

func (e *ErrorClassifier) SetClassified(errs []error) {
	for _, err := range errs {
		e.Classified = append(e.Classified, err)
	}
}

// Classify 返回map中是否包含此错误
func (e *ErrorClassifier) Classify(err error) bool {
	if err == nil {
		return e.DefaultValue
	}
	// 只要err不为空，则匹配
	if e.Matcher == AnyMatch {
		return true
	}
	var errMsg = err.Error()
	var result = false
	for _, errorC := range e.Classified {
		var key = errorC.Error()
		switch e.Matcher {
		case ExactMatch:
			result = key == errMsg
		case PrefixMatch:
			result = strings.HasPrefix(errMsg, key)
		case ContainMatch:
			result = strings.Contains(errMsg, key)
		case SuffixMatch:
			result = strings.HasSuffix(errMsg, key)
		case RegularMatch:
			// 首先断言val为字符串，如果不为字符串，则直接跳出，否则进行正则匹配,如果匹配错误，则尝试包含匹配
			if ok, err := regexp.MatchString(key, errMsg); err == nil {
				result = ok
			} else {
				logging.Warn("error regular match error,then try contains match", "error", err)
				result = strings.Contains(errMsg, key)
			}
		default:
			result = strings.Contains(errMsg, key)
		}
		if result {
			return e.DefaultValue
		}
	}
	return !e.DefaultValue
}
