// Package git 提供Git操作的具体实现
package git

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"daily-commit-summarizer/internal/config"
	"daily-commit-summarizer/internal/utils"
)

// DefaultCollector 默认Git收集器实现
// 实现了Collector接口，提供Git仓库信息收集功能
type DefaultCollector struct {
	config *config.Config
	shell  utils.ShellExecutor
}

// NewCollector 创建Git收集器实例
// config: 配置信息，包含分支限制等参数
// shell: Shell命令执行器，用于执行Git命令
func NewCollector(config *config.Config, shell utils.ShellExecutor) Collector {
	return &DefaultCollector{
		config: config,
		shell:  shell,
	}
}

// FetchRemoteBranches 获取所有远程分支
// 执行git fetch获取最新的远程分支信息，然后列出所有origin/*分支
// 排除origin/HEAD，返回分支名称列表
func (gc *DefaultCollector) FetchRemoteBranches() ([]string, error) {
	// 尝试获取所有远程分支
	_, err := gc.shell.Execute("git fetch --all --prune --tags")
	if err != nil {
		log.Printf("获取远程分支时出现警告，继续执行: %v", err)
	}

	// 列出所有origin/*远程分支，排除origin/HEAD
	cmd := `git for-each-ref --format="%(refname:short)" refs/remotes/origin | grep -v "^origin/HEAD$" || true`
	output, err := gc.shell.Execute(cmd)
	if err != nil {
		return nil, fmt.Errorf("获取远程分支列表失败: %w", err)
	}

	var branches []string
	for _, branch := range strings.Split(output, "\n") {
		branch = strings.TrimSpace(branch)
		if branch != "" {
			branches = append(branches, branch)
		}
	}

	return branches, nil
}

// CollectCommits 收集指定时间范围内的提交
// since: 开始时间，支持Git时间格式
// until: 结束时间，支持Git时间格式
// 返回按时间顺序排列的提交元数据列表
func (gc *DefaultCollector) CollectCommits(since, until string) ([]CommitMeta, error) {
	// 获取远程分支
	remoteBranches, err := gc.FetchRemoteBranches()
	if err != nil {
		return nil, err
	}

	// 创建分支到提交的映射
	branchToCommits := make(map[string][]string)
	for _, rb := range remoteBranches {
		cmd := fmt.Sprintf(`git log %s --no-merges --since="%s" --until="%s" --pretty=format:%%H --reverse || true`, rb, since, until)
		output, err := gc.shell.Execute(cmd)
		if err != nil {
			log.Printf("获取分支提交时出现警告，跳过该分支 %s: %v", rb, err)
			continue
		}

		var commits []string
		for _, commit := range strings.Split(output, "\n") {
			commit = strings.TrimSpace(commit)
			if commit != "" {
				commits = append(commits, commit)
			}
		}

		// 限制每个分支的提交数量
		if len(commits) > gc.config.Git.MaxCommits {
			commits = commits[len(commits)-gc.config.Git.MaxCommits:]
		}

		branchToCommits[rb] = commits
	}

	// 创建提交到分支的映射
	shaToBranches := make(map[string]map[string]bool)
	for rb, shas := range branchToCommits {
		for _, sha := range shas {
			if _, exists := shaToBranches[sha]; !exists {
				shaToBranches[sha] = make(map[string]bool)
			}
			shaToBranches[sha][rb] = true
		}
	}

	// 获取所有分支上的提交，按时间排序
	cmd := fmt.Sprintf(`git log --no-merges --since="%s" --until="%s" --all --pretty=format:%%H --reverse || true`, since, until)
	output, err := gc.shell.Execute(cmd)
	if err != nil {
		return nil, fmt.Errorf("获取所有提交失败: %w", err)
	}

	// 过滤并去重提交
	seen := make(map[string]bool)
	var commitShas []string
	for _, sha := range strings.Split(output, "\n") {
		sha = strings.TrimSpace(sha)
		if sha == "" {
			continue
		}

		if seen[sha] {
			continue
		}

		if _, exists := shaToBranches[sha]; !exists {
			continue
		}

		seen[sha] = true
		commitShas = append(commitShas, sha)
	}

	if len(commitShas) == 0 {
		return nil, nil // 没有有效提交
	}

	// 收集提交元数据
	serverURL := "https://github.com"
	var commitMetas []CommitMeta
	for _, sha := range commitShas {
		titleCmd := fmt.Sprintf("git show -s --format=%%s %s", sha)
		title, err := gc.shell.Execute(titleCmd)
		if err != nil {
			log.Printf("获取提交标题失败，跳过该提交 %s: %v", sha, err)
			continue
		}

		authorCmd := fmt.Sprintf("git show -s --format=%%an %s", sha)
		author, err := gc.shell.Execute(authorCmd)
		if err != nil {
			log.Printf("获取提交作者失败，跳过该提交 %s: %v", sha, err)
			continue
		}

		// 构建提交URL，这里需要从环境变量或配置中获取仓库信息
		// 暂时使用简单的URL格式
		url := fmt.Sprintf("%s/commit/%s", serverURL, sha)

		// 收集分支信息
		var branches []string
		for branch := range shaToBranches[sha] {
			branches = append(branches, branch)
		}
		sort.Strings(branches)

		commitMetas = append(commitMetas, CommitMeta{
			Sha:      sha,
			Title:    strings.TrimSpace(title),
			Author:   strings.TrimSpace(author),
			URL:      url,
			Branches: branches,
		})
	}

	return commitMetas, nil
}