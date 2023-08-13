package util

import (
	"github.com/liuhailove/seamiter-golang/logging"
	"github.com/pkg/errors"
)

func RunWithRecover(f func()) {
	defer func() {
		if err := recover(); err != nil {
			logging.Error(errors.Errorf("%+v", err), "Unexpected panic in util.RunWithRecover()")
		}
	}()
	f()
}
