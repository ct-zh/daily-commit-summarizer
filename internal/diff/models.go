// Package diff 提供代码差异处理功能的数据模型
package diff

// DefaultExcludePatterns 默认排除的文件模式
// 这些文件通常不包含有意义的业务逻辑变更
var DefaultExcludePatterns = []string{
	":!**/*.lock",           // 锁文件
	":!**/dist/**",          // 构建输出目录
	":!**/build/**",         // 构建输出目录
	":!**/.next/**",         // Next.js构建目录
	":!**/.vite/**",         // Vite构建目录
	":!**/out/**",           // 输出目录
	":!**/coverage/**",      // 测试覆盖率目录
	":!package-lock.json",   // npm锁文件
	":!pnpm-lock.yaml",      // pnpm锁文件
	":!yarn.lock",           // yarn锁文件
	":!**/*.min.*",          // 压缩文件
	":!**/node_modules/**",  // 依赖目录
	":!**/.git/**",          // Git目录
	":!**/*.log",            // 日志文件
	":!**/*.tmp",            // 临时文件
	":!**/*.temp",           // 临时文件
	":!**/.DS_Store",        // macOS系统文件
	":!**/Thumbs.db",        // Windows系统文件
}

// DiffChunk 表示一个diff片段
type DiffChunk struct {
	// Content diff内容
	Content string
	// Index 片段索引（从0开始）
	Index int
	// Total 总片段数
	Total int
	// Size 片段大小（字符数）
	Size int
}

// DiffResult 表示diff处理结果
type DiffResult struct {
	// CommitSHA 提交SHA
	CommitSHA string
	// ParentSHA 父提交SHA
	ParentSHA string
	// FullDiff 完整的diff内容
	FullDiff string
	// FileParts 按文件拆分的diff片段
	FileParts []string
	// Chunks 按大小拆分的diff片段
	Chunks []DiffChunk
	// HasChanges 是否有有效变更
	HasChanges bool
	// IsEmpty 是否为空提交
	IsEmpty bool
}

// ProcessOptions diff处理选项
type ProcessOptions struct {
	// MaxChunkSize 单个片段的最大字符数
	MaxChunkSize int
	// ExcludePatterns 排除的文件模式
	ExcludePatterns []string
	// IncludeEmpty 是否包含空提交
	IncludeEmpty bool
	// UseMinimalDiff 是否使用最小化diff
	UseMinimalDiff bool
	// UnifiedContext diff上下文行数（0表示不显示上下文）
	UnifiedContext int
}

// DefaultProcessOptions 返回默认的处理选项
func DefaultProcessOptions() *ProcessOptions {
	return &ProcessOptions{
		MaxChunkSize:    80000, // 80KB
		ExcludePatterns: DefaultExcludePatterns,
		IncludeEmpty:    false,
		UseMinimalDiff:  true,
		UnifiedContext:  0, // 不显示上下文，只显示变更行
	}
}

// Validate 验证处理选项
func (opts *ProcessOptions) Validate() error {
	if opts.MaxChunkSize <= 0 {
		opts.MaxChunkSize = 80000
	}
	if opts.ExcludePatterns == nil {
		opts.ExcludePatterns = DefaultExcludePatterns
	}
	return nil
}