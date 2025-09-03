package git

import (
	"errors"
	"testing"

	"daily-commit-summarizer/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockShellExecutor 模拟Shell执行器
type MockShellExecutor struct {
	mock.Mock
}

func (m *MockShellExecutor) Execute(cmd string) (string, error) {
	args := m.Called(cmd)
	return args.String(0), args.Error(1)
}

// TestNewCollector 测试创建Git收集器
func TestNewCollector(t *testing.T) {
	config := &config.Config{
		Git: config.GitConfig{
			MaxCommits: 200,
		},
	}
	shell := &MockShellExecutor{}
	
	collector := NewCollector(config, shell)
	
	assert.NotNil(t, collector)
	assert.IsType(t, &DefaultCollector{}, collector)
}

// TestFetchRemoteBranches_Success 测试成功获取远程分支
func TestFetchRemoteBranches_Success(t *testing.T) {
	config := &config.Config{
		Git: config.GitConfig{
			MaxCommits: 200,
		},
	}
	shell := &MockShellExecutor{}
	
	// 模拟git fetch命令
	shell.On("Execute", "git fetch --all --prune --tags").Return("", nil)
	
	// 模拟获取分支列表命令
	branchOutput := "origin/main\norigin/develop\norigin/feature/test"
	shell.On("Execute", `git for-each-ref --format="%(refname:short)" refs/remotes/origin | grep -v "^origin/HEAD$" || true`).Return(branchOutput, nil)
	
	collector := NewCollector(config, shell)
	branches, err := collector.FetchRemoteBranches()
	
	assert.NoError(t, err)
	assert.Len(t, branches, 3)
	assert.Contains(t, branches, "origin/main")
	assert.Contains(t, branches, "origin/develop")
	assert.Contains(t, branches, "origin/feature/test")
	
	shell.AssertExpectations(t)
}

// TestFetchRemoteBranches_FetchError 测试fetch命令出错的情况
func TestFetchRemoteBranches_FetchError(t *testing.T) {
	config := &config.Config{
		Git: config.GitConfig{
			MaxCommits: 200,
		},
	}
	shell := &MockShellExecutor{}
	
	// 模拟git fetch命令失败
	shell.On("Execute", "git fetch --all --prune --tags").Return("", errors.New("fetch failed"))
	
	// 模拟获取分支列表命令成功
	branchOutput := "origin/main"
	shell.On("Execute", `git for-each-ref --format="%(refname:short)" refs/remotes/origin | grep -v "^origin/HEAD$" || true`).Return(branchOutput, nil)
	
	collector := NewCollector(config, shell)
	branches, err := collector.FetchRemoteBranches()
	
	// fetch失败不应该影响获取分支列表
	assert.NoError(t, err)
	assert.Len(t, branches, 1)
	assert.Contains(t, branches, "origin/main")
	
	shell.AssertExpectations(t)
}

// TestFetchRemoteBranches_ListError 测试获取分支列表失败
func TestFetchRemoteBranches_ListError(t *testing.T) {
	config := &config.Config{
		Git: config.GitConfig{
			MaxCommits: 200,
		},
	}
	shell := &MockShellExecutor{}
	
	// 模拟git fetch命令
	shell.On("Execute", "git fetch --all --prune --tags").Return("", nil)
	
	// 模拟获取分支列表命令失败
	shell.On("Execute", `git for-each-ref --format="%(refname:short)" refs/remotes/origin | grep -v "^origin/HEAD$" || true`).Return("", errors.New("list failed"))
	
	collector := NewCollector(config, shell)
	branches, err := collector.FetchRemoteBranches()
	
	assert.Error(t, err)
	assert.Nil(t, branches)
	assert.Contains(t, err.Error(), "获取远程分支列表失败")
	
	shell.AssertExpectations(t)
}

// TestCollectCommits_Success 测试成功收集提交
func TestCollectCommits_Success(t *testing.T) {
	config := &config.Config{
		Git: config.GitConfig{
			MaxCommits: 200,
		},
	}
	shell := &MockShellExecutor{}
	
	// 模拟FetchRemoteBranches的调用
	shell.On("Execute", "git fetch --all --prune --tags").Return("", nil)
	shell.On("Execute", `git for-each-ref --format="%(refname:short)" refs/remotes/origin | grep -v "^origin/HEAD$" || true`).Return("origin/main", nil)
	
	// 模拟获取分支提交
	shell.On("Execute", `git log origin/main --no-merges --since="midnight" --until="now" --pretty=format:%H --reverse || true`).Return("abc123\ndef456", nil)
	
	// 模拟获取所有提交
	shell.On("Execute", `git log --no-merges --since="midnight" --until="now" --all --pretty=format:%H --reverse || true`).Return("abc123\ndef456", nil)
	
	// 模拟获取提交信息
	shell.On("Execute", "git show -s --format=%s abc123").Return("feat: add new feature", nil)
	shell.On("Execute", "git show -s --format=%an abc123").Return("John Doe", nil)
	shell.On("Execute", "git show -s --format=%s def456").Return("fix: bug fix", nil)
	shell.On("Execute", "git show -s --format=%an def456").Return("Jane Smith", nil)
	
	collector := NewCollector(config, shell)
	commits, err := collector.CollectCommits("midnight", "now")
	
	assert.NoError(t, err)
	assert.Len(t, commits, 2)
	
	// 验证第一个提交
	assert.Equal(t, "abc123", commits[0].Sha)
	assert.Equal(t, "feat: add new feature", commits[0].Title)
	assert.Equal(t, "John Doe", commits[0].Author)
	assert.Contains(t, commits[0].URL, "abc123")
	assert.Contains(t, commits[0].Branches, "origin/main")
	
	// 验证第二个提交
	assert.Equal(t, "def456", commits[1].Sha)
	assert.Equal(t, "fix: bug fix", commits[1].Title)
	assert.Equal(t, "Jane Smith", commits[1].Author)
	assert.Contains(t, commits[1].URL, "def456")
	assert.Contains(t, commits[1].Branches, "origin/main")
	
	shell.AssertExpectations(t)
}

// TestCollectCommits_NoCommits 测试没有提交的情况
func TestCollectCommits_NoCommits(t *testing.T) {
	config := &config.Config{
		Git: config.GitConfig{
			MaxCommits: 200,
		},
	}
	shell := &MockShellExecutor{}
	
	// 模拟FetchRemoteBranches的调用
	shell.On("Execute", "git fetch --all --prune --tags").Return("", nil)
	shell.On("Execute", `git for-each-ref --format="%(refname:short)" refs/remotes/origin | grep -v "^origin/HEAD$" || true`).Return("origin/main", nil)
	
	// 模拟没有提交
	shell.On("Execute", `git log origin/main --no-merges --since="midnight" --until="now" --pretty=format:%H --reverse || true`).Return("", nil)
	shell.On("Execute", `git log --no-merges --since="midnight" --until="now" --all --pretty=format:%H --reverse || true`).Return("", nil)
	
	collector := NewCollector(config, shell)
	commits, err := collector.CollectCommits("midnight", "now")
	
	assert.NoError(t, err)
	assert.Nil(t, commits)
	
	shell.AssertExpectations(t)
}

// TestCollectCommits_FetchBranchesError 测试获取分支失败
func TestCollectCommits_FetchBranchesError(t *testing.T) {
	config := &config.Config{
		Git: config.GitConfig{
			MaxCommits: 200,
		},
	}
	shell := &MockShellExecutor{}
	
	// 模拟FetchRemoteBranches失败
	shell.On("Execute", "git fetch --all --prune --tags").Return("", nil)
	shell.On("Execute", `git for-each-ref --format="%(refname:short)" refs/remotes/origin | grep -v "^origin/HEAD$" || true`).Return("", errors.New("fetch branches failed"))
	
	collector := NewCollector(config, shell)
	commits, err := collector.CollectCommits("midnight", "now")
	
	assert.Error(t, err)
	assert.Nil(t, commits)
	
	shell.AssertExpectations(t)
}

// BenchmarkCollectCommits 性能测试
func BenchmarkCollectCommits(b *testing.B) {
	config := &config.Config{
		Git: config.GitConfig{
			MaxCommits: 200,
		},
	}
	shell := &MockShellExecutor{}
	
	// 设置模拟调用
	shell.On("Execute", "git fetch --all --prune --tags").Return("", nil)
	shell.On("Execute", `git for-each-ref --format="%(refname:short)" refs/remotes/origin | grep -v "^origin/HEAD$" || true`).Return("origin/main", nil)
	shell.On("Execute", `git log origin/main --no-merges --since="midnight" --until="now" --pretty=format:%H --reverse || true`).Return("abc123", nil)
	shell.On("Execute", `git log --no-merges --since="midnight" --until="now" --all --pretty=format:%H --reverse || true`).Return("abc123", nil)
	shell.On("Execute", "git show -s --format=%s abc123").Return("test commit", nil)
	shell.On("Execute", "git show -s --format=%an abc123").Return("Test User", nil)
	
	collector := NewCollector(config, shell)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = collector.CollectCommits("midnight", "now")
	}
}