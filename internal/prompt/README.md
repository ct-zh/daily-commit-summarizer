# 提示词模块 (Prompt Module)

提示词模块负责为不同场景构建结构化的LLM提示词，用于生成高质量的代码变更摘要。

## 功能概述

该模块提供三种类型的提示词构建功能：

1. **单个diff片段摘要** - 为大型提交的单个diff片段生成详细摘要
2. **提交级别摘要合并** - 将多个diff片段摘要合并为单个提交的完整摘要
3. **每日报告生成** - 将当日所有提交摘要整合为开发变更日报

## 接口设计

### Builder 接口

```go
type Builder interface {
    // CommitChunkPrompt 为单个diff片段构建提示词
    CommitChunkPrompt(meta git.CommitMeta, partIdx, total int, patch string) string
    
    // CommitMergePrompt 为合并多个diff片段的摘要构建提示词
    CommitMergePrompt(meta git.CommitMeta, parts []string) string
    
    // DailyMergePrompt 为生成每日报告构建提示词
    DailyMergePrompt(dateLabel string, items []CommitSummary, repo string) string
}
```

### 数据结构

```go
// CommitSummary 存储提交摘要
type CommitSummary struct {
    Meta    git.CommitMeta
    Summary string
}
```

## 使用示例

### 基本使用

```go
package main

import (
    "daily-commit-summarizer/internal/prompt"
    "daily-commit-summarizer/internal/git"
)

func main() {
    // 创建提示词构建器
    builder := prompt.NewBuilder()
    
    // 构建提交元数据
    meta := git.CommitMeta{
        Sha:      "1234567890abcdef",
        Title:    "feat: add new feature",
        Author:   "John Doe",
        URL:      "https://github.com/repo/commit/1234567890abcdef",
        Branches: []string{"origin/main", "origin/develop"},
    }
    
    // 1. 为diff片段构建提示词
    patch := "diff --git a/file.go b/file.go\n+func NewFunction() {}"
    chunkPrompt := builder.CommitChunkPrompt(meta, 1, 2, patch)
    
    // 2. 为提交摘要合并构建提示词
    parts := []string{"片段1摘要", "片段2摘要"}
    mergePrompt := builder.CommitMergePrompt(meta, parts)
    
    // 3. 为每日报告构建提示词
    items := []prompt.CommitSummary{
        {
            Meta:    meta,
            Summary: "功能实现摘要",
        },
    }
    dailyPrompt := builder.DailyMergePrompt("2024-01-15", items, "my-repo")
}
```

## 提示词模板

### 1. Diff片段摘要模板

生成的提示词包含以下结构：
- 提交基本信息（SHA、标题、作者、分支、链接）
- 要求输出的结构化内容：
  - 变更要点（面向工程师与产品）
  - 影响范围（模块/接口/关键文件）
  - 风险&回滚点
  - 测试建议
- diff内容

### 2. 提交摘要合并模板

用于将多个diff片段摘要合并为单个提交的完整摘要：
- 变更概述（不超过5条要点）
- 影响范围（模块/接口/配置）
- 风险与回滚点
- 测试建议
- 面向用户的可见影响（如有）

### 3. 每日报告模板

生成结构化的每日开发变更日报：
- 今日概览（不超过5条）
- 按分支的关键改动清单
- 跨分支风险与回滚策略
- 建议测试与验证清单
- 其他备注（重构/依赖升级/格式化等）

## 设计特点

1. **结构化输出** - 所有提示词都要求LLM生成结构化的摘要内容
2. **中文输出** - 专门针对中文团队设计，要求LLM使用中文回复
3. **工程师友好** - 提示词设计考虑了工程师和产品经理的需求
4. **风险意识** - 特别强调风险识别和回滚策略
5. **测试导向** - 每个层级都包含测试建议

## 扩展性

该模块设计为接口驱动，可以轻松扩展：

1. **多语言支持** - 可以实现不同语言版本的Builder
2. **自定义模板** - 可以为不同团队定制提示词模板
3. **动态配置** - 可以从配置文件加载提示词模板
4. **A/B测试** - 可以实现多个Builder版本进行效果对比

## 测试

运行单元测试：

```bash
GO111MODULE=on go test ./internal/prompt/...
```

测试覆盖了以下场景：
- 正常输入的提示词生成
- 空输入的处理
- 长内容的处理
- 边界条件测试