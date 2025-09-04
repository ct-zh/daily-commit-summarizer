# GitLab 仓库信息初始化提示词

## 任务描述
你需要初始化一个包含所有 GitLab 仓库信息的 JSON 文件。流程包括：扫描用户提供的某个目录、查找 Git 仓库、查询 GitLab API 获取仓库详细信息，并生成结构化的 JSON 配置文件。

## 完整工作流程

### 步骤 1：扫描本地 Git 仓库
1. 在用户提供的目录下搜索所有 Git 仓库
2. 限制目录层级不超过四级（使用 `find` 命令的 `-maxdepth 4` 参数）
3. 提取仓库名称，移除 `.git` 后缀和目录前缀

### 步骤 2：查询 GitLab API 获取仓库信息
1. 使用环境变量 `GITLAB_URL` 和 `GITLAB_TOKEN` 进行 API 认证
2. 对每个本地仓库，在 GitLab 中搜索对应的项目
3. 获取每个仓库的完整信息：
   - `id`: 项目 ID
   - `name`: 仓库名称
   - `url`: 仓库 Web URL
   - `path_with_namespace`: 完整命名空间路径

### 步骤 3：生成 JSON 文件
1. 创建目录结构：`ai/gitlab/`
2. 生成 JSON 文件，格式为数组包含多个仓库对象
3. 每个仓库对象包含四个字段：`id`、`name`、`url`、`path_with_namespace`
4. 如果仓库在 GitLab 中不存在，对应字段设置为 `null`

## 输出文件格式
```json
[
  {
    "id": 21148,
    "name": "yuban.user.buz",
    "url": "https://code.inke.cn/changsha/yuban/server/user/yuban.user.buz",
    "path_with_namespace": "changsha/yuban/server/user/yuban.user.buz"
  },
  {
    "id": 21210,
    "name": "yuban.room.callbuz",
    "url": "https://code.inke.cn/changsha/yuban/server/room/yuban.room.callbuz",
    "path_with_namespace": "changsha/yuban/server/room/yuban.room.callbuz"
  }
]
```

## 关键命令参考
```bash
# 查找所有 Git 仓库
find server -type d -name ".git" -maxdepth 4

# 提取仓库名称
find server -type d -name ".git" -maxdepth 4 | sed 's|/\.git$||' | sed 's|^server/||'

# 查询 GitLab API
curl -H "Private-Token: $GITLAB_TOKEN" "$GITLAB_URL/projects?search={仓库名}"

# 创建目录
mkdir -p ai/gitlab
```

## 注意事项
- 确保环境变量 `GITLAB_URL` 和 `GITLAB_TOKEN` 已正确设置
- 如果某个仓库在 GitLab 中不存在，保留仓库名称但将其他字段设为 null
- JSON 文件必须使用数组格式，即使只有一个仓库
- 输出文件路径：`ai/gitlab/repository.json`（可根据需要修改文件名）