package diff

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"daily-commit-summarizer/internal/config"
	"daily-commit-summarizer/internal/utils"
)

// MockShellExecutor 模拟Shell执行器
type MockShellExecutor struct {
	mock.Mock
}

func (m *MockShellExecutor) Execute(cmd string) (string, error) {
	args := m.Called(cmd)
	return args.String(0), args.Error(1)
}

// MockConfigProvider 模拟配置提供者
type MockConfigProvider struct {
	mock.Mock
}

func (m *MockConfigProvider) GetConfig() *config.Config {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*config.Config)
}

func (m *MockConfigProvider) IsLoaded() bool {
	args := m.Called()
	return args.Bool(0)
}

func TestNewProcessor(t *testing.T) {
	tests := []struct {
		name     string
		config   config.ConfigProvider
		shell    utils.ShellExecutor
		options  *ProcessOptions
		wantType string
	}{
		{
			name:     "创建处理器成功",
			config:   &MockConfigProvider{},
			shell:    &MockShellExecutor{},
			options:  DefaultProcessOptions(),
			wantType: "*diff.DefaultProcessor",
		},
		{
			name:     "使用默认选项",
			config:   &MockConfigProvider{},
			shell:    &MockShellExecutor{},
			options:  nil,
			wantType: "*diff.DefaultProcessor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置mock期望
			mockConfig := tt.config.(*MockConfigProvider)
			mockConfig.On("GetConfig").Return(&config.Config{
				Git: config.GitConfig{
					RepoPath: "/test/repo",
				},
			})
			
			processor := NewProcessor(tt.config, tt.shell, tt.options)
			assert.NotNil(t, processor)
			assert.IsType(t, &DefaultProcessor{}, processor)
			
			// 验证mock调用
			mockConfig.AssertExpectations(t)
		})
	}
}

func TestDefaultProcessor_GetParentSha(t *testing.T) {
	tests := []struct {
		name       string
		sha        string
		mockOutput string
		mockError  error
		wantResult string
		wantError  bool
	}{
		{
			name:       "获取父提交成功",
			sha:        "abc123",
			mockOutput: "abc123 def456",
			mockError:  nil,
			wantResult: "def456",
			wantError:  false,
		},
		{
			name:       "root提交无父提交",
			sha:        "abc123",
			mockOutput: "abc123",
			mockError:  nil,
			wantResult: "",
			wantError:  false,
		},
		{
			name:       "空SHA",
			sha:        "",
			mockOutput: "",
			mockError:  nil,
			wantResult: "",
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockShell := &MockShellExecutor{}
		mockConfig := &MockConfigProvider{}
		
		// 设置config mock期望
		mockConfig.On("GetConfig").Return(&config.Config{
			Git: config.GitConfig{
				RepoPath: "/test/repo",
			},
		})
		
		if tt.sha != "" {
			mockShell.On("Execute", mock.AnythingOfType("string")).Return(tt.mockOutput, tt.mockError)
		}
		
		processor := NewProcessor(mockConfig, mockShell, DefaultProcessOptions())
			
			result, err := processor.GetParentSha(context.Background(), tt.sha)
			
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}
			
			mockShell.AssertExpectations(t)
			mockConfig.AssertExpectations(t)
		})
	}
}

func TestDefaultProcessor_SplitPatchByFile(t *testing.T) {
	tests := []struct {
		name       string
		patch      string
		wantResult []string
	}{
		{
		name:       "空patch",
		patch:      "",
		wantResult: []string{},
		},
		{
			name: "单文件diff",
			patch: `diff --git a/file1.go b/file1.go
index 1234567..abcdefg 100644
--- a/file1.go
+++ b/file1.go
@@ -1,3 +1,3 @@
-old line
+new line`,
			wantResult: []string{"index 1234567..abcdefg 100644\n--- a/file1.go\n+++ b/file1.go\n@@ -1,3 +1,3 @@\n-old line\n+new line"},
		},
		{
			name: "多文件diff",
			patch: `diff --git a/file1.go b/file1.go
index 1234567..abcdefg 100644
--- a/file1.go
+++ b/file1.go
@@ -1,3 +1,3 @@
-old line
+new line
diff --git a/file2.go b/file2.go
index 7890123..defghij 100644
--- a/file2.go
+++ b/file2.go
@@ -1,3 +1,3 @@
-another old line
+another new line`,
			wantResult: []string{
				"index 1234567..abcdefg 100644\n--- a/file1.go\n+++ b/file1.go\n@@ -1,3 +1,3 @@\n-old line\n+new line",
				"index 7890123..defghij 100644\n--- a/file2.go\n+++ b/file2.go\n@@ -1,3 +1,3 @@\n-another old line\n+another new line",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockShell := &MockShellExecutor{}
		mockConfig := &MockConfigProvider{}
		
		// 设置config mock期望
		mockConfig.On("GetConfig").Return(&config.Config{
			Git: config.GitConfig{
				RepoPath: "/test/repo",
			},
		})
		
		processor := NewProcessor(mockConfig, mockShell, DefaultProcessOptions())
			
			result := processor.SplitPatchByFile(tt.patch)
			
			assert.Equal(t, len(tt.wantResult), len(result))
			for i, expected := range tt.wantResult {
				if i < len(result) {
					assert.Equal(t, expected, result[i])
				}
			}
			
			mockConfig.AssertExpectations(t)
		})
	}
}

func TestDefaultProcessor_ChunkBySize(t *testing.T) {
	tests := []struct {
		name       string
		parts      []string
		limit      int
		wantResult []string
	}{
		{
			name:       "空parts",
			parts:      []string{},
			limit:      100,
			wantResult: []string{},
		},
		{
			name:       "单个小片段",
			parts:      []string{"small content"},
			limit:      100,
			wantResult: []string{"small content"},
		},
		{
			name:       "多个小片段合并",
			parts:      []string{"part1", "part2", "part3"},
			limit:      100,
			wantResult: []string{"part1\n\npart2\n\npart3"},
		},
		{
			name:       "超过限制的片段拆分",
			parts:      []string{strings.Repeat("a", 150)},
			limit:      100,
			wantResult: []string{strings.Repeat("a", 100), strings.Repeat("a", 50)},
		},
		{
			name:  "混合大小片段",
			parts: []string{"small1", strings.Repeat("b", 150), "small2"},
			limit: 100,
			wantResult: []string{
				"small1",
				strings.Repeat("b", 100),
				strings.Repeat("b", 50),
				"small2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockShell := &MockShellExecutor{}
		mockConfig := &MockConfigProvider{}
		
		// 设置config mock期望
		mockConfig.On("GetConfig").Return(&config.Config{
			Git: config.GitConfig{
				RepoPath: "/test/repo",
			},
		})
		
		processor := NewProcessor(mockConfig, mockShell, DefaultProcessOptions())
			
			result := processor.ChunkBySize(tt.parts, tt.limit)
			
			assert.Equal(t, len(tt.wantResult), len(result))
			for i, expected := range tt.wantResult {
				if i < len(result) {
					assert.Equal(t, expected, result[i])
				}
			}
			
			// 验证mock期望
			mockConfig.AssertExpectations(t)
		})
	}
}

func TestDefaultProcessor_ProcessDiff(t *testing.T) {
	tests := []struct {
		name           string
		sha            string
		parentOutput   string
		diffOutput     string
		mockError      error
		wantHasChanges bool
		wantIsEmpty    bool
		wantError      bool
	}{
		{
			name:           "正常提交有变更",
			sha:            "abc123",
			parentOutput:   "abc123 def456",
			diffOutput:     "diff --git a/file.go b/file.go\n+new line",
			mockError:      nil,
			wantHasChanges: true,
			wantIsEmpty:    false,
			wantError:      false,
		},
		{
			name:           "空提交无变更",
			sha:            "abc123",
			parentOutput:   "abc123 def456",
			diffOutput:     "",
			mockError:      nil,
			wantHasChanges: false,
			wantIsEmpty:    true,
			wantError:      false,
		},
		{
			name:           "空SHA错误",
			sha:            "",
			parentOutput:   "",
			diffOutput:     "",
			mockError:      nil,
			wantHasChanges: false,
			wantIsEmpty:    false,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockShell := &MockShellExecutor{}
		mockConfig := &MockConfigProvider{}
		
		// 设置config mock期望
		mockConfig.On("GetConfig").Return(&config.Config{
			Git: config.GitConfig{
				RepoPath: "/test/repo",
			},
		})
		
		if tt.sha != "" {
			// 模拟获取父提交的调用
			mockShell.On("Execute", mock.MatchedBy(func(cmd string) bool {
				return strings.Contains(cmd, "git rev-list --parents")
			})).Return(tt.parentOutput, tt.mockError)
			
			// 模拟获取diff的调用
			mockShell.On("Execute", mock.MatchedBy(func(cmd string) bool {
				return strings.Contains(cmd, "git diff")
			})).Return(tt.diffOutput, tt.mockError)
		}
		
		processor := NewProcessor(mockConfig, mockShell, DefaultProcessOptions())
			
			result, err := processor.(*DefaultProcessor).ProcessDiff(context.Background(), tt.sha)
			
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.sha, result.CommitSHA)
				assert.Equal(t, tt.wantHasChanges, result.HasChanges)
				assert.Equal(t, tt.wantIsEmpty, result.IsEmpty)
			}
			
			mockShell.AssertExpectations(t)
		})
	}
}

func TestProcessOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		options *ProcessOptions
		want    *ProcessOptions
	}{
		{
			name: "有效选项",
			options: &ProcessOptions{
				MaxChunkSize:    50000,
				ExcludePatterns: []string{":!*.log"},
			},
			want: &ProcessOptions{
				MaxChunkSize:    50000,
				ExcludePatterns: []string{":!*.log"},
			},
		},
		{
			name: "无效MaxChunkSize自动修正",
			options: &ProcessOptions{
				MaxChunkSize:    0,
				ExcludePatterns: nil,
			},
			want: &ProcessOptions{
				MaxChunkSize:    80000,
				ExcludePatterns: DefaultExcludePatterns,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.Validate()
			assert.NoError(t, err)
			assert.Equal(t, tt.want.MaxChunkSize, tt.options.MaxChunkSize)
			assert.Equal(t, tt.want.ExcludePatterns, tt.options.ExcludePatterns)
		})
	}
}