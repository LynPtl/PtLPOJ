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

## 4. 实施阶段与达成情况 (Implementation Status)

项目所有核心阶段均已完成开发、验证与本地封包。

### Phase 1: 数据模型与存储层 (Data Layer) [DONE]
- [x] sqlite3/gorm 架构与 WAL 模式开启。
- [x] 静态题库解析与缓存系统。

### Phase 2: 沙盒评测引擎 (Sandbox Engine) [DONE]
- [x] Docker SDK 集成与容器生命周期管理。
- [x] 基于 cgroups 的资源硬隔离与网络静默。
- [x] 实时内存峰值统计 (Streaming Docker Stats)。

### Phase 3: 鉴权与安全性 (Auth & Security) [DONE]
- [x] OTP 验证流与 mock 邮件分发。
- [x] JWT 签发与鉴权中间件。
- [x] IP 级令牌桶限流。

### Phase 4: 传输层 (Transport Layer) [DONE]
- [x] RESTful API 路由实现。
- [x] SSE (Server-Sent Events) 判题进度推流。
- [x] 服务崩溃恢复逻辑。

### Phase 5: VS Code 客户端 (VS Code Extension) [DONE]
- [x] TreeView 题库展示与本地文件同步 (Sync)。
- [x] 沉浸式分屏刷题 (Immersive Workarea)。
- [x] VSIX 本地离线安装包构建。

### Phase 6: 全链路验收与加固 (Testing & Hardening) [DONE]
- [x] E2E 完整生命周期联调。
- [x] TLE/WA/RE 边界情况模拟。
- [x] 关键路由与内存采集 Bug 修复。
