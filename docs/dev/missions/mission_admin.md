# Project Specification: PtLPOJ 快速管理控制台与 API (Admin Control Panel)

## 1. 项目背景 (Context)
在早期的架构设计中，教师和管理员（系统拥有者）如果需要管理用户白名单，或者向系统中补充新的编程题目，只能通过极其底层的手段：
- 登录服务器终端，使用 `sqlite3` 直插 `users` 数据表。
- 将 `.py` 题目代码通过 SCP 等方式手动上传至服务器。
这对于非技术背景的教务人员来说维护成本极高、体验较差。因此，我们需要一个原生的、基于 VS Code 生态环境的“可视化管理层”来支持这些高级操作权限。

## 2. 核心目标 (Objectives)
构建一套基于 `AdminToken` 鉴权的轻量级后台体系，并通过 VS Code 原生 Webview 及系统的原生文件交互（OpenDialog）为教师提供低学习门槛的可视化增删改查面板。其设计应当满足：
- **安全却极简**：无需引入庞大的 RBAC 角色域表，只需环境变量即可赋权。
- **题目智能解析**：上传 `.py` 文件时，后端必须能通过 AST 自动切分函数模板与 doctest。
- **大批量处理**：管理员经常面临几百人开课与大量题单，需要支持 Excel 级别的复制粘贴体验和批量文件推送。

## 3. 主要功能与架构 (Architecture & Features)

### 3.1 极简 Admin 鉴权体系
- **机制**：由后端配置环境变量 `PTLPOJ_ADMIN_TOKEN = xxx` 决定。当空置时探测默认值为 `"ptlpoj_default_admin"` 以降低试用门槛。
- **中间件“双盲验证”**：`RequireAdminToken` 拦截属于 `/api/admin/*` 的所有路由，现要求 HTTP Header 携带双重口令：即 `Authorization: Bearer <UserJWT>` 以及 `X-Admin-Token: <AdminToken>`。后端会进一步反查提取出的 User ID 是否具备 `ADMIN` 角色（Role）。
- **客户端储藏与重发**：避免秘钥在设置面板中被明文看到，插件通过 VS Code 暴露的 `context.secrets` 构建安全的操作系统级密钥链 (Keychain) 存储机制。支持遇到 401 时主动拦截并下发原生警告框引导重置命令 `ptlpoj.resetAdminToken`，规避 CSP 对原生 Webview Alert/Confirm 的静默屏蔽问题。

### 3.2 可视化白名单与批量操控
- **Webview UI**：创建极简后台。具备表格形态展示目前白名单与 Role。
- **批量并发引擎**：前端 `textarea` 框支持通过正则表达式 `/[,\n; ]+/` 抓取并切分大段文本，自动把如 “user1@qq.com, user2@gmail.com” 等转换为批量并发的 POST (添加) 或 DELETE (删除) 动作推向 `/api/admin/users`。自带进度条阻断与局部回显能力。

### 3.3 题库自动化解析引擎 (AST Parse Worker)
- **痛点**：题面的 Markdown 文件、基础代码脚手架 Scaffold 以及用于 Docker 沙盒验证的 Hidden Tests 在旧版体系中需要三份隔离的文件进行人工维护。
- **Python 解析策略**：引入 `parse_worker.py`。该脚本直接调用系统 `ast` 抽象语法树模块。其功能包括：
  - 读取函数级或模块级 Docstring，并将其转译为题面 Markdown。
  - 读取 Doctest 断言块，并将其转化为标准评测使用的 `tests.txt` 验证数据。
  - 移除原函数实现体，仅保留函数签名与脚手架结构（如 `def solution(): ...`），并下发给用户侧。
- **并发批量上传**：为了规避常规 Webview 环境中 HTML `<input type="file">` 标签受限于沙箱安全机制而无法获取绝对路径的限制，系统通过宿主通信机制抛出 `selectAndUploadProblems` 事件，调用 VS Code 原生的文件选取组件 `window.showOpenDialog({canSelectMany: true})` 选取批量 `.py` 源码，由扩展程序执行轮询与并发上传。

## 4. 后续交付清单 (Execution Plan)
*(已交付至 v0.3.1 的核心里程碑)*
- [x] 开发并覆盖全量 `AdminToken` 与鉴权 Fallback 机制。
- [x] 开发 Python AST 剥析器。
- [x] 上线 `AdminViewPanel` 并对接原生文件框。
- [x] 开发前后端配套 CRUD API 通信层。

---
*Created during Phase 13-14 for system empowerment.*
