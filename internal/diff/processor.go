// Package diff 提供代码差异处理功能的核心实现
package diff

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"daily-commit-summarizer/internal/config"
	"daily-commit-summarizer/internal/utils"
)

// DefaultProcessor 默认Diff处理器实现
type DefaultProcessor struct {
	config   config.ConfigProvider
	shell    utils.ShellExecutor
	options  *ProcessOptions
	repoPath string
}

// NewProcessor 创建Diff处理器
func NewProcessor(config config.ConfigProvider, shell utils.ShellExecutor, options *ProcessOptions) Processor {
	if options == nil {
		options = DefaultProcessOptions()
	}
	
	// 验证选项
	if err := options.Validate(); err != nil {
		// 使用默认选项
		options = DefaultProcessOptions()
	}
	
	repoPath := ""
	if config != nil && config.GetConfig() != nil {
		repoPath = config.GetConfig().Git.RepoPath
	}
	
	return &DefaultProcessor{
		config:   config,
		shell:    shell,
		options:  options,
		repoPath: repoPath,
	}
}

// GetParentSha 获取提交的父提交SHA
func (dp *DefaultProcessor) GetParentSha(ctx context.Context, sha string) (string, error) {
	if sha == "" {
		return "", fmt.Errorf("提交SHA不能为空")
	}
	
	cmd := fmt.Sprintf("git rev-list --parents -n 1 %s || true", sha)
	if dp.repoPath != "" {
		cmd = fmt.Sprintf("cd %s && %s", dp.repoPath, cmd)
	}
	
	output, err := dp.shell.Execute(cmd)
	if err != nil {
		return "", fmt.Errorf("获取父提交SHA失败: %w", err)
	}
	
	parts := strings.Fields(output)
	if len(parts) > 1 {
		return parts[1], nil // 非merge情况parent通常只有一个
	}
	
	return "", nil // root commit无parent
}

// GetDiff 获取提交的代码差异
func (dp *DefaultProcessor) GetDiff(ctx context.Context, sha string) (string, error) {
	if sha == "" {
		return "", fmt.Errorf("提交SHA不能为空")
	}
	
	// 获取父提交
	parent, err := dp.GetParentSha(ctx, sha)
	if err != nil {
		return "", fmt.Errorf("获取父提交失败: %w", err)
	}
	
	var base string
	if parent != "" {
		base = parent
	} else {
		// 对于root commit，使用空树
		baseCmd := "git hash-object -t tree /dev/null"
		if dp.repoPath != "" {
			baseCmd = fmt.Sprintf("cd %s && %s", dp.repoPath, baseCmd)
		}
		
		base, err = dp.shell.Execute(baseCmd)
		if err != nil {
			return "", fmt.Errorf("创建空树对象失败: %w", err)
		}
	}
	
	// 构建排除模式
	excludes := strings.Join(dp.options.ExcludePatterns, " ")
	
	// 构建diff命令
	unifiedFlag := fmt.Sprintf("--unified=%d", dp.options.UnifiedContext)
	minimalFlag := ""
	if dp.options.UseMinimalDiff {
		minimalFlag = "--minimal"
	}
	
	cmd := fmt.Sprintf("git diff %s %s %s %s -- . %s || true", 
		unifiedFlag, minimalFlag, base, sha, excludes)
	
	if dp.repoPath != "" {
		cmd = fmt.Sprintf("cd %s && %s", dp.repoPath, cmd)
	}
	
	diff, err := dp.shell.Execute(cmd)
	if err != nil {
		return "", fmt.Errorf("获取diff失败: %w", err)
	}
	
	return diff, nil
}

// SplitPatchByFile 将整个diff按文件拆分
func (dp *DefaultProcessor) SplitPatchByFile(patch string) []string {
	if patch == "" {
		return []string{}
	}
	
	// 使用正则表达式按文件拆分diff
	// 匹配 "diff --git" 开头的行
	re := regexp.MustCompile(`(?m)^diff --git.*$`)
	parts := re.Split(patch, -1)
	
	var result []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	
	return result
}

// ChunkBySize 将大型diff按大小拆分为多个片段
func (dp *DefaultProcessor) ChunkBySize(parts []string, limit int) []string {
	if limit <= 0 {
		limit = dp.options.MaxChunkSize
	}
	
	var out []string
	buf := ""
	
	for _, p := range parts {
		var candidate string
		if buf != "" {
			candidate = buf + "\n\n" + p
		} else {
			candidate = p
		}
		
		if len(candidate) > limit {
			if buf != "" {
				out = append(out, buf)
			}
			
			if len(p) > limit {
				// 如果单个部分超过限制，按字符数切分
				for i := 0; i < len(p); i += limit {
					end := i + limit
					if end > len(p) {
						end = len(p)
					}
					out = append(out, p[i:end])
				}
				buf = ""
			} else {
				buf = p
			}
		} else {
			buf = candidate
		}
	}
	
	if buf != "" {
		out = append(out, buf)
	}
	
	return out
}

// ProcessDiff 处理提交的完整diff流程
func (dp *DefaultProcessor) ProcessDiff(ctx context.Context, sha string) (*DiffResult, error) {
	if sha == "" {
		return nil, fmt.Errorf("提交SHA不能为空")
	}
	
	// 获取父提交SHA
	parentSHA, err := dp.GetParentSha(ctx, sha)
	if err != nil {
		return nil, fmt.Errorf("获取父提交SHA失败: %w", err)
	}
	
	// 获取完整diff
	fullDiff, err := dp.GetDiff(ctx, sha)
	if err != nil {
		return nil, fmt.Errorf("获取diff失败: %w", err)
	}
	
	// 检查是否有变更
	hasChanges := strings.TrimSpace(fullDiff) != ""
	isEmpty := !hasChanges
	
	result := &DiffResult{
		CommitSHA:  sha,
		ParentSHA:  parentSHA,
		FullDiff:   fullDiff,
		HasChanges: hasChanges,
		IsEmpty:    isEmpty,
	}
	
	// 如果没有变更且不包含空提交，直接返回
	if isEmpty && !dp.options.IncludeEmpty {
		return result, nil
	}
	
	// 按文件拆分
	fileParts := dp.SplitPatchByFile(fullDiff)
	result.FileParts = fileParts
	
	// 按大小拆分
	chunkStrings := dp.ChunkBySize(fileParts, dp.options.MaxChunkSize)
	
	// 创建DiffChunk对象
	chunks := make([]DiffChunk, len(chunkStrings))
	for i, content := range chunkStrings {
		chunks[i] = DiffChunk{
			Content: content,
			Index:   i,
			Total:   len(chunkStrings),
			Size:    len(content),
		}
	}
	result.Chunks = chunks
	
	return result, nil
}

// GetProcessOptions 获取当前处理选项
func (dp *DefaultProcessor) GetProcessOptions() *ProcessOptions {
	return dp.options
}

// SetProcessOptions 设置处理选项
func (dp *DefaultProcessor) SetProcessOptions(options *ProcessOptions) error {
	if options == nil {
		return fmt.Errorf("处理选项不能为空")
	}
	
	if err := options.Validate(); err != nil {
		return fmt.Errorf("处理选项验证失败: %w", err)
	}
	
	dp.options = options
	return nil
}