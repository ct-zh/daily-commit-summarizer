# 工具模块 (Utils)

本模块提供了Daily Commit Summarizer项目中使用的通用工具函数和接口。

## 模块结构

```
internal/utils/
├── shell.go          # Shell命令执行器
├── shell_test.go     # Shell执行器测试
├── errors.go         # 错误处理和重试机制
├── errors_test.go    # 错误处理测试
└── README.md         # 本文档
```

## 功能说明

### Shell命令执行器 (shell.go)

提供了统一的Shell命令执行接口，用于执行Git命令和其他系统命令。

#### 接口定义

```go
type ShellExecutor interface {
    Execute(cmd string) (string, error)
}
```

#### 使用示例

```go
package main

import (
    "fmt"
    "daily-commit-summarizer/internal/utils"
)

func main() {
    executor := utils.NewShellExecutor()
    
    // 执行Git命令
    output, err := executor.Execute("git status")
    if err != nil {
        fmt.Printf("命令执行失败: %v\n", err)
        return
    }
    
    fmt.Printf("Git状态: %s\n", output)
}
```

#### 特性

- **统一接口**: 提供一致的命令执行接口
- **错误处理**: 完整的错误信息返回
- **输出清理**: 自动去除输出中的前后空白字符
- **跨平台**: 使用`sh -c`确保命令在不同系统上的兼容性

### 错误处理和重试机制 (errors.go)

提供了重试机制和错误类型判断功能，用于处理网络请求、API调用等可能失败的操作。

#### 主要函数

##### WithRetry

执行带重试的操作：

```go
func WithRetry(fn RetryableFunc, attempts int, delay time.Duration) (interface{}, error)
```

**参数说明**:
- `fn`: 要重试的函数
- `attempts`: 重试次数
- `delay`: 重试间隔时间

**使用示例**:

```go
package main

import (
    "fmt"
    "time"
    "daily-commit-summarizer/internal/utils"
)

func main() {
    // 定义一个可能失败的操作
    operation := func() (interface{}, error) {
        // 模拟API调用
        return callAPI()
    }
    
    // 执行重试操作
    result, err := utils.WithRetry(operation, 3, 2*time.Second)
    if err != nil {
        fmt.Printf("操作失败: %v\n", err)
        return
    }
    
    fmt.Printf("操作成功: %v\n", result)
}
```

##### IsRetryableError

判断错误是否可重试：

```go
func IsRetryableError(err error) bool
```

**支持的可重试错误类型**:
- 超时错误 (timeout)
- 连接拒绝 (connection refused)
- 网络错误 (network)
- 临时错误 (temporary)

**使用示例**:

```go
if utils.IsRetryableError(err) {
    // 可以重试的错误
    fmt.Println("这是一个可重试的错误")
} else {
    // 不应该重试的错误
    fmt.Println("这是一个不可重试的错误")
}
```

## 设计原则

### 1. 接口驱动

使用接口定义功能，便于测试和模块替换：

```go
// 可以轻松创建模拟对象进行测试
type MockShellExecutor struct{}

func (m *MockShellExecutor) Execute(cmd string) (string, error) {
    return "mocked output", nil
}
```

### 2. 错误处理一致性

所有函数都提供详细的错误信息，包含中文描述：

```go
return nil, fmt.Errorf("操作失败，已重试 %d 次: %w", attempts, err)
```

### 3. 可测试性

每个功能都有对应的测试用例，确保代码质量：

- `shell_test.go`: 测试Shell命令执行
- `errors_test.go`: 测试重试机制和错误处理

### 4. 性能考虑

- 提供基准测试 (Benchmark)
- 避免不必要的内存分配
- 使用高效的字符串操作

## 测试

运行所有测试：

```bash
GO111MODULE=on go test ./internal/utils/... -v
```

运行性能测试：

```bash
GO111MODULE=on go test ./internal/utils/... -bench=.
```

## 依赖关系

本模块只依赖Go标准库，无外部依赖：

- `os/exec`: 用于执行Shell命令
- `strings`: 用于字符串处理
- `time`: 用于重试延迟
- `fmt`: 用于错误格式化

## 扩展指南

### 添加新的Shell执行器

如果需要支持不同的Shell或执行环境：

```go
type CustomShellExecutor struct {
    shell string
}

func (e *CustomShellExecutor) Execute(cmd string) (string, error) {
    command := exec.Command(e.shell, "-c", cmd)
    // 自定义实现
}
```

### 扩展重试策略

可以添加更复杂的重试策略：

```go
type RetryConfig struct {
    MaxAttempts int
    BaseDelay   time.Duration
    MaxDelay    time.Duration
    Multiplier  float64
}

func WithAdvancedRetry(fn RetryableFunc, config RetryConfig) (interface{}, error) {
    // 实现指数退避等高级重试策略
}
```

## 注意事项

1. **Shell命令安全**: 使用Shell执行器时要注意命令注入风险
2. **重试次数**: 合理设置重试次数，避免无限重试
3. **错误分类**: 正确判断哪些错误应该重试，哪些不应该
4. **资源清理**: 确保在重试过程中正确清理资源

## 版本历史

- v1.0.0: 初始版本，包含Shell执行器和重试机制