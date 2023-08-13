package util

import (
	"errors"
	"fmt"
	"testing"
)

func TestTryCatchFinally(t *testing.T) {
	Try(func() {
		fmt.Println("hello world")
		panic(errors.New("panic zero"))
	}).CatchAll(func(err error) {
		fmt.Println("catch all error", err.Error())
	}).Finally(func() {
		fmt.Println("finally handler")
	})
}
