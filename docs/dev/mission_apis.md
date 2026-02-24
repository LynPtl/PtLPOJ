# Project Specification: PtLPOJ Content Management & APIs

## 1. 项目背景与目标 (Project Context & Objectives)

在早期的 PtLPOJ 架构中，题目数据（Markdown 题面、Scaffold 模板、隐藏测试用例）是通过本地文件系统（如 `problems.json` 和目录结构）静态加载的。这种方式虽然轻量，但随着系统投入教学使用，暴露出明显的维护瓶颈。

本阶段（Content APIs）的目标是**将静态的本地配置升级为动态的数据库管理**。我们需要赋能“教师/管理员”角色，使其能够通过 API 上传新题、管理测试数据；同时赋能“学生”角色，让其能够通过 VS Code 插件按特定的分类、标签（Tag）或难度进行有体系的“刷题”。

## 2. 核心角色与权限 (Roles & Permissions)

目前系统中的身份是扁平的。我们需要引入基于角色的访问控制 (RBAC)：
* **Student (学生, 默认)**: 拥有只读的题目视图权限、提交代码权限、查看自己的提交记录和统计信息权限。
* **Teacher (教师/管理员)**: 拥有系统特权，可以执行题目的增删改查 (CRUD)、上传隐藏测试用例文件、管理系统的标签库。

## 3. 核心需求描述 (Core Requirements)

### 3.1 题目分类与标签系统 (Categories & Tags)
* **需求痛点**：题目列表是扁平的，学生无法针对特定知识点（如“动态规划”、“快速排序”）或题单进行专项练习。
* **主要功能**：
    * **多维标签 (Tags)**：每道题目可以绑定多个标签（如 `string`, `math`, `two-pointers`）。
    * **查询过滤**：客户端可按标签、难度、状态（已通过/未通过）组合过滤题目列表。
    * **标签云聚合**：系统能统计每个标签下的题目数量，供前端展示分类导航。

### 3.2 教师端：题目动态管理 (Problem Management)
* **需求痛点**：目前加题需要登录服务器修改本地文件，对非硬核开发者的教师极不友好。
* **主要功能**：
    * **元数据录入**：支持设置 Title, Difficulty, Time Limit, Memory Limit。
    * **富文本入库**：将原先基于文件的 Markdown 描述和 Python Scaffold 直接写入数据库（或对象存储）。
    * **测试用例加密上传**：支持通过 API 上传标准的评测用例文件（如 `tests.txt`），并放置到沙盒能够隔离挂载的安全路径下。

### 3.3 学生端：搜索与发现体验 (Discovery UX)
* **需求痛点**：茫茫题海，找题困难。
* **主要功能**：
    * **VS Code 侧边栏重构**：TreeView 从纯粹的列表变更为支持“按标签折叠”的树状结构（例如展开 "动态规划" 节点，下面才是具体题目）。
    * **命令面板集成**：支持 `Ctrl+Shift+P` 呼出类似 `PtLPOJ: Search Problems` 的模糊搜索框。

## 4. API 接口设计规范 (API Specification)

我们将大幅扩展 `/api` 路由，分为 Public 端和 Admin 端。

### 4.1 教师管理接口 (Admin API - Requires User + Teacher Role)
所有此路径下的接口需经过强鉴权拦截器。

* `POST /api/admin/problems`: 创建新题目草稿（包含元数据、题面、初始模板）。
* `PUT /api/admin/problems/{id}`: 更新题目内容。
* `POST /api/admin/problems/{id}/cases`: 上传或覆写评测用的 `tests.txt` (Multipart / Form-data)。
* `DELETE /api/admin/problems/{id}`: 隐藏或彻底删除题目。

### 4.2 学生发现接口 (Public API - Requires User Role)
这是对原有接口的增强。

* `GET /api/problems`: 重点增强此接口，增加 Query 参数：
    * `?tags=dp,greedy` (交集或并集过滤)
    * `?difficulty=Hard`
    * `?search=两数之和`
    * `?status=UNATTEMPTED`
* `GET /api/tags`: 返回系统活跃的所有标签及统计数据 (e.g., `[{"name": "dp", "count": 12}, ...]`)。

## 5. 存储架构演进 (Storage Evolution)

为了支持上述检索，我们需要对 SQLite 数据库引入新的实体并执行数据迁移 (Migration)：

1. **改造 User 表**：新增 `role VARCHAR` 字段。
2. **新增 Problem 表**：将题目标题、描述、脚手架、难度、资源限制等信息从原本的 `problems.json` 和 `.md` 文件转移为数据库记录。
3. **新增 Tag 表**：记录所有标签。
4. **新增 Problem_Tags 关联表**：多对多映射。

*注：庞大的测试用例（可能长达数 MB 的纯文本）依然建议使用本地文件系统 (Storage Layer) 按 `problem_id` 物理隔离存储，数据库中仅保留路径指针，以防止 SQLite 文件过度膨胀。*

## 6. 任务执行顺序 (Execution Plan)

### Phase 8.1: 数据模型层爆发 (Database Migration)
- [ ] 使用 GORM 编写新的模型结构 (`Problem`, `Tag`) 和关联关系。
- [ ] 编写数据迁移脚本 (Migration Script)，读取已有的 `problems.json` 和目录配置，将其批量导入 SQLite。

### Phase 8.2: 接口改造与教师端支持 (Server Expansion)
- [ ] 编写 RBAC 中间件 `RequireRole("teacher")`。
- [ ] 完成 `admin` 路由组的 CRUD 接口。
- [ ] 改造现有的 `GET /api/problems` 以支持高级条件过滤 (Where 拼接)。

### Phase 8.3: 客户端树视图重塑 (Client TreeView)
- [ ] 重写 `treeProvider.ts`，支持将第一层级渲染为 “标签 (Tags)” 或 “难度 (Difficulty)”，第二层级渲染为 “题目”。
- [ ] 集成 VS Code 的搜索与输入过滤框。

---
*Generated based on user requirements for API & Content Management.*
