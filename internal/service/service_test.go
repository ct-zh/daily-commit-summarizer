package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"daily-commit-summarizer/internal/config"
	"daily-commit-summarizer/internal/git"
	"daily-commit-summarizer/internal/prompt"
)

// MockCollector 模拟Git收集器
type MockCollector struct {
	mock.Mock
}

func (m *MockCollector) FetchRemoteBranches() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockCollector) CollectCommits(since, until string) ([]git.CommitMeta, error) {
	args := m.Called(since, until)
	return args.Get(0).([]git.CommitMeta), args.Error(1)
}

// MockProcessor 模拟Diff处理器
type MockProcessor struct {
	mock.Mock
}

func (m *MockProcessor) GetDiff(ctx context.Context, sha string) (string, error) {
	args := m.Called(ctx, sha)
	return args.String(0), args.Error(1)
}

func (m *MockProcessor) SplitPatchByFile(patch string) []string {
	args := m.Called(patch)
	return args.Get(0).([]string)
}

func (m *MockProcessor) ChunkBySize(parts []string, limit int) []string {
	args := m.Called(parts, limit)
	return args.Get(0).([]string)
}

func (m *MockProcessor) GetParentSha(ctx context.Context, sha string) (string, error) {
	args := m.Called(ctx, sha)
	return args.String(0), args.Error(1)
}

// MockLLMClient 模拟LLM客户端
type MockLLMClient struct {
	mock.Mock
}

func (m *MockLLMClient) Chat(prompt string) (string, error) {
	args := m.Called(prompt)
	return args.String(0), args.Error(1)
}

func (m *MockLLMClient) ChatWithContext(ctx context.Context, prompt string) (string, error) {
	args := m.Called(ctx, prompt)
	return args.String(0), args.Error(1)
}

// MockPromptBuilder 模拟提示词构建器
type MockPromptBuilder struct {
	mock.Mock
}

func (m *MockPromptBuilder) CommitChunkPrompt(meta git.CommitMeta, partIdx, total int, patch string) string {
	args := m.Called(meta, partIdx, total, patch)
	return args.String(0)
}

func (m *MockPromptBuilder) CommitMergePrompt(meta git.CommitMeta, parts []string) string {
	args := m.Called(meta, parts)
	return args.String(0)
}

func (m *MockPromptBuilder) DailyMergePrompt(dateLabel string, items []prompt.CommitSummary, repo string) string {
	args := m.Called(dateLabel, items, repo)
	return args.String(0)
}

// MockNotifier 模拟通知器
type MockNotifier struct {
	mock.Mock
}

func (m *MockNotifier) Send(message string) error {
	args := m.Called(message)
	return args.Error(0)
}

func TestNewService(t *testing.T) {
	cfg := &config.Config{}
	mockGit := &MockCollector{}
	mockDiff := &MockProcessor{}
	mockLLM := &MockLLMClient{}
	mockPrompt := &MockPromptBuilder{}
	mockNotifier := &MockNotifier{}

	service := NewService(cfg, mockGit, mockDiff, mockLLM, mockPrompt, mockNotifier)

	assert.NotNil(t, service)
	assert.Equal(t, cfg, service.config)
	assert.Equal(t, mockGit, service.git)
	assert.Equal(t, mockDiff, service.diff)
	assert.Equal(t, mockLLM, service.llm)
	assert.Equal(t, mockPrompt, service.prompt)
	assert.Equal(t, mockNotifier, service.notifier)
}

func TestService_Run_NoCommits(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{Name: "test-repo"},
	}
	mockGit := &MockCollector{}
	mockDiff := &MockProcessor{}
	mockLLM := &MockLLMClient{}
	mockPrompt := &MockPromptBuilder{}
	mockNotifier := &MockNotifier{}

	service := NewService(cfg, mockGit, mockDiff, mockLLM, mockPrompt, mockNotifier)

	// 模拟无提交的情况
	mockGit.On("CollectCommits", "midnight", "now").Return([]git.CommitMeta{}, nil)

	err := service.Run()

	assert.NoError(t, err)
	mockGit.AssertExpectations(t)
}

func TestService_Run_CollectCommitsError(t *testing.T) {
	cfg := &config.Config{}
	mockGit := &MockCollector{}
	mockDiff := &MockProcessor{}
	mockLLM := &MockLLMClient{}
	mockPrompt := &MockPromptBuilder{}
	mockNotifier := &MockNotifier{}

	service := NewService(cfg, mockGit, mockDiff, mockLLM, mockPrompt, mockNotifier)

	// 模拟收集提交失败
	mockGit.On("CollectCommits", "midnight", "now").Return([]git.CommitMeta{}, errors.New("git error"))

	err := service.Run()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "收集提交失败")
	mockGit.AssertExpectations(t)
}

func TestService_Run_Success(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{Name: "test-repo"},
	}
	mockGit := &MockCollector{}
	mockDiff := &MockProcessor{}
	mockLLM := &MockLLMClient{}
	mockPrompt := &MockPromptBuilder{}
	mockNotifier := &MockNotifier{}

	service := NewService(cfg, mockGit, mockDiff, mockLLM, mockPrompt, mockNotifier)

	// 模拟提交数据
	commits := []git.CommitMeta{
		{
			Sha:      "abc123def456",
			Title:    "Test commit",
			Author:   "Test Author",
			URL:      "https://github.com/test/repo/commit/abc123def456",
			Branches: []string{"origin/main"},
		},
	}

	// 设置模拟期望
	mockGit.On("CollectCommits", "midnight", "now").Return(commits, nil)
	mockDiff.On("GetDiff", mock.Anything, "abc123def456").Return("diff content", nil)
	mockDiff.On("SplitPatchByFile", "diff content").Return([]string{"file1 diff"})
	mockDiff.On("ChunkBySize", []string{"file1 diff"}, 80000).Return([]string{"chunk1"})
	mockPrompt.On("CommitChunkPrompt", commits[0], 1, 1, "chunk1").Return("chunk prompt")
	mockLLM.On("Chat", "chunk prompt").Return("chunk summary", nil)
	mockPrompt.On("CommitMergePrompt", commits[0], []string{"chunk summary"}).Return("merge prompt")
	mockLLM.On("Chat", "merge prompt").Return("commit summary", nil)
	mockPrompt.On("DailyMergePrompt", mock.AnythingOfType("string"), mock.AnythingOfType("[]prompt.CommitSummary"), "test-repo").Return("daily prompt")
	mockLLM.On("Chat", "daily prompt").Return("daily summary", nil)
	mockNotifier.On("Send", "daily summary").Return(nil)

	err := service.Run()

	assert.NoError(t, err)
	mockGit.AssertExpectations(t)
	mockDiff.AssertExpectations(t)
	mockLLM.AssertExpectations(t)
	mockPrompt.AssertExpectations(t)
	mockNotifier.AssertExpectations(t)
}

func TestService_Run_NotificationError(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{Name: "test-repo"},
	}
	mockGit := &MockCollector{}
	mockDiff := &MockProcessor{}
	mockLLM := &MockLLMClient{}
	mockPrompt := &MockPromptBuilder{}
	mockNotifier := &MockNotifier{}

	service := NewService(cfg, mockGit, mockDiff, mockLLM, mockPrompt, mockNotifier)

	// 模拟提交数据
	commits := []git.CommitMeta{
		{
			Sha:      "abc123def456",
			Title:    "Test commit",
			Author:   "Test Author",
			URL:      "https://github.com/test/repo/commit/abc123def456",
			Branches: []string{"origin/main"},
		},
	}

	// 设置模拟期望
	mockGit.On("CollectCommits", "midnight", "now").Return(commits, nil)
	mockDiff.On("GetDiff", mock.Anything, "abc123def456").Return("diff content", nil)
	mockDiff.On("SplitPatchByFile", "diff content").Return([]string{"file1 diff"})
	mockDiff.On("ChunkBySize", []string{"file1 diff"}, 80000).Return([]string{"chunk1"})
	mockPrompt.On("CommitChunkPrompt", commits[0], 1, 1, "chunk1").Return("chunk prompt")
	mockLLM.On("Chat", "chunk prompt").Return("chunk summary", nil)
	mockPrompt.On("CommitMergePrompt", commits[0], []string{"chunk summary"}).Return("merge prompt")
	mockLLM.On("Chat", "merge prompt").Return("commit summary", nil)
	mockPrompt.On("DailyMergePrompt", mock.AnythingOfType("string"), mock.AnythingOfType("[]prompt.CommitSummary"), "test-repo").Return("daily prompt")
	mockLLM.On("Chat", "daily prompt").Return("daily summary", nil)
	mockNotifier.On("Send", "daily summary").Return(errors.New("notification error"))

	err := service.Run()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "发送通知失败")
	mockGit.AssertExpectations(t)
	mockDiff.AssertExpectations(t)
	mockLLM.AssertExpectations(t)
	mockPrompt.AssertExpectations(t)
	mockNotifier.AssertExpectations(t)
}

func TestService_processCommit_EmptyDiff(t *testing.T) {
	cfg := &config.Config{}
	mockGit := &MockCollector{}
	mockDiff := &MockProcessor{}
	mockLLM := &MockLLMClient{}
	mockPrompt := &MockPromptBuilder{}
	mockNotifier := &MockNotifier{}

	service := NewService(cfg, mockGit, mockDiff, mockLLM, mockPrompt, mockNotifier)

	meta := git.CommitMeta{
		Sha:      "abc123def456",
		Title:    "Test commit",
		Author:   "Test Author",
		URL:      "https://github.com/test/repo/commit/abc123def456",
		Branches: []string{"origin/main"},
	}

	// 模拟空diff
	mockDiff.On("GetDiff", mock.Anything, "abc123def456").Return("", nil)

	result := service.processCommit(meta)

	assert.Equal(t, meta, result.Meta)
	assert.Contains(t, result.Summary, "无有效业务改动")
	mockDiff.AssertExpectations(t)
}

func TestService_processCommit_GetDiffError(t *testing.T) {
	cfg := &config.Config{}
	mockGit := &MockCollector{}
	mockDiff := &MockProcessor{}
	mockLLM := &MockLLMClient{}
	mockPrompt := &MockPromptBuilder{}
	mockNotifier := &MockNotifier{}

	service := NewService(cfg, mockGit, mockDiff, mockLLM, mockPrompt, mockNotifier)

	meta := git.CommitMeta{
		Sha:      "abc123def456",
		Title:    "Test commit",
		Author:   "Test Author",
		URL:      "https://github.com/test/repo/commit/abc123def456",
		Branches: []string{"origin/main"},
	}

	// 模拟获取diff失败
	mockDiff.On("GetDiff", mock.Anything, "abc123def456").Return("", errors.New("diff error"))

	result := service.processCommit(meta)

	assert.Equal(t, meta, result.Meta)
	assert.Contains(t, result.Summary, "获取diff失败")
	mockDiff.AssertExpectations(t)
}