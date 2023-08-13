package api

// InitialFunc 初始化Func
type InitialFunc interface {
	//Initial 初始化
	Initial() error
	//Order 排序
	Order() int
}
