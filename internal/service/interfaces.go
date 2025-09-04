// Package service 应用服务模块
// 提供日报生成的核心业务逻辑，协调各个子模块完成完整的日报生成流程
package service

// Runner 应用服务运行器接口
// 定义了日报生成服务的核心运行方法
type Runner interface {
	// Run 运行日报生成服务
	// 执行完整的日报生成流程：收集提交 -> 处理diff -> 生成摘要 -> 发送通知
	// 返回错误信息，如果执行成功则返回nil
	Run() error
}