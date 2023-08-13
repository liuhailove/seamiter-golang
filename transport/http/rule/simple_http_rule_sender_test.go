package rule

import (
	"fmt"
	"testing"
)

func TestRuleTypeStr(t *testing.T) {
	s := simpleHttpRuleSender{}
	fmt.Println(s.RuleTypeStr())
}
