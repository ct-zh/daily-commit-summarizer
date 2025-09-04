// internal/prompt/builder_test.go
package prompt

import (
	"strings"
	"testing"
	
	"daily-commit-summarizer/internal/git"
	"github.com/stretchr/testify/assert"
)

func TestDefaultBuilder_CommitChunkPrompt(t *testing.T) {
	builder := NewBuilder()
	
	meta := git.CommitMeta{
		Sha:      "1234567890abcdef",
		Title:    "feat: add new feature",
		Author:   "John Doe",
		URL:      "https://github.com/repo/commit/1234567890abcdef",
		Branches: []string{"origin/main", "origin/develop"},
	}
	
	patch := "diff --git a/file.go b/file.go\n+func NewFunction() {}"
	
	prompt := builder.CommitChunkPrompt(meta, 1, 2, patch)
	
	// 验证提示词包含必要信息
	assert.Contains(t, prompt, "1234567") // SHA前7位
	assert.Contains(t, prompt, "feat: add new feature")
	assert.Contains(t, prompt, "John Doe")
	assert.Contains(t, prompt, "origin/main, origin/develop")
	assert.Contains(t, prompt, "第 1/2 段")
	assert.Contains(t, prompt, patch)
	assert.Contains(t, prompt, "变更要点")
	assert.Contains(t, prompt, "影响范围")
	assert.Contains(t, prompt, "风险&回滚点")
	assert.Contains(t, prompt, "测试建议")
}

func TestDefaultBuilder_CommitMergePrompt(t *testing.T) {
	builder := NewBuilder()
	
	meta := git.CommitMeta{
		Sha:      "1234567890abcdef",
		Title:    "feat: add new feature",
		Author:   "John Doe",
		URL:      "https://github.com/repo/commit/1234567890abcdef",
		Branches: []string{"origin/main"},
	}
	
	parts := []string{
		"片段1的摘要内容",
		"片段2的摘要内容",
	}
	
	prompt := builder.CommitMergePrompt(meta, parts)
	
	// 验证提示词包含必要信息
	assert.Contains(t, prompt, "1234567") // SHA前7位
	assert.Contains(t, prompt, "【片段1】")
	assert.Contains(t, prompt, "【片段2】")
	assert.Contains(t, prompt, "片段1的摘要内容")
	assert.Contains(t, prompt, "片段2的摘要内容")
	assert.Contains(t, prompt, "变更概述")
	assert.Contains(t, prompt, "影响范围")
	assert.Contains(t, prompt, "风险与回滚点")
	assert.Contains(t, prompt, "测试建议")
}

func TestDefaultBuilder_DailyMergePrompt(t *testing.T) {
	builder := NewBuilder()
	
	items := []CommitSummary{
		{
			Meta: git.CommitMeta{
				Sha:      "1234567890abcdef",
				Title:    "feat: add feature A",
				Author:   "John Doe",
				Branches: []string{"origin/main"},
			},
			Summary: "添加了功能A的实现",
		},
		{
			Meta: git.CommitMeta{
				Sha:      "abcdef1234567890",
				Title:    "fix: fix bug B",
				Author:   "Jane Smith",
				Branches: []string{"origin/develop"},
			},
			Summary: "修复了bug B的问题",
		},
	}
	
	dateLabel := "2024-01-15"
	repo := "test-repo"
	
	prompt := builder.DailyMergePrompt(dateLabel, items, repo)
	
	// 验证提示词包含必要信息
	assert.Contains(t, prompt, "2024-01-15 开发变更日报（test-repo)")
	assert.Contains(t, prompt, "[1234567] feat: add feature A — John Doe — origin/main")
	assert.Contains(t, prompt, "[abcdef1] fix: fix bug B — Jane Smith — origin/develop")
	assert.Contains(t, prompt, "添加了功能A的实现")
	assert.Contains(t, prompt, "修复了bug B的问题")
	assert.Contains(t, prompt, "今日概览")
	assert.Contains(t, prompt, "按分支")
	assert.Contains(t, prompt, "跨分支风险")
	assert.Contains(t, prompt, "建议测试")
}

func TestDefaultBuilder_EmptyInputs(t *testing.T) {
	builder := NewBuilder()
	
	// 测试空patch
	meta := git.CommitMeta{
		Sha:      "1234567890abcdef",
		Title:    "test commit",
		Author:   "Test Author",
		URL:      "https://github.com/test",
		Branches: []string{},
	}
	
	prompt := builder.CommitChunkPrompt(meta, 1, 1, "")
	assert.NotEmpty(t, prompt)
	assert.Contains(t, prompt, "1234567")
	
	// 测试空parts
	prompt = builder.CommitMergePrompt(meta, []string{})
	assert.NotEmpty(t, prompt)
	assert.Contains(t, prompt, "1234567")
	
	// 测试空items
	prompt = builder.DailyMergePrompt("2024-01-15", []CommitSummary{}, "test-repo")
	assert.NotEmpty(t, prompt)
	assert.Contains(t, prompt, "2024-01-15")
	assert.Contains(t, prompt, "test-repo")
}

func TestDefaultBuilder_LongContent(t *testing.T) {
	builder := NewBuilder()
	
	meta := git.CommitMeta{
		Sha:      "1234567890abcdef",
		Title:    strings.Repeat("very long title ", 100),
		Author:   "Test Author",
		URL:      "https://github.com/test",
		Branches: []string{"origin/main"},
	}
	
	longPatch := strings.Repeat("line of diff content\n", 1000)
	
	prompt := builder.CommitChunkPrompt(meta, 1, 1, longPatch)
	assert.NotEmpty(t, prompt)
	assert.Contains(t, prompt, "1234567")
	assert.Contains(t, prompt, longPatch)
}