# PtLPOJ System Architecture & Design

本文档详细描述了 PtLPOJ (Personal/Private Local Python Online Judge) 的系统架构、组件交互方式以及核心技术选型背后的思考。

---

## 1. 宏观架构概览 (High-Level Architecture)

PtLPOJ 采用经典的 **Client-Server-Sandbox (C-S-S)** 三层解耦架构，专为局域网内的内部团队或教学场景设计。它的核心设计理念是**“重判题，轻前端”**，通过 VS Code 插件提供沉浸式的开发者体验。

```mermaid
graph TD
    %% Client Tier
    subgraph " Client (Developer Machine) "
        VSCode[VS Code Extension]
        Editor[Monaco Editor / Local Workspace]
        Webview[Rich Problem Webview]
    end

    %% Network & Auth
    subgraph " Network Boundary "
        AuthMiddleware((JWT / OTP Auth))
        RateLimiter((Rate Limiter))
    end

    %% Server Tier
    subgraph " Judge Server (Golang) "
        Gateway[API Gateway]
        JobQueue(Internal Job Queue / Channels)
        Scheduler[Judge Scheduler / Workers]
        SSE[SSE Streamer]
    end

    %% Storage Tier
    subgraph " Data & Storage "
        SQLite[(SQLite WAL)]
        FileSystem[(Local Problems & Tests Files)]
    end

    %% Sandbox Tier
    subgraph " Isolation Environment "
        DockerEngine[Docker Engine API]
        Container1(alpine-python3 Sandbox)
        Container2(alpine-python3 Sandbox)
    end

    %% Connections
    VSCode ---> | POST /api/submissions | RateLimiter
    RateLimiter ---> AuthMiddleware
    AuthMiddleware ---> Gateway

    VSCode <... | Server-Sent Events | SSE

    Gateway ---> | Read/Write | SQLite
    Gateway ---> | Read | FileSystem
    Gateway ---> | Push | JobQueue

    JobQueue ---> Scheduler
    Scheduler ---> | Create/Start/Stop | DockerEngine
    Scheduler ---> | Stream Logs | DockerEngine
    Scheduler ---> | Update State | SQLite
    Scheduler ---> | Notify Push | SSE

    DockerEngine ---> Container1
    DockerEngine ---> Container2
```

---

## 2. 核心组件详解 (Component Deep Dive)

### 2.1 客户端层：VS Code 插件 (Client Tier)
* **职责**：替代传统的 Web 浏览器前端，负责身份验证交互、题目树状导航、Markdown 富文本渲染以及代码的本地读写与提交。
* **技术选型**：`TypeScript` + `VS Code Extension API`。
* **安全凭据管理**：JWT Token 绝不硬编码或存放在普通配置文件中，而是利用 `context.secrets.store()` 写入宿主操作系统的高级凭据管理器 (如 Windows Credential Manager, macOS Keychain)。
* **离线容忍度**：题目模板拉取到本地工作区后，断网状态下依然能够继续编写核心逻辑代码或使用本地 Python 解释器进行调试。

### 2.2 接入层：网关与中间件 (API Gateway & Middleware)
* **职责**：防刷限流与访问控制。
* **限流 (Rate Limiting)**：基于 Golang 内置的 `x/time/rate` 令牌桶算法。特别是在登录/OTP 验证节点，限制同 IP 的高频重试，防止暴力破解。
* **无密鉴权 (Passwordless Auth)**：采用基于邮箱的动态验证码 (OTP) 体系，系统验证成功后签发无状态的 JWT (JSON Web Token)，包含用户的 `UUID` 和 `Role` 声明。

### 2.3 核心业务与调度层 (Judge Server)
* **职责**：处理 RESTful 请求、维护高压判题队列、并协调底层执行器。
* **技术选型**：`Golang` 的高并发模型。
* **背压设计 (Backpressure)**：当学生短时间进行海量提交（如 100 份/秒），Server 不会立即开启 100 个沙盒压垮宿主机。提交记录落库为 `PENDING` 状态后，会压入带缓冲的 Go Channel。固定数量的 Judge Worker 协程循环消费队列，平滑削峰。
* **断电恢复 (Crash Recovery)**：系统在 Boot 引导阶段，会主动轮询 SQLite 中遗留的驻留状态为 `RUNNING` 甚至 `PENDING` 的孤儿判题，并重新推入消费队列。

### 2.4 沙盒引擎层 (Sandbox Engine)
* **职责**：提供对不受信学生代码的绝对隔离与执行计时。
* **技术选型**：`Docker Engine API SDK`，通过 Go 直接编排容器。
* **安全硬化 (Hardening Rules)**：
    * **NetworkDisabled: true**：拔除容器虚机网卡，斩断反向 Shell 和外发接口调用的可能。
    * **cgroups**：利用 Linux `cgroups` 严格划定容器内存上限 (如 256MB) 并在触发时由内核直接 `SIGKILL` (得到 OLE 评判)；设定 CPU 绝对运行周期限定 (Quota)。
    * **CapDrop ["ALL"] & User Namespace**：卸载一切 root 特权，以最低权限用户启动。
    * **Wall-Time 熔断**：在发起 `Docker Wait` 等待进程退出时，注入 Go `Context.WithTimeout`。一旦超过 5 秒物理时间（防死循环 `while True` 或休眠 `time.sleep`），直接调用 Docker Remove 强杀容器。
    * **I/O 截断**：对于程序的 `stdout` 和 `stderr` 流捕获，使用 `io.LimitReader` 包裹，防止学生恶意打桩输出数百兆垃圾数据撑爆服务端内存空间。

### 2.5 数据存储层 (Data Layer)
* **关系数据**：采用 `SQLite` 单体数据库。利用 `GORM` 框架管理 `Users`, `Submissions`, `Problems` 等结构化状态。
    * **并发写优化**：为了避免 Go 协程并发更新判题状态遇到的 `Database is locked` 问题，初始化 DSN 时强制注入 `_journal_mode=WAL`（Write-Ahead Logging）和高容忍阈值的忙等待重试机制 `_busy_timeout=5000`。
* **非结构化文件**：题目的 Markdown 描述文本以及最为私密的隐藏验证数据（`tests.txt`），以纯文件形式存储在服务端配置目录下（如 `data/problems/1001/`）。这些文件仅在被 Worker 拉起评测时挂载或直接读入内存并注入代码流。

---

## 3. 核心设计妥协与思考 (Design Trade-offs)

1. **为什么用 SQLite 而不是 MySQL/PostgreSQL？**
   本系统的首要目标是“极简部署”，特别针对中小规模班级教学场景或内部小团队（<200 人）。SQLite 的 WAL 模式已经能够完美承载轻量级的高并发写请求，免去了运维数据库集群的心智负担。
2. **为什么彻底抛弃 Web 前端？**
   为了保证开发“沉浸流”。在 Web 浏览器上写算法题总是面临没有智能提示（IntelliSense）、无法引入第三方包在本地做验证的问题。在 VS Code 侧完成闭环是下一代开发者工具的标准。
3. **为什么使用 SSE 代替 WebSocket？**
   对于判题结果推流，需求是 **单向的服务器至客户端通知**（仅仅告诉客户端判题的阶段性进度）。`Server-Sent Events (SSE)` 架构比 WebSocket 更轻量、占用连接资源更少，天生支持断线重连标准，特别切合该场景。
