# Phase 15 开发日志：队列重构与代码健康度优化 (v0.3.2)

## 1. 阶段概述

本次开发重构了判题队列的核心模式，并完成了多项代码健康度改进。主要目标是用 Go Channel 替代原有的 SQL 轮询模式，消除数据库的无效负载，同时提升系统整体的鲁棒性。

## 2. 核心改动

### 2.1 SQL 队列 → Go Channel 队列

**背景**：原实现使用 SQLite `submissions` 表的 `status` 字段作为队列状态，Worker 每 2 秒轮询一次 DB 获取 PENDING 任务，SSE 每 500ms 轮询 DB 推送结果。这种模式效率低、延迟高、浪费数据库 IO。

**实现**：
- 新增 `scheduler.JobQueue` 结构体，包含 `jobs chan`（带缓冲的 submission ID 队列）和 `sseSubs map[uuid.UUID]chan Result`（SSE 订阅表）
- Worker 从 channel 拉取任务（fan-out 模式），不再轮询数据库
- 评测完成后通过 `sseSubs` 定向推送给对应 SSE 客户端，实现真正的实时推送
- Submission 达到终态后自动清理 SSE 订阅，后台 goroutine 每 5 分钟清理孤儿订阅

**相关文件**：`server/scheduler/worker.go`、`server/api/handlers/submission_handler.go`

### 2.2 SSE 推送改造

**问题**：原 SSE handler 每 500ms 轮询 DB 判断状态，通过重复发送 `event: complete` + `data: {"finished": true}` 导致客户端弹两次框（第二次因数据不完整显示 undefined）。

**修复**：
- 移除了 500ms ticker 轮询
- 改为订阅 `GlobalQueue.Subscribe(subID)` channel 等待结果
- `event: complete` 不再携带 data payload，避免客户端重复解析

### 2.3 TLE 判断修复

**问题**：原实现用墙钟时间（wall-clock time）判断 TLE，包含容器启动、Python 解释器初始化等开销。并发增加时容器调度变慢，墙钟时间超过 `timeLimitMs` 即触发 TLE，导致正常代码被判超时。

**修复**：
- 收集 `CPUStats.CPUUsage.TotalUsage`（纳秒级累计 CPU 时间）
- TLE 判断改用容器实际 CPU 时间，不再受容器调度延迟影响
- 容器等待超时从 `timeLimitMs + 500ms` 增加到 `timeLimitMs + 2000ms`，给启动留足时间

**相关文件**：`server/sandbox/runner.go`

### 2.4 Problem 缓存自动 Reload

**问题**：启动时加载 `problems.json` 后永不更新，管理员通过 VS Code 插件更新题库后需要重启服务。

**实现**：后台 goroutine 每 30 秒检查 `problems.json` 的 mtime，有变化则自动重新加载缓存。

**相关文件**：`server/storage/problem_repo.go`

### 2.5 Rate Limiter LRU

**问题**：原 `map[string]*rate.Limiter` 随 IP 访问无限增长，存在内存耗尽风险。

**实现**：引入 `hashicorp/golang-lru` 的 ARC（Adaptive Replacement Cache），限制最大 10000 个 IP 条目，超出后自动淘汰最久未使用的记录。

**相关文件**：`server/middleware/ratelimit.go`

### 2.6 Submission 统计查询优化

**问题**：`getBestUserStatus` 加载所有 submission 再遍历，存在 N+1 查询问题。

**实现**：改为 3 次精准 `COUNT(*)` 聚合查询替代全量加载。

**相关文件**：`server/api/handlers/problem_handler.go`

## 3. 鲁棒性改进

### 3.1 Worker Panic Recovery

Worker goroutine 添加 `recover()` 处理，panic 后自动重启，防止单个 worker 挂掉导致队列积压。

### 3.2 Graceful Shutdown

服务收到 `SIGINT`/`SIGTERM` 时：
1. 停止接收新连接（`http.Server.Shutdown`）
2. 等待 SSE 连接自然关闭
3. 关闭 JobQueue（停止 workers）
4. 数据库连接安全关闭

### 3.3 Sandbox Stats Goroutine 资源泄漏修复

原 `runner.go` 中 stats goroutine 在提前返回时未关闭 HTTP response body，导致 HTTP 连接泄漏。修复为 `defer stats.Body.Close()` 确保执行路径都会关闭。

## 4. 沙箱安全加固：用户 print() 输出隔离

**问题**：用户的 `print()` 调用会污染 doctest 的输出，导致评测结果误判。

**实现**：在用户代码执行前，替换 `builtins.print` 为静默函数，doctest 运行前恢复原函数。这样用户的 stdout 输出被丢弃，不影响 Go 端对 doctest 输出的解析。

**相关文件**：`server/sandbox/injector.go`

## 5. 容器化探索记录

本次开发过程中曾探索通过 `docker-compose` + `sysbox` 运行时实现安全的容器化部署方案，但在 WSL2 + Docker Desktop 环境下遭遇兼容性问题：Docker Desktop 的 dockerd 运行在独立 Linux VM 中，无法通过 `/etc/docker/daemon.json` 配置 sysbox 运行时。因此容器化方案暂时搁置，保持宿主机裸机部署架构。

## 5. 交付总结

本次迭代显著提升了系统的架构质量和鲁棒性，消除了原有的 DB 轮询瓶颈，实现了真正的实时评测推送。所有改动均已通过测试验证并推送至 master 分支。
