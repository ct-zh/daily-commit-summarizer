# Diff模块

Diff模块负责处理Git提交的代码差异，提供diff获取、文件拆分和大小分片等功能。

## 功能概述

- **获取提交差异**: 获取指定提交相对于其父提交的代码差异
- **文件级拆分**: 将完整的diff按文件进行拆分
- **大小分片**: 将大型diff按字符数限制拆分为多个片段
- **排除文件**: 支持排除特定类型的文件（如锁文件、构建产物等）
- **配置灵活**: 支持自定义处理选项和排除模式

## 核心接口

### Processor

```go
type Processor interface {
    // 获取提交的代码差异
    GetDiff(ctx context.Context, sha string) (string, error)
    
    // 按文件拆分diff
    SplitPatchByFile(patch string) []string
    
    // 按大小拆分diff片段
    ChunkBySize(parts []string, limit int) []string
    
    // 获取父提交SHA
    GetParentSha(ctx context.Context, sha string) (string, error)
}
```

## 使用示例

### 基本使用

```go
package main

import (
    "context"
    "fmt"
    
    "daily-commit-summarizer/internal/config"
    "daily-commit-summarizer/internal/diff"
    "daily-commit-summarizer/internal/utils"
)

func main() {
    // 创建依赖
    shell := utils.NewShellExecutor()
    configProvider := config.NewDefaultConfigLoader("config.yaml")
    
    // 创建处理器
    processor := diff.NewProcessor(configProvider, shell, nil)
    
    // 获取提交差异
    ctx := context.Background()
    diffContent, err := processor.GetDiff(ctx, "abc123")
    if err != nil {
        panic(err)
    }
    
    // 按文件拆分
    fileParts := processor.SplitPatchByFile(diffContent)
    
    // 按大小拆分
    chunks := processor.ChunkBySize(fileParts, 80000)
    
    fmt.Printf("共拆分为 %d 个片段\n", len(chunks))
}
```

### 自定义配置

```go
// 创建自定义处理选项
options := &diff.ProcessOptions{
    MaxChunkSize: 50000,
    ExcludePatterns: []string{
        ":!**/*.lock",
        ":!**/dist/**",
        ":!**/*.min.*",
    },
    IncludeEmpty: false,
    UseMinimalDiff: true,
    UnifiedContext: 0,
}

// 使用自定义选项创建处理器
processor := diff.NewProcessor(configProvider, shell, options)
```

### 完整处理流程

```go
// 使用ProcessDiff方法进行完整处理
result, err := processor.(*diff.DefaultProcessor).ProcessDiff(ctx, "abc123")
if err != nil {
    panic(err)
}

fmt.Printf("提交SHA: %s\n", result.CommitSHA)
fmt.Printf("父提交SHA: %s\n", result.ParentSHA)
fmt.Printf("是否有变更: %v\n", result.HasChanges)
fmt.Printf("是否为空提交: %v\n", result.IsEmpty)
fmt.Printf("文件片段数: %d\n", len(result.FileParts))
fmt.Printf("大小片段数: %d\n", len(result.Chunks))

// 遍历处理每个片段
for i, chunk := range result.Chunks {
    fmt.Printf("片段 %d/%d (大小: %d 字符)\n", 
        chunk.Index+1, chunk.Total, chunk.Size)
    // 处理chunk.Content...
}
```

## 配置选项

### ProcessOptions

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| MaxChunkSize | int | 80000 | 单个片段的最大字符数 |
| ExcludePatterns | []string | DefaultExcludePatterns | 排除的文件模式 |
| IncludeEmpty | bool | false | 是否包含空提交 |
| UseMinimalDiff | bool | true | 是否使用最小化diff |
| UnifiedContext | int | 0 | diff上下文行数 |

### 默认排除模式

```go
var DefaultExcludePatterns = []string{
    ":!**/*.lock",           // 锁文件
    ":!**/dist/**",          // 构建输出目录
    ":!**/build/**",         // 构建输出目录
    ":!**/.next/**",         // Next.js构建目录
    ":!**/.vite/**",         // Vite构建目录
    ":!**/out/**",           // 输出目录
    ":!**/coverage/**",      // 测试覆盖率目录
    ":!package-lock.json",   // npm锁文件
    ":!pnpm-lock.yaml",      // pnpm锁文件
    ":!yarn.lock",           // yarn锁文件
    ":!**/*.min.*",          // 压缩文件
    ":!**/node_modules/**",  // 依赖目录
    ":!**/.git/**",          // Git目录
    ":!**/*.log",            // 日志文件
    ":!**/*.tmp",            // 临时文件
    ":!**/*.temp",           // 临时文件
    ":!**/.DS_Store",        // macOS系统文件
    ":!**/Thumbs.db",        // Windows系统文件
}
```

## 数据结构

### DiffResult

```go
type DiffResult struct {
    CommitSHA   string      // 提交SHA
    ParentSHA   string      // 父提交SHA
    FullDiff    string      // 完整的diff内容
    FileParts   []string    // 按文件拆分的diff片段
    Chunks      []DiffChunk // 按大小拆分的diff片段
    HasChanges  bool        // 是否有有效变更
    IsEmpty     bool        // 是否为空提交
}
```

### DiffChunk

```go
type DiffChunk struct {
    Content string // diff内容
    Index   int    // 片段索引（从0开始）
    Total   int    // 总片段数
    Size    int    // 片段大小（字符数）
}
```

## 错误处理

模块提供详细的错误信息，包括：

- 提交SHA验证错误
- Git命令执行错误
- 父提交获取错误
- diff获取错误
- 配置验证错误

```go
// 错误处理示例
diffContent, err := processor.GetDiff(ctx, sha)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "提交SHA不能为空"):
        // 处理SHA验证错误
    case strings.Contains(err.Error(), "获取diff失败"):
        // 处理diff获取错误
    default:
        // 处理其他错误
    }
}
```

## 测试

运行测试：

```bash
# 运行所有测试
go test ./internal/diff

# 运行测试并显示覆盖率
go test -cover ./internal/diff

# 运行特定测试
go test -run TestDefaultProcessor_GetDiff ./internal/diff
```

## 性能考虑

1. **内存使用**: 大型diff会占用较多内存，建议合理设置MaxChunkSize
2. **Git命令**: 频繁的Git命令调用可能影响性能，考虑批量处理
3. **正则表达式**: 文件拆分使用正则表达式，对于超大diff可能较慢
4. **字符串操作**: 大量字符串拼接和拆分操作，注意内存分配

## 最佳实践

1. **合理设置片段大小**: 根据LLM的token限制设置MaxChunkSize
2. **排除无关文件**: 使用ExcludePatterns排除构建产物和依赖文件
3. **错误处理**: 妥善处理Git命令可能的失败情况
4. **上下文管理**: 使用context.Context进行超时和取消控制
5. **资源清理**: 及时释放大型字符串占用的内存

## 依赖

- `daily-commit-summarizer/internal/config`: 配置管理
- `daily-commit-summarizer/internal/utils`: 工具函数
- `github.com/stretchr/testify`: 测试框架