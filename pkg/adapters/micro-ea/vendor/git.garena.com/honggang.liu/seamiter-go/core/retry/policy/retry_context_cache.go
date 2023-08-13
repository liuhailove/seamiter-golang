package policy

import "git.garena.com/honggang.liu/seamiter-go/core/retry"

//RtyContextCache 上下文重试cache
type RtyContextCache interface {

	// Get 根据key获取重试上下文
	Get(key interface{}) retry.RtyContext

	// Put 把Key，ctx加入缓存
	Put(key interface{}, ctx retry.RtyContext)

	// Remove 从cache中移除
	Remove(key interface{})

	// ContainsKey 判断cache中是否包含key
	ContainsKey(key interface{}) bool
}
