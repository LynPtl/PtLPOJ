# Project Specification: Local Python Online Judge (VS Code + Go)

## 1. 项目概述 (Project Overview)

本项目旨在为内部团队开发一款基于局域网部署的轻量级 Python 专属 Online Judge (OJ) 系统。项目核心目标是实现**测试数据的绝对物理隔离**与**评测过程的高效自动化**，抛弃传统的 Web 前端，通过 VS Code 插件提供沉浸式的开发者体验。

### 1.1 核心需求边界

* **权限管控 (Access Control):** 采用无密码的 OTP（一次性密码）邮箱验证机制，结合 JWT 实现接口级鉴权。基于白名单/指定域名拦截非法注册。
* **数据防泄漏 (Data Loss Prevention):** 题目列表与详情采用懒加载分离；隐藏用的测试用例数据仅在服务端沙盒内存周期内可见，严格禁止向客户端下发具体的报错用例数据。
* **语言限制:** 仅支持 Python 3，免去编译阶段开销，缩短判题生命周期。

## 2. 架构设计与技术栈 (Architecture & Tech Stack)

### 2.1 核心组件划分

* **Client (VS Code Extension):** 基于 TypeScript。负责文件系统 I/O（生成代码模板）、Tree View UI 渲染、状态流转展示。
* **API Gateway & Middleware:** 基于 Go。负责 JWT 鉴权拦截、基于 IP/User 的 Rate Limiting 限流。
* **Judge Server (调度核心):** 基于 Go。处理并发提交，维护任务队列，调度底层容器生命周期。
* **Sandbox (评测沙盒):** 基于 Docker Engine API。动态拉起 `alpine-python3` 容器，利用 `cgroups` 限制 CPU/Memory。
* **Storage (数据持久化):** * 关系型数据：SQLite (存储 User, Token, Submission 状态)。
* 题目与用例：本地 File System (存储题单 JSON 索引、Markdown 题面、tests.txt 隐藏测试用例)。



### 2.2 通信协议

* **常规业务流:** 标准 HTTP RESTful API (JSON Payload)。
* **异步评测反馈:** Server-Sent Events (SSE) 单向长连接推送。

## 3. 核心业务生命周期 (Core Workflow & Lifecycle)

### 3.1 鉴权与会话建立 (AuthN & AuthZ)
* **Trigger (触发):** 客户端（VS Code 插件）发起登录请求，提交用户邮箱。
* **Validation (校验):** Server 侧网关拦截请求，查阅本地白名单表（Whitelist）。未命中则直接返回 `403 Forbidden`。
* **OTP Generation (凭证生成):** 命中白名单后，Server 生成带 TTL（如 5 分钟）的随机 6 位验证码，存入内存缓存，并通过 SMTP (或控制台) 下发至目标邮箱。
* **Token Issuance (会话签发):** 客户端提交验证码，Server 校验一致后，签发 JWT (JSON Web Token)。此后客户端所有后续请求均需在 Header 中携带 `Authorization: Bearer <JWT>`。

### 3.2 元数据同步与按需加载 (Metadata Sync & Lazy Loading)
* **Index Fetch (拉取索引):** 客户端携带 JWT 请求`/problems` 接口。Server 返回包含该用户评测状态的 JSON 数组（带 `problem_id`, `title`, `difficulty`, 当前状态等），用于前端渲染 Tree View 和图标。
* **Detail Fetch (懒加载详情):** 用户点击特定题目，客户端发起详情请求。Server 下发对应的 Markdown 题面描述和 Python 初始化代码片段（Scaffold）。
* **Local Provision (本地初始化):** 客户端根据下发的数据，在用户本地工作区动态拼装并展示题目描述与代码编辑界面。

### 3.3 本地开发阶段 (Local Development)
* **Offline Coding (离线编写):** 此阶段为纯客户端行为。用户在本地 VS Code 环境中编写 Python 核心逻辑代码，也可运行本地自带的 Public Test，后端无压力。

### 3.4 提交与沙盒评测 (Submission & Sandbox Execution)
* **Payload Submission (投递):** 用户触发提交，客户端将 `{ "problem_id": 123, "source_code": "..." }` 封装为 JSON POST 至 Server。
* **Job Queuing (任务入队):** Server 接收并落库（状态记为 `PENDING`），将评测任务压入内部缓冲队列（Go Channel），并立即向客户端返回 `submission_id`。
* **Sandbox Provision (环境初始化):** Judge Worker 取出任务，调用 Docker API 启动极轻量运行时容器 `alpine-python3`。
* **Resource Limits & Isolation (安全硬化):** 严格配置 `cgroups`（内存及 CPU Quota限制），移除容器网络（`NetworkDisabled: true`）防止反弹 Shell，并丢弃特权权限。
* **Execution & I/O (动态注入与执行):** 将服务端持有的隐藏测试用例脚本注入到用户代码中，通过 `doctest` 静默运行验证。捕获进程 `stdout/stderr`，并实施时间截断熔断。
* **Teardown (环境销毁):** 进程结束或超时，立即强制回收容器。

### 3.5 判题与结果反馈 (Evaluation & Notification)
* **Evaluation (判题分析):** Judge 服务收集退出码及执行结果，更新 Submission 库的最终状态（`AC`, `WA`, `TLE`, `RE` 等）。
* **Data Masking (数据脱敏):** 组装处理结果时严格执行脱敏：如果是 `WA` 或 `RE`，绝对不包含触发崩溃的具体隐藏用例的输入数据，直接斩断学生企图暴力枚举或特判打表的可能性。
* **Client Render (前端推流渲染):** 根据长连接拉取或 SSE (Server-Sent Events) 单向推送机制，将最终结果动态反馈给客户端，VS Code 进行答题卡状态的渲染更新。

---

## 4. 实施计划 (Implementation Walkplan)

项目落地将划分为 6 个阶段（Phases），遵循自底向上（Bottom-up）的构建逻辑，确保每个阶段均可进行单元测试或接口验证。

### Phase 1: 数据模型与存储层初始化 (Data Layer)

* [ ] **1.1 定义静态数据结构:** 制定本地题目配置规范（设计 `problems.json` 的 Schema 以及测试用例目录结构）。
* [ ] **1.2 关系型数据库 Schema 设计:** 在 Go 中使用 GORM 或原生 SQL 编写 SQLite 建表脚本（包含 `users`, `submissions` 表）。
* [ ] **1.3 基础 CRUD 封装:** 完成服务端对 SQLite 和本地题目配置的读取逻辑封装。
* [ ] **1.4 数据库并发写优化:** 在 SQLite DSN 连接配置中显式开启 Write-Ahead Logging (WAL) 模式，以支持多个 Judge Worker 的并发状态更新，防止 `Database is locked` 错误。

### Phase 2: 沙盒评测引擎开发与安全加固 (Sandbox Engine)

* [ ] **2.1 Docker SDK 集成:** 在 Go 服务中引入 Docker API SDK，编写函数实现镜像拉取与容器生命周期管理（Create, Start, Stop, Remove）。
* [ ] **2.2 隔离限制与高阶安全防护:** 
  * 显式声明 `NetworkDisabled: true`，利用 `HostConfig` 设置 Memory/CPU 限制。
  * **(防爆炸)** 添加 `PidsLimit` 防止 Fork Bomb 耗尽宿主机资源。
  * **(防逃逸)** 设置 `CapDrop: ["ALL"]` 移除特权，并以低权限非 root 用户拉起执行；开启 `ReadonlyRootfs: true`，不挂载任何写权限目录。
* [ ] **2.3 评测流与熔断实现:** 编写 Worker 协程，负责执行环境拉起与流捕获：
  * **(防 OLE)** 获取 `stdout/stderr` 时包裹 `io.LimitReader` 实施严格的上限截断，触发即拦截判为 OLE。
  * **(防超时)** 在调用 Docker SDK API 时传入基于 Wall-Time 超时的 `Context`，主动猎杀类似 `time.sleep` 等待阻塞攻击的沙箱。
* [ ] **2.4 评测结果解析 (Result Parsing):** 解析 `doctest` 静默运行后的输出流，精确捕获 `Exit Code` 与控制台报错信息，提取第一处 Failed 异常点以定位 `failed_at_case` 序号。

### Phase 3: 鉴权模块与安全性加固 (Auth & Security)

* [ ] **3.1 邮件服务集成:** 封装 SMTP 伪装适配层（当前开发阶段仅在服务端 Console 打印 OTP）。**关键代办**：未来需要对接真实的 SMTP 服务器 (如 AWS SES, 腾讯云邮件) 完成真正的邮件下发，防止遗忘。
* [ ] **3.2 JWT 签发与鉴权:** 完成 `/login` 和 `/verify` 接口。
* [ ] **3.3 中间件开发:** 编写 Go Middleware，实现 JWT Header 拦截与基础的 Token Bucket 限流策略。

### Phase 4: REST API 与 SSE 通信层 (Transport Layer)

* [ ] **4.1 题目分发接口:** 实现带鉴权的 `GET /problems` 和 `GET /problems/{id}` 接口。**重点：除了返回题目描述和下发 Python 代码脚手架 (Scaffolding)，还需要在列表中附带该用户对每道题的评测状态 (例如 AC 通过 / WA 未通过 / 未尝试)，供前端渲染完成度。**
* [ ] **4.2 提交接收接口:** 实现 `POST /submissions` 接口，完成请求 Payload 解析并推入评测队列。
* [ ] **4.3 历史记录接口:** 实现 `GET /submissions` 接口，允许获取当前用户的历史提交记录列表（含代码片段与每次评测的详细状态结果）。
* [ ] **4.3 SSE 推送接口:** 实现 `GET /submissions/{id}/stream`，接管 `http.Flusher`，将评测结果异步推流至客户端。
* [ ] **4.4 消息队列容灾恢复:** 增加服务初始化的崩溃恢复逻辑 (Crash Recovery)——系统启动时读取 SQLite，重新拉起由于意外断电导致的、状态驻留在 `PENDING/RUNNING` 的孤儿判题。

### Phase 5: VS Code 插件开发 (Client Application)

* [ ] **5.1 工程搭建与安全凭据:** 搭建 TypeScript 框架，实现交互。**切记使用 VS Code 核心的 `context.secrets.store()` 进行 JWT Token 的安全保管与读取**。
* [ ] **5.2 HTTP 客户端与主动轮询降级:** 封装带自动注入 JWT header 的 Axios 请求；开发 SSE 的 Fallback 机制——若等待状态超限，提供长轮询或主动 API 请求拉取最终成绩防假死。
* [ ] **5.3 智能文件呈现与答题卡展示:** 根据后端返回的题目状态，在 VS Code 侧边栏的 Tree View 中为每道题挂载不同的状态图标 (例如 ✅ 已通过、❌ 报错、⚪ 未尝试)。同时点击节点后，动态拉取并拼接 Markdown 题面与代码模板。
* [ ] **5.4 提交与推流订阅:** 绑定 `Submit` 快捷键，打包代码发起 POST 请求，并建立 EventSource 监听 SSE 结果渲染到右下角通知或输出面板。

### Phase 6: 系统联调与测试验证 (E2E & Hardening)

* [ ] **6.1 端到端联调:** 在内网环境下跑通“登录 -> 拉题 -> 提交 -> 返回结果”的完整主干链路 (Happy Path)。
* [ ] **6.2 异常用例测试:** 故意提交包含死循环、内存泄漏、恶意系统调用的 Python 代码，验证 Sandbox 的熔断机制。
* [ ] **6.3 压力测试:** 模拟多用户并发提交，验证 Go 后端 Channel 队列的背压（Backpressure）表现。
