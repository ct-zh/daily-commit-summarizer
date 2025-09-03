// internal/utils/errors_test.go
package utils

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestWithRetry_Success(t *testing.T) {
	callCount := 0
	fn := func() (interface{}, error) {
		callCount++
		if callCount < 3 {
			return nil, errors.New("temporary error")
		}
		return "success", nil
	}
	
	result, err := WithRetry(fn, 5, 10*time.Millisecond)
	
	if err != nil {
		t.Errorf("WithRetry() error = %v, want nil", err)
	}
	
	if result != "success" {
		t.Errorf("WithRetry() result = %v, want 'success'", result)
	}
	
	if callCount != 3 {
		t.Errorf("Function called %d times, want 3", callCount)
	}
}

func TestWithRetry_Failure(t *testing.T) {
	callCount := 0
	fn := func() (interface{}, error) {
		callCount++
		return nil, errors.New("persistent error")
	}
	
	result, err := WithRetry(fn, 3, 5*time.Millisecond)
	
	if err == nil {
		t.Error("WithRetry() error = nil, want error")
	}
	
	if result != nil {
		t.Errorf("WithRetry() result = %v, want nil", result)
	}
	
	if callCount != 3 {
		t.Errorf("Function called %d times, want 3", callCount)
	}
	
	// 验证错误消息包含重试信息
	if !strings.Contains(err.Error(), "已重试 3 次") {
		t.Errorf("Error message should contain retry information: %v", err)
	}
}

func TestWithRetry_ImmediateSuccess(t *testing.T) {
	callCount := 0
	fn := func() (interface{}, error) {
		callCount++
		return "immediate success", nil
	}
	
	result, err := WithRetry(fn, 3, 10*time.Millisecond)
	
	if err != nil {
		t.Errorf("WithRetry() error = %v, want nil", err)
	}
	
	if result != "immediate success" {
		t.Errorf("WithRetry() result = %v, want 'immediate success'", result)
	}
	
	if callCount != 1 {
		t.Errorf("Function called %d times, want 1", callCount)
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil错误",
			err:  nil,
			want: false,
		},
		{
			name: "超时错误",
			err:  errors.New("connection timeout"),
			want: true,
		},
		{
			name: "网络错误",
			err:  errors.New("network unreachable"),
			want: true,
		},
		{
			name: "连接拒绝错误",
			err:  errors.New("connection refused"),
			want: true,
		},
		{
			name: "临时错误",
			err:  errors.New("temporary failure"),
			want: true,
		},
		{
			name: "语法错误",
			err:  errors.New("syntax error"),
			want: false,
		},
		{
			name: "权限错误",
			err:  errors.New("permission denied"),
			want: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRetryableError(tt.err); got != tt.want {
				t.Errorf("IsRetryableError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToLowerCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "全大写",
			input: "HELLO",
			want:  "hello",
		},
		{
			name:  "混合大小写",
			input: "Hello World",
			want:  "hello world",
		},
		{
			name:  "全小写",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "空字符串",
			input: "",
			want:  "",
		},
		{
			name:  "包含数字和符号",
			input: "Hello123!@#",
			want:  "hello123!@#",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toLowerCase(tt.input); got != tt.want {
				t.Errorf("toLowerCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		substr string
		want   bool
	}{
		{
			name:   "包含子字符串",
			s:      "Hello World",
			substr: "World",
			want:   true,
		},
		{
			name:   "不包含子字符串",
			s:      "Hello World",
			substr: "xyz",
			want:   false,
		},
		{
			name:   "大小写不敏感",
			s:      "Hello World",
			substr: "WORLD",
			want:   true,
		},
		{
			name:   "空子字符串",
			s:      "Hello",
			substr: "",
			want:   true,
		},
		{
			name:   "相等字符串",
			s:      "Hello",
			substr: "Hello",
			want:   true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contains(tt.s, tt.substr); got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

// BenchmarkWithRetry 性能测试
func BenchmarkWithRetry(b *testing.B) {
	fn := func() (interface{}, error) {
		return "success", nil
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := WithRetry(fn, 3, time.Millisecond)
		if err != nil {
			b.Fatalf("WithRetry() error = %v", err)
		}
	}
}