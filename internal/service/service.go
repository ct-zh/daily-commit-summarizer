package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"daily-commit-summarizer/internal/config"
	"daily-commit-summarizer/internal/diff"
	"daily-commit-summarizer/internal/git"
	"daily-commit-summarizer/internal/llm"
	"daily-commit-summarizer/internal/notification"
	"daily-commit-summarizer/internal/prompt"
)

// Service 应用服务
// 协调各个子模块完成日报生成的完整流程
type Service struct {
	config   *config.Config
	git      git.Collector
	diff     diff.Processor
	llm      llm.Client
	prompt   prompt.Builder
	notifier notification.Notifier
}

// NewService 创建应用服务
func NewService(
	config *config.Config,
	git git.Collector,
	diff diff.Processor,
	llm llm.Client,
	prompt prompt.Builder,
	notifier notification.Notifier,
) *Service {
	return &Service{
		config:   config,
		git:      git,
		diff:     diff,
		llm:      llm,
		prompt:   prompt,
		notifier: notifier,
	}
}

// Run 运行应用服务
// 执行完整的日报生成流程：收集提交 -> 处理diff -> 生成摘要 -> 发送通知
func (s *Service) Run() error {
	// 设置时间范围
	since := "midnight" // 受TZ环境变量影响
	until := "now"

	// 收集提交
	commitMetas, err := s.git.CollectCommits(since, until)
	if err != nil {
		return fmt.Errorf("收集提交失败: %w", err)
	}

	if len(commitMetas) == 0 {
		return nil // 无提交时直接返回
	}

	// 处理每个提交
	perCommitFinal := s.processCommitsInParallel(commitMetas)

	// 生成当地日期标签
	todayLabel := time.Now().Format("2006-01-02")

	// 汇总当日总览
	var items []prompt.CommitSummary
	for _, item := range perCommitFinal {
		items = append(items, item)
	}

	// 使用默认仓库名称
	repo := "repository"
	if s.config.App.Name != "" {
		repo = s.config.App.Name
	}

	// 生成每日报告
	promptText := s.prompt.DailyMergePrompt(todayLabel, items, repo)
	daily, err := s.llm.Chat(promptText)
	if err != nil {
		// 如果汇总失败，则简单拼接
		var sb strings.Builder
		sb.WriteString("（当日汇总失败，以下为逐提交原始小结拼接）\n\n")

		for i, item := range perCommitFinal {
			if i > 0 {
				sb.WriteString("\n\n---\n\n")
			}
			sb.WriteString(fmt.Sprintf("[%s] %s — %s\n%s",
				item.Meta.Sha[:7], item.Meta.Title, strings.Join(item.Meta.Branches, ", "), item.Summary))
		}

		daily = sb.String()
	}

	// 发送通知
	if err := s.notifier.Send(daily); err != nil {
		return fmt.Errorf("发送通知失败: %w", err)
	}

	return nil
}

// processCommitsInParallel 并行处理提交
func (s *Service) processCommitsInParallel(commitMetas []git.CommitMeta) []prompt.CommitSummary {
	results := make([]prompt.CommitSummary, len(commitMetas))
	wg := sync.WaitGroup{}
	// 使用固定的并发数限制
	maxConcurrency := 3
	semaphore := make(chan struct{}, maxConcurrency)
	mutex := sync.Mutex{}

	for i, meta := range commitMetas {
		wg.Add(1)
		semaphore <- struct{}{} // 获取信号量

		go func(i int, meta git.CommitMeta) {
			defer wg.Done()
			defer func() { <-semaphore }() // 释放信号量

			// 处理单个提交
			summary := s.processCommit(meta)

			// 线程安全地更新结果
			mutex.Lock()
			results[i] = summary
			mutex.Unlock()
		}(i, meta)
	}

	wg.Wait()
	return results
}

// processCommit 处理单个提交
func (s *Service) processCommit(meta git.CommitMeta) prompt.CommitSummary {
	ctx := context.Background()

	// 获取diff
	fullPatch, err := s.diff.GetDiff(ctx, meta.Sha)
	if err != nil {
		return prompt.CommitSummary{
			Meta:    meta,
			Summary: fmt.Sprintf("（获取diff失败：%s）", err.Error()),
		}
	}

	if fullPatch == "" || strings.TrimSpace(fullPatch) == "" {
		return prompt.CommitSummary{
			Meta:    meta,
			Summary: "（无有效业务改动或改动已被过滤，例如 lockfile/构建产物/二进制，或空提交）",
		}
	}

	// 分片diff
	fileParts := s.diff.SplitPatchByFile(fullPatch)
	// 使用固定的分片大小限制
	maxChunkSize := 80000
	chunks := s.diff.ChunkBySize(fileParts, maxChunkSize)

	// 为每个片段生成摘要
	partSummaries := s.processChunksInParallel(meta, chunks)

	// 合并为单提交摘要
	promptText := s.prompt.CommitMergePrompt(meta, partSummaries)
	merged, err := s.llm.Chat(promptText)
	if err != nil {
		merged = strings.Join(partSummaries, "\n\n")
	}

	return prompt.CommitSummary{
		Meta:    meta,
		Summary: merged,
	}
}

// processChunksInParallel 并行处理diff片段
func (s *Service) processChunksInParallel(meta git.CommitMeta, chunks []string) []string {
	results := make([]string, len(chunks))
	wg := sync.WaitGroup{}
	// 使用固定的并发数限制
	maxConcurrency := 3
	semaphore := make(chan struct{}, maxConcurrency)
	mutex := sync.Mutex{}

	for i, chunk := range chunks {
		wg.Add(1)
		semaphore <- struct{}{} // 获取信号量

		go func(i int, chunk string) {
			defer wg.Done()
			defer func() { <-semaphore }() // 释放信号量

			// 处理单个片段
			promptText := s.prompt.CommitChunkPrompt(meta, i+1, len(chunks), chunk)
			sum, err := s.llm.Chat(promptText)

			var result string
			if err != nil {
				result = fmt.Sprintf("（片段%d调用失败：%s）", i+1, err.Error())
			} else if sum == "" {
				result = fmt.Sprintf("（片段%d摘要为空）", i+1)
			} else {
				result = sum
			}

			// 线程安全地更新结果
			mutex.Lock()
			results[i] = result
			mutex.Unlock()
		}(i, chunk)
	}

	wg.Wait()
	return results
}