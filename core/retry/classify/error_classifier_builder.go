package classify

type ErrorClassifierBuilder struct {
	// IsWhiteList 构建符号类型. true：白名单,false：黑名单
	IsWhiteList *bool
	// Errors 错误集合
	Errors []error
}

func (e *ErrorClassifierBuilder) RetryOn(err error) *ErrorClassifierBuilder {
	if e.IsWhiteList != nil && !*e.IsWhiteList {
		panic("Please use only retryOn() or only notRetryOn()")
	}
	if e.IsWhiteList == nil {
		e.IsWhiteList = new(bool)
	}
	*e.IsWhiteList = true
	e.Errors = append(e.Errors, err)
	return e
}

func (e *ErrorClassifierBuilder) NotRetryOn(err error) *ErrorClassifierBuilder {
	if e.IsWhiteList != nil && *e.IsWhiteList {
		panic("Please use only retryOn() or only notRetryOn()")
	}
	if e.IsWhiteList == nil {
		e.IsWhiteList = new(bool)
	}
	*e.IsWhiteList = false
	e.Errors = append(e.Errors, err)
	return e
}

func (e *ErrorClassifierBuilder) Build(matcher PatternMatcher) *ErrorClassifier {
	var classifier = new(ErrorClassifier)
	classifier.SetClassified(e.Errors)
	if e.IsWhiteList != nil {
		classifier.SetDefaultValue(*e.IsWhiteList)
	}
	classifier.Matcher = matcher
	return classifier
}
