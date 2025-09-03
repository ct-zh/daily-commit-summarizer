// internal/utils/errors.go
package utils

import (
	"fmt"
	"time"
)

// RetryableFunc 定义可重试的函数类型
type RetryableFunc func() (interface{}, error)

// WithRetry 执行带重试的操作
// fn: 要重试的函数
// attempts: 重试次数
// delay: 重试间隔时间
func WithRetry(fn RetryableFunc, attempts int, delay time.Duration) (interface{}, error) {
	var err error
	var result interface{}
	
	for i := 0; i < attempts; i++ {
		result, err = fn()
		if err == nil {
			return result, nil
		}
		
		// 如果不是最后一次尝试，则等待后重试
		if i < attempts-1 {
			time.Sleep(delay)
		}
	}
	
	return nil, fmt.Errorf("操作失败，已重试 %d 次: %w", attempts, err)
}

// IsRetryableError 判断错误是否可重试
// 可以根据具体需求扩展此函数来判断哪些错误类型应该重试
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	// 这里可以根据错误类型或错误消息来判断是否应该重试
	// 例如：网络错误、超时错误等通常是可重试的
	errorMsg := err.Error()
	retryablePatterns := []string{
		"timeout",
		"connection refused",
		"network",
		"temporary",
	}
	
	for _, pattern := range retryablePatterns {
		if contains(errorMsg, pattern) {
			return true
		}
	}
	
	return false
}

// contains 检查字符串是否包含子字符串（不区分大小写）
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
			len(s) > len(substr) && 
			(hasPrefix(s, substr) || hasSuffix(s, substr) || containsInner(s, substr)))
}

// hasPrefix 检查字符串是否以指定前缀开始（简化版，不区分大小写）
func hasPrefix(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	return toLowerCase(s[:len(prefix)]) == toLowerCase(prefix)
}

// hasSuffix 检查字符串是否以指定后缀结束（简化版，不区分大小写）
func hasSuffix(s, suffix string) bool {
	if len(s) < len(suffix) {
		return false
	}
	return toLowerCase(s[len(s)-len(suffix):]) == toLowerCase(suffix)
}

// containsInner 检查字符串内部是否包含子字符串（简化版，不区分大小写）
func containsInner(s, substr string) bool {
	s = toLowerCase(s)
	substr = toLowerCase(substr)
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// toLowerCase 将字符串转换为小写（简化版）
func toLowerCase(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			result[i] = s[i] + 32
		} else {
			result[i] = s[i]
		}
	}
	return string(result)
}