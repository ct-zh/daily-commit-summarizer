// internal/prompt/interfaces.go
package prompt

import (
	"daily-commit-summarizer/internal/git"
)

// Builder 定义提示词构建器接口
type Builder interface {
	// CommitChunkPrompt 为单个diff片段构建提示词
	CommitChunkPrompt(meta git.CommitMeta, partIdx, total int, patch string) string
	
	// CommitMergePrompt 为合并多个diff片段的摘要构建提示词
	CommitMergePrompt(meta git.CommitMeta, parts []string) string
	
	// DailyMergePrompt 为生成每日报告构建提示词
	DailyMergePrompt(dateLabel string, items []CommitSummary, repo string) string
}

// CommitSummary 存储提交摘要
type CommitSummary struct {
	Meta    git.CommitMeta
	Summary string
}