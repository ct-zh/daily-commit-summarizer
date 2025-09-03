// Package git 提供Git操作相关的接口和实现
package git

// Collector 定义Git收集器接口
// 用于收集Git仓库中的分支和提交信息
type Collector interface {
	// FetchRemoteBranches 获取所有远程分支
	// 返回远程分支列表，格式为 "origin/branch-name"
	FetchRemoteBranches() ([]string, error)
	
	// CollectCommits 收集指定时间范围内的提交
	// since: 开始时间，支持Git时间格式如"midnight"、"2023-01-01"
	// until: 结束时间，支持Git时间格式如"now"、"2023-01-02"
	// 返回提交元数据列表，按时间顺序排列
	CollectCommits(since, until string) ([]CommitMeta, error)
}