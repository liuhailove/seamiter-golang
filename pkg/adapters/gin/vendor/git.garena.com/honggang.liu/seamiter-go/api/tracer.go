package api

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"github.com/pkg/errors"
)

// TraceError records the provided error to the given seaEntry.
func TraceError(entry *base.SeaEntry, err error) {
	defer func() {
		if e := recover(); e != nil {
			logging.Error(errors.Errorf("%+v", e), "Failed to api.TraceError()")
			return
		}
	}()

	if entry == nil || err == nil {
		return
	}

	entry.SetError(err)
}
