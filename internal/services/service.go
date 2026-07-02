// Package services 核心业务逻辑处理层
//
// 如果一个接口存在有多个业务微服务的rpc调用，那么将在这一层做处理。
// handlers -> services -> 下游业务服务rpc调用。
//
// 禁止在 handlers 直接调用下游业务服务。
package services

// Init 初始化所有服务层
func Init() {
	InitUser()
}
