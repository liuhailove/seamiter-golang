package cache

import (
	"strconv"
	"testing"
)

func TestLRU_Add(t *testing.T) {
	c := NewLRUCacheMap(100)
	for i := 1; i <= 100; i++ {
		val := int64(i)
		c.Add(strconv.Itoa(i), &val)
	}
}
