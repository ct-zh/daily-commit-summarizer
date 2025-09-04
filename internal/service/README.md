# 应用服务模块 (Service)

应用服务模块是整个日报生成系统的核心协调器，负责整合各个子模块，实现完整的日报生成和发送流程。

## 功能概述

应用服务模块提供以下核心功能：

- **提交收集**：调用Git模块收集指定时间范围内的提交
- **并行处理**：支持多个提交的并行处理，提高处理效率
- **差异分析**：获取每个提交的代码差异并进行分片处理
- **智能摘要**：利用LLM生成提交摘要和日报内容
- **通知发送**：将生成的日报通过飞书等渠道发送

## 接口设计

### Runner 接口

```go
type Runner interface {
    Run() error
}
```

`Runner` 接口定义了应用服务的核心运行方法，实现了统一的服务启动接口。

## 核心结构

### Service 结构体

```go
type Service struct {
    config   *config.Config
    git      git.Collector
    diff     diff.Processor
    llm      llm.Client
    prompt   prompt.Builder
    notifier notification.Notifier
}
```

`Service` 结构体整合了系统的所有核心模块：

- **config**: 配置管理模块
- **git**: Git操作模块
- **diff**: 代码差异处理模块
- **llm**: 大语言模型客户端
- **prompt**: 提示词构建模块
- **notifier**: 通知发送模块

## 主要方法

### NewService

```go
func NewService(
    config *config.Config,
    git git.Collector,
    diff diff.Processor,
    llm llm.Client,
    prompt prompt.Builder,
    notifier notification.Notifier,
) *Service
```

构造函数，创建新的服务实例。

### Run

```go
func (s *Service) Run() error
```

主要的业务流程方法，执行完整的日报生成流程：

1. **收集提交**：获取从昨天午夜到现在的所有提交
2. **并行处理**：对收集到的提交进行并行处理
3. **生成日报**：汇总所有提交摘要，生成最终日报
4. **发送通知**：将日报发送到指定的通知渠道

### processCommitsInParallel

```go
func (s *Service) processCommitsInParallel(commits []git.CommitMeta) []prompt.CommitSummary
```

并行处理多个提交，使用goroutine和channel实现并发处理，提高处理效率。

### processCommit

```go
func (s *Service) processCommit(meta git.CommitMeta) prompt.CommitSummary
```

处理单个提交的核心逻辑：

1. **获取差异**：获取提交的代码差异
2. **分片处理**：将大的差异分割成小块
3. **生成摘要**：为每个分片生成摘要
4. **合并结果**：将多个分片摘要合并成最终的提交摘要

### processChunksInParallel

```go
func (s *Service) processChunksInParallel(meta git.CommitMeta, chunks []string) []string
```

并行处理提交的多个代码分片，为每个分片生成独立的摘要。

## 处理流程

### 1. 提交收集阶段

- 调用Git模块的 `CollectCommits` 方法
- 收集从昨天午夜到当前时间的所有提交
- 如果没有提交，直接返回成功

### 2. 并行处理阶段

- 使用固定数量的goroutine（默认5个）并行处理提交
- 每个提交独立处理，互不影响
- 通过channel收集处理结果

### 3. 提交处理阶段

对于每个提交：

1. **获取代码差异**：调用diff模块获取提交的完整差异
2. **文件分割**：将差异按文件进行分割
3. **大小分片**：将大文件差异分割成小块（默认80KB）
4. **并行摘要**：为每个分片并行生成摘要
5. **结果合并**：将多个分片摘要合并成最终的提交摘要

### 4. 日报生成阶段

- 收集所有提交的摘要结果
- 使用提示词模块构建日报合并提示词
- 调用LLM生成最终的日报内容

### 5. 通知发送阶段

- 将生成的日报内容发送到配置的通知渠道
- 支持飞书等多种通知方式

## 错误处理

应用服务模块实现了完善的错误处理机制：

- **提交收集失败**：返回详细的错误信息
- **差异获取失败**：为该提交生成错误摘要，不影响其他提交
- **LLM调用失败**：返回错误信息，终止处理
- **通知发送失败**：返回详细的错误信息

## 并发控制

- **提交级并发**：最多5个提交同时处理
- **分片级并发**：每个提交的分片也支持并行处理
- **资源保护**：使用channel和goroutine实现安全的并发控制

## 配置要求

应用服务模块需要以下配置：

- **应用配置**：应用名称等基本信息
- **Git配置**：仓库路径和相关设置
- **LLM配置**：API密钥和模型设置
- **通知配置**：Webhook URL等通知设置

## 依赖关系

应用服务模块依赖以下模块：

- `config`: 配置管理
- `git`: Git操作
- `diff`: 代码差异处理
- `llm`: 大语言模型客户端
- `prompt`: 提示词构建
- `notification`: 通知发送

## 测试

运行单元测试：

```bash
go test ./internal/service -v
```

测试覆盖了以下场景：

- 服务创建和初始化
- 无提交情况的处理
- 提交收集失败的处理
- 完整流程的成功执行
- 通知发送失败的处理
- 单个提交处理的各种情况

## 设计特点

1. **模块化设计**：通过依赖注入实现松耦合
2. **并发处理**：支持多级并发，提高处理效率
3. **错误隔离**：单个提交的错误不影响整体流程
4. **可测试性**：通过接口设计支持完整的单元测试
5. **可扩展性**：易于添加新的处理步骤和通知渠道