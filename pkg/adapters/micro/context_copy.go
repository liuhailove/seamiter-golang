package micro

//import (
//	"context"
//	"reflect"
//	"runtime/debug"
//	"sync"
//	"unsafe"
//)
//
//type iface struct {
//	itab, data uintptr
//}
//
//type valueCtx struct {
//	context.Context
//	key, val interface{}
//}
//
//// A canceler is a context type that can be canceled directly. The
//// implementations are *cancelCtx and *timerCtx.
//type canceler interface {
//	cancel(removeFromParent bool, err error)
//	Done() <-chan struct{}
//}
//
//// A cancelCtx can be canceled. When canceled, it also cancels any children
//// that implement canceler.
//type cancelCtx struct {
//	context.Context
//
//	mu       sync.Mutex            // protects following fields
//	done     chan struct{}         // created lazily, closed by first cancel call
//	children map[canceler]struct{} // set to nil by the first cancel call
//	err      error                 // set to non-nil by the first cancel call
//}
//
//func GetKeyValues(ctx context.Context) map[interface{}]interface{} {
//	m := make(map[interface{}]interface{})
//	getKeyValue(ctx, m)
//	return m
//}
//
//func getKeyValue(ctx context.Context, m map[interface{}]interface{}) {
//
//	rtType := reflect.TypeOf(ctx).String()
//
//	// 遍历到顶级类型，直接过滤
//	if rtType == "*context.emptyCtx" {
//		return
//	}
//
//	ictx := *(*iface)(unsafe.Pointer(&ctx))
//	if ictx.data == 0 {
//		return
//	}
//	valCtx := (*valueCtx)(unsafe.Pointer(ictx.data))
//	if valCtx != nil && valCtx.key != nil && valCtx.val != nil {
//		t := reflect.TypeOf(valCtx.key)
//		k := t.Kind() // 获取的是值的种类
//		t.IsVariadic()
//
//		m[valCtx.key] = valCtx.val
//	}
//	getKeyValue(valCtx.Context, m)
//}
//
//func CopyCtxClearCancel(ctx context.Context) context.Context {
//
//	// ---------------------------
//	//		崩溃保护
//	// ---------------------------
//	defer func() {
//		if r := recover(); r != nil {
//			debug.PrintStack()
//		}
//	}()
//
//	newCtx := context.Background()
//	rtMap := GetKeyValues(ctx)
//	for k, v := range rtMap {
//		newCtx = context.WithValue(newCtx, k, v)
//	}
//	return newCtx
//}
