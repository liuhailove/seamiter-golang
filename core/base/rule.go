package base

import "fmt"

type SeaRule interface {
	fmt.Stringer
	ResourceName() string
}
