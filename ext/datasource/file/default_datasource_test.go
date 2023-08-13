package file

import (
	"github.com/liuhailove/seamiter-golang/ext/datasource/util"
	"testing"
)

func TestName(t *testing.T) {
	Initialize()
	util.GetSystemSource().Write([]byte(TestSystemRules))
}
