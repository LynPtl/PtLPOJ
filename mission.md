# Project Specification: Local Python Online Judge (VS Code + Go)

## 1. 项目概述 (Project Overview)

本项目旨在为内部团队开发一款基于局域网部署的轻量级 Python 专属 Online Judge (OJ) 系统。项目核心目标是实现**测试数据的绝对物理隔离**与**评测过程的高效自动化**，抛弃传统的 Web 前端，通过 VS Code 插件提供沉浸式的开发者体验。

### 1.1 核心需求边界

* **权限管控 (Access Control):** 采用无密码的 OTP（一次性密码）邮箱验证机制，结合 JWT 实现接口级鉴权。基于白名单/指定域名拦截非法注册。
* **数据防泄漏 (Data Loss Prevention):** 题目列表与详情采用懒加载分离；测试用例（`.in/.out`）仅在服务端沙盒内存周期内可见，严格禁止向客户端下发具体的报错用例数据。
* **语言限制:** 仅支持 Python 3，免去编译阶段开销，缩短判题生命周期。

## 2. 架构设计与技术栈 (Architecture & Tech Stack)

### 2.1 核心组件划分

* **Client (VS Code Extension):** 基于 TypeScript。负责文件系统 I/O（生成代码模板）、Tree View UI 渲染、状态流转展示。
* **API Gateway & Middleware:** 基于 Go。负责 JWT 鉴权拦截、基于 IP/User 的 Rate Limiting 限流。
* **Judge Server (调度核心):** 基于 Go。处理并发提交，维护任务队列，调度底层容器生命周期。
* **Sandbox (评测沙盒):** 基于 Docker Engine API。动态拉起 `alpine-python3` 容器，利用 `cgroups` 限制 CPU/Memory。
* **Storage (数据持久化):** * 关系型数据：SQLite (存储 User, Token, Submission 状态)。
* 题目与用例：本地 File System (存储题单 JSON 索引、Markdown 题面、`.in/.out` I/O 文件)。



### 2.2 通信协议

* **常规业务流:** 标准 HTTP RESTful API (JSON Payload)。
* **异步评测反馈:** Server-Sent Events (SSE) 单向长连接推送。

## 3. 核心业务流转 (Core Workflow)

1. **鉴权 (Authentication):** 用户在插件端输入邮箱 -> Server 校验白名单并发送 OTP -> 用户提交 OTP -> Server 签发 JWT。
2. **初始化 (Provisioning):** 插件携带 JWT 请求 `/problems` 接口渲染侧边栏；用户点击节点，请求 `/problems/{id}` 获取 Markdown 描述与初始代码，并写入本地 `.py` 文件。
3. **提交与评测 (Submission & Execution):** * 插件 POST 代码至 `/submissions`。
* Server 生成 `submission_id` 并将任务压入 Channel 队列。
* Judge Worker 取出任务，通过 Docker SDK 拉起隔离容器，挂载 I/O 用例进行黑盒测试。


4. **状态推送 (Notification):** 容器销毁，Server 比对 stdout 与 `.out` 文件生成 AC/WA/TLE 等状态，通过 SSE 接口向插件推送最终结果。

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
  * **(防逃逸)** 设置 `CapDrop: ["ALL"]` 移除特权，并低权限用户运行；开启 `ReadonlyRootfs: true`，仅挂载评测执行必须的 I/O 目录为读写。
* [ ] **2.3 评测流与熔断实现:** 编写 Worker 协程，负责执行环境拉起与流捕获：
  * **(防 OLE)** 获取 `stdout/stderr` 时包裹 `io.LimitReader` 实施严格的上限截断，触发即拦截判为 OLE。
  * **(防超时)** 在调用 Docker SDK API 时传入基于 Wall-Time 超时的 `Context`，主动猎杀类似 `time.sleep` 等待阻塞攻击的沙箱。
* [ ] **2.4 健壮 Diff 逻辑:** 实现容错型结果对比函数，支持自动忽略行末不可见字符(空格/制表符)及文件末尾的空余换行。

### Phase 3: 鉴权模块与安全性加固 (Auth & Security)

* [ ] **3.1 邮件服务集成:** 封装 SMTP 或云端邮件 API（如 AWS SES），实现 OTP 验证码的生成与下发机制（含 5 分钟 TTL 缓存逻辑）。
* [ ] **3.2 JWT 签发与鉴权:** 完成 `/login` 和 `/verify` 接口。
* [ ] **3.3 中间件开发:** 编写 Go Middleware，实现 JWT Header 拦截与基础的 Token Bucket 限流策略。

### Phase 4: REST API 与 SSE 通信层 (Transport Layer)

* [ ] **4.1 题目分发接口:** 实现带鉴权的 `GET /problems` 和 `GET /problems/{id}` 接口。**服务端除了返回题目描述，需要附加下发基于 Python 语言的初始化代码脚手架 (Code Scaffolding/Function Signature)。**
* [ ] **4.2 提交接收接口:** 实现 `POST /submissions` 接口，完成请求 Payload 解析并推入评测队列。
* [ ] **4.3 SSE 推送接口:** 实现 `GET /submissions/{id}/stream`，接管 `http.Flusher`，将评测结果异步推流至客户端。
* [ ] **4.4 消息队列容灾恢复:** 增加服务初始化的崩溃恢复逻辑 (Crash Recovery)——系统启动时读取 SQLite，重新拉起由于意外断电导致的、状态驻留在 `PENDING/RUNNING` 的孤儿判题。

### Phase 5: VS Code 插件开发 (Client Application)

* [ ] **5.1 工程搭建与安全凭据:** 搭建 TypeScript 框架，实现交互。**切记使用 VS Code 核心的 `context.secrets.store()` 进行 JWT Token 的安全保管与读取**。
* [ ] **5.2 HTTP 客户端与主动轮询降级:** 封装带自动注入 JWT header 的 Axios 请求；开发 SSE 的 Fallback 机制——若等待状态超限，提供长轮询或主动 API 请求拉取最终成绩防假死。
* [ ] **5.3 智能文件呈现:** 监听树节点点击事件，通过 VS Code API 读取云端 Markdown 和题目初始代码脚手架模板，在本地动态组合并以无缝的体验唤起 `.py` 编辑画面。
* [ ] **5.4 提交与推流订阅:** 绑定 `Submit` 快捷键，打包代码发起 POST 请求，并建立 EventSource 监听 SSE 结果渲染到右下角通知或输出面板。

### Phase 6: 系统联调与测试验证 (E2E & Hardening)

* [ ] **6.1 端到端联调:** 在内网环境下跑通“登录 -> 拉题 -> 提交 -> 返回结果”的完整主干链路 (Happy Path)。
* [ ] **6.2 异常用例测试:** 故意提交包含死循环、内存泄漏、恶意系统调用的 Python 代码，验证 Sandbox 的熔断机制。
* [ ] **6.3 压力测试:** 模拟多用户并发提交，验证 Go 后端 Channel 队列的背压（Backpressure）表现。
