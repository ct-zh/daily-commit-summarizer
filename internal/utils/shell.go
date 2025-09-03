// internal/utils/shell.go
package utils

import (
	"os/exec"
	"strings"
)

// ShellExecutor 定义Shell命令执行器接口
type ShellExecutor interface {
	// Execute 执行shell命令并返回输出
	Execute(cmd string) (string, error)
}

// DefaultShellExecutor 默认Shell命令执行器
type DefaultShellExecutor struct{}

// NewShellExecutor 创建Shell命令执行器
func NewShellExecutor() ShellExecutor {
	return &DefaultShellExecutor{}
}

// Execute 执行shell命令并返回输出
func (e *DefaultShellExecutor) Execute(cmd string) (string, error) {
	command := exec.Command("sh", "-c", cmd)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}