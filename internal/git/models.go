// Package git 提供Git操作相关的数据模型
package git

// CommitMeta 存储提交的元数据
// 包含提交的基本信息和所属分支信息
type CommitMeta struct {
	// Sha 提交的SHA哈希值
	Sha string `json:"sha"`
	
	// Title 提交的标题（第一行提交信息）
	Title string `json:"title"`
	
	// Author 提交作者姓名
	Author string `json:"author"`
	
	// URL 提交在远程仓库的访问链接
	URL string `json:"url"`
	
	// Branches 该提交所属的分支列表
	// 一个提交可能存在于多个分支中
	Branches []string `json:"branches"`
}