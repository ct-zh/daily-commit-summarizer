// Package diff 提供代码差异处理功能
// 负责获取Git提交的diff内容，并将大型diff拆分为可管理的片段
package diff

import "context"

// Processor 定义Diff处理器接口
// 负责获取提交的代码差异并进行分片处理
type Processor interface {
	// GetDiff 获取指定提交的代码差异
	// sha: 提交的SHA值
	// 返回: diff内容字符串和可能的错误
	GetDiff(ctx context.Context, sha string) (string, error)
	
	// SplitPatchByFile 将整个diff按文件拆分
	// patch: 完整的diff内容
	// 返回: 按文件拆分的diff片段数组
	SplitPatchByFile(patch string) []string
	
	// ChunkBySize 将大型diff按大小拆分为多个片段
	// parts: 按文件拆分的diff片段
	// limit: 每个片段的最大字符数限制
	// 返回: 按大小拆分的diff片段数组
	ChunkBySize(parts []string, limit int) []string
	
	// GetParentSha 获取提交的父提交SHA
	// sha: 提交的SHA值
	// 返回: 父提交SHA和可能的错误
	GetParentSha(ctx context.Context, sha string) (string, error)
}

// Config 定义Diff模块配置接口
type Config interface {
	// GetExcludePatterns 获取排除的文件模式
	GetExcludePatterns() []string
	
	// GetMaxChunkSize 获取单个diff片段的最大字符数
	GetMaxChunkSize() int
	
	// GetRepoPath 获取仓库路径
	GetRepoPath() string
}