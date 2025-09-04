// internal/prompt/builder.go
package prompt

import (
	"fmt"
	"strings"
	
	"daily-commit-summarizer/internal/git"
)

// DefaultBuilder 默认提示词构建器实现
type DefaultBuilder struct{}

// NewBuilder 创建提示词构建器
func NewBuilder() Builder {
	return &DefaultBuilder{}
}

// CommitChunkPrompt 为单个diff片段构建提示词
func (pb *DefaultBuilder) CommitChunkPrompt(meta git.CommitMeta, partIdx, total int, patch string) string {
	return fmt.Sprintf(`你是一名资深工程师与发布经理。以下是提交 %s（%s）的 diff 片段（第 %d/%d 段），请用中文输出结构化摘要：

提交信息：
- SHA: %s
- 标题: %s
- 作者: %s
- 分支: %s
- 链接: %s

要求输出：
1) 变更要点（面向工程师与产品）：列出此片段涉及的主要改动与意图
2) 影响范围：模块/接口/关键文件
3) 风险&回滚点
4) 测试建议
注意：仅基于当前片段，不要臆测；不要贴长代码；如果只是格式化/重命名也请明确指出。

=== DIFF PART BEGIN ===
%s
=== DIFF PART END ===`,
		meta.Sha[:7], meta.Title, partIdx, total, meta.Sha, meta.Title, meta.Author,
		strings.Join(meta.Branches, ", "), meta.URL, patch)
}

// CommitMergePrompt 为合并多个diff片段的摘要构建提示词
func (pb *DefaultBuilder) CommitMergePrompt(meta git.CommitMeta, parts []string) string {
	var joinedParts strings.Builder
	for i, p := range parts {
		if i > 0 {
			joinedParts.WriteString("\n\n")
		}
		joinedParts.WriteString(fmt.Sprintf("【片段%d】\n%s", i+1, p))
	}

	return fmt.Sprintf(`下面是提交 %s 的各片段小结，请合并为**单条提交**的最终摘要（中文），输出以下小节：
- 变更概述（不超过5条要点）
- 影响范围（模块/接口/配置）
- 风险与回滚点
- 测试建议
- 面向用户的可见影响（如有）

请避免重复、合并同类项，标注"可能不完整"当某些片段缺失或被截断。

=== 片段小结集合 BEGIN ===
%s
=== 片段小结集合 END ===`,
		meta.Sha[:7], joinedParts.String())
}

// DailyMergePrompt 为生成每日报告构建提示词
func (pb *DefaultBuilder) DailyMergePrompt(dateLabel string, items []CommitSummary, repo string) string {
	var body strings.Builder
	for i, item := range items {
		if i > 0 {
			body.WriteString("\n\n---\n\n")
		}
		body.WriteString(fmt.Sprintf("[%s] %s — %s — %s\n%s",
			item.Meta.Sha[:7], item.Meta.Title, item.Meta.Author,
			strings.Join(item.Meta.Branches, ", "), item.Summary))
	}

	return fmt.Sprintf(`请将以下"当日各提交摘要"整合成**当日开发变更日报（中文）**，输出结构如下：
# %s 开发变更日报（%s)
1. 今日概览（不超过5条）
2. **按分支**的关键改动清单（每条含模块/影响、是否潜在破坏性）
3. 跨分支风险与回滚策略（如同一提交在多个分支、存在 cherry-pick/divergence）
4. 建议测试与验证清单
5. 其他备注（如重构/依赖升级/仅格式化）

=== 当日提交摘要 BEGIN ===
%s
=== 当日提交摘要 END ===`,
		dateLabel, repo, body.String())
}