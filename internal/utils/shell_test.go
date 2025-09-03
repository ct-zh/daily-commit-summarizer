// internal/utils/shell_test.go
package utils

import (
	"strings"
	"testing"
)

func TestDefaultShellExecutor_Execute(t *testing.T) {
	executor := NewShellExecutor()
	
	tests := []struct {
		name        string
		cmd         string
		wantErr     bool
		wantContains string
	}{
		{
			name:        "简单echo命令",
			cmd:         "echo 'hello world'",
			wantErr:     false,
			wantContains: "hello world",
		},
		{
			name:        "pwd命令",
			cmd:         "pwd",
			wantErr:     false,
			wantContains: "/",
		},
		{
			name:        "无效命令",
			cmd:         "nonexistentcommand123",
			wantErr:     true,
			wantContains: "",
		},
		{
			name:        "ls命令",
			cmd:         "ls -la | head -1",
			wantErr:     false,
			wantContains: "total",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.Execute(tt.cmd)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && tt.wantContains != "" {
				if !strings.Contains(result, tt.wantContains) {
					t.Errorf("Execute() result = %v, want contains %v", result, tt.wantContains)
				}
			}
			
			// 验证输出不包含多余的空白字符
			if !tt.wantErr && strings.HasPrefix(result, " ") || strings.HasSuffix(result, " ") {
				t.Errorf("Execute() result has leading/trailing spaces: '%s'", result)
			}
		})
	}
}

func TestNewShellExecutor(t *testing.T) {
	executor := NewShellExecutor()
	if executor == nil {
		t.Error("NewShellExecutor() returned nil")
	}
	
	// 验证返回的是正确的类型
	if _, ok := executor.(*DefaultShellExecutor); !ok {
		t.Error("NewShellExecutor() did not return *DefaultShellExecutor")
	}
}

// BenchmarkShellExecutor_Execute 性能测试
func BenchmarkShellExecutor_Execute(b *testing.B) {
	executor := NewShellExecutor()
	cmd := "echo 'benchmark test'"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := executor.Execute(cmd)
		if err != nil {
			b.Fatalf("Execute() error = %v", err)
		}
	}
}