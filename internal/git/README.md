# Git 模块

Git 模块提供了与 Git 仓库交互的功能，用于收集提交信息和分支数据。该模块是 daily-commit-summarizer 项目的核心组件之一。

## 功能概述

- **远程分支获取**：自动获取并列出所有远程分支
- **提交收集**：按时间范围收集提交记录
- **元数据提取**：提取提交的详细信息（SHA、标题、作者、分支等）
- **分支映射**：建立提交与分支的关联关系
- **数据去重**：自动去除重复的提交记录

## 模块结构

```
internal/git/
├── interfaces.go      # 接口定义
├── models.go          # 数据模型
├── collector.go       # 核心实现
├── collector_test.go  # 测试文件
└── README.md          # 文档说明
```

## 核心接口

### Collector 接口

```go
type Collector interface {
    // FetchRemoteBranches 获取所有远程分支
    FetchRemoteBranches() ([]string, error)
    
    // CollectCommits 收集指定时间范围内的提交
    CollectCommits(since, until string) ([]CommitMeta, error)
}
```

## 数据模型

### CommitMeta 结构体

```go
type CommitMeta struct {
    Sha      string   `json:"sha"`      // 提交的SHA哈希值
    Title    string   `json:"title"`    // 提交标题
    Author   string   `json:"author"`   // 提交作者
    URL      string   `json:"url"`      // 提交链接
    Branches []string `json:"branches"` // 所属分支列表
}
```

## 使用示例

### 基本用法

```go
package main

import (
    "fmt"
    "log"
    
    "daily-commit-summarizer/internal/config"
    "daily-commit-summarizer/internal/git"
    "daily-commit-summarizer/internal/utils"
)

func main() {
    // 创建配置
    cfg := &config.Config{
        Git: config.GitConfig{
            MaxCommits: 200,
        },
    }
    
    // 创建Shell执行器
    shell := utils.NewShellExecutor()
    
    // 创建Git收集器
    collector := git.NewCollector(cfg, shell)
    
    // 获取远程分支
    branches, err := collector.FetchRemoteBranches()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("找到 %d 个远程分支\n", len(branches))
    
    // 收集今日提交
    commits, err := collector.CollectCommits("midnight", "now")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("收集到 %d 个提交\n", len(commits))
    for _, commit := range commits {
        fmt.Printf("- %s: %s (by %s)\n", 
            commit.Sha[:8], commit.Title, commit.Author)
    }
}
```

### 自定义时间范围

```go
// 收集昨天的提交
commits, err := collector.CollectCommits("yesterday", "midnight")

// 收集指定日期范围的提交
commits, err := collector.CollectCommits("2023-01-01", "2023-01-02")

// 收集最近一周的提交
commits, err := collector.CollectCommits("1 week ago", "now")
```

## 实现特性

### 1. 错误处理

- **容错机制**：Git fetch 失败时会记录警告但继续执行
- **分支跳过**：单个分支获取失败时跳过该分支，不影响其他分支
- **提交跳过**：单个提交信息获取失败时跳过该提交

### 2. 性能优化

- **分支限制**：通过 `MaxCommits` 配置限制每个分支的提交数量
- **去重处理**：自动去除重复的提交记录
- **批量操作**：使用 Git 命令批量获取信息

### 3. 数据完整性

- **分支映射**：准确记录每个提交所属的分支
- **时间排序**：按提交时间顺序返回结果
- **元数据完整**：提供完整的提交元数据信息

## Git 命令说明

### 获取远程分支

```bash
# 更新远程分支信息
git fetch --all --prune --tags

# 列出所有远程分支（排除 HEAD）
git for-each-ref --format="%(refname:short)" refs/remotes/origin | grep -v "^origin/HEAD$"
```

### 收集提交记录

```bash
# 获取指定分支的提交（按时间排序）
git log origin/main --no-merges --since="midnight" --until="now" --pretty=format:%H --reverse

# 获取所有分支的提交
git log --no-merges --since="midnight" --until="now" --all --pretty=format:%H --reverse
```

### 获取提交信息

```bash
# 获取提交标题
git show -s --format=%s <sha>

# 获取提交作者
git show -s --format=%an <sha>
```

## 配置要求

### 必需配置

```go
type Config struct {
    Git: GitConfig{
        MaxCommits: 200,  // 每个分支最大提交数量
    },
}
```

### 环境要求

- Git 版本 >= 2.0
- 当前目录必须是 Git 仓库
- 需要有远程仓库访问权限

## 测试

### 运行测试

```bash
# 运行所有测试
GO111MODULE=on go test ./internal/git/... -v

# 运行特定测试
GO111MODULE=on go test ./internal/git/... -v -run TestCollectCommits

# 运行性能测试
GO111MODULE=on go test ./internal/git/... -v -bench=.
```

### 测试覆盖率

```bash
# 生成覆盖率报告
GO111MODULE=on go test ./internal/git/... -v -cover

# 生成详细覆盖率报告
GO111MODULE=on go test ./internal/git/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 错误处理

### 常见错误

1. **Git 仓库未初始化**
   ```
   fatal: not a git repository
   ```
   解决方案：确保在 Git 仓库目录中运行

2. **远程仓库不可访问**
   ```
   fatal: unable to access 'https://github.com/...': ...
   ```
   解决方案：检查网络连接和仓库权限

3. **分支不存在**
   ```
   fatal: bad revision 'origin/nonexistent'
   ```
   解决方案：检查分支名称或更新远程分支信息

### 错误恢复

模块实现了多层错误恢复机制：

- Git fetch 失败时继续执行后续操作
- 单个分支失败时跳过该分支
- 单个提交信息获取失败时跳过该提交
- 返回部分成功的结果而不是完全失败

## 扩展指南

### 添加新的 Git 操作

1. 在 `Collector` 接口中添加新方法
2. 在 `DefaultCollector` 中实现该方法
3. 添加相应的测试用例
4. 更新文档说明

### 自定义分支过滤

```go
// 可以在 FetchRemoteBranches 中添加分支过滤逻辑
func (gc *DefaultCollector) FetchRemoteBranches() ([]string, error) {
    // ... 现有逻辑 ...
    
    // 添加分支过滤
    var filteredBranches []string
    for _, branch := range branches {
        if shouldIncludeBranch(branch) {
            filteredBranches = append(filteredBranches, branch)
        }
    }
    
    return filteredBranches, nil
}
```

## 注意事项

1. **性能考虑**：大型仓库中应适当限制 `MaxCommits` 值
2. **内存使用**：收集大量提交时注意内存消耗
3. **网络依赖**：需要稳定的网络连接来访问远程仓库
4. **权限要求**：确保有足够的权限访问所需的分支和提交
5. **时区处理**：时间参数会受到系统时区设置的影响

## 依赖关系

- `daily-commit-summarizer/internal/config`：配置管理
- `daily-commit-summarizer/internal/utils`：工具函数
- `github.com/stretchr/testify`：测试框架
- 系统 Git 命令：核心 Git 操作

## 版本兼容性

- Go 版本：>= 1.22
- Git 版本：>= 2.0
- 测试框架：testify v1.11.1

---

更多信息请参考项目根目录的 `Go语言实现方案.md` 文档。