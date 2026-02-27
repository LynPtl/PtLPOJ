# PtLPOJ 开发日志 - Phase 4: REST API 与 SSE 通信层 (Transport Layer)

**日期**: 2026-02-24
**阶段**: Phase 4 (REST API & Transport Layer)
**状态**: ✅ 已完成

## 1. 本阶段核心构建
Phase 4 的核心使命是将底层强大的 SQLite 仓储模块、Docker 沙盒执行引擎、以及严密的认证模块桥接成对外暴露的 Web 服务。这一切都是为了给下一阶段（VS Code 客户端）打造稳固的信息“高速公路”。

我们没有使用类似 Gin 或 Fiber 这样庞大的第三方 Go Web 框架，而是遵循极致轻量的原则，仅仅依托 Go 语言标准库 `net/http` 就实现了全套 REST 语义和流式传输机制。

## 2. API 架构演进与全功能覆盖

### 2.1 智能状态透传 (Problem Fetching)
在 `GET /api/problems` 接口中，难点不在于拿题目，而在于**基于不同登录用户实现智能的试卷打码**：
- 当用户请求题库时，系统会基于 JWT 中的 `user_id` 在 `submissions` 表里快速聚合这名用户的历史提交。
- 采用智能的“降级状态匹配”：如果这道题做对过一次就标 `AC` 绿灯，无论后续再怎么错；仅当一次都没成功过时标记 `WA`，或者标记 `UNATTEMPTED`。
- 在 `GET /api/problems/{id}` 中，还会向用户额外下发该题目配对的 Markdown 题面和 Python 初始化函数代码块片段 (Scaffold)，使 VS Code 的离线体验无比顺滑。

### 2.2 评测提交与解耦背压 (Submission Queueing)
`POST /api/submissions` 实现了一套高度解耦的**异步提交流**。
请求到达后只会：
1. 校验负载和安全上下文。
2. 调用持久层创建记录并置为 `PENDING` 初始态。
3. 立即切断 HTTP 握手，直接返回 `202 Accepted`。
这种极简的高背压忍受架构，完美规避了前端“转圈圈等待 Docker 执行完毕导致 API Gateway 网关超时”的惨案。任务将自然转移给 Phase 2 编写的 Worker 进程池缓慢发酵消化。

### 2.3 Server-Sent Events (SSE) 长连接监控
为了让 VS Code 等到判题完成并实时获知 `RUNNING -> TLE/AC` 的激变，我们在 `GET /api/submissions/{id}/stream` 中实现了真正的 HTTP 服务器推送 (Server-Sent Events):
- 这是本项目最现代化的 Web 通信亮点之一。在 Go 中使用 `http.Flusher` 劫持 HTTP 连接通道不断网。
- 采用轻量化的**轮询退避 (Polling)**：每 500ms 短睡眠查询一次 SQLite。虽然这不是通过 Redis Pub/Sub 触发，但对于本系统的内网规模而言，极大地降低了外部中间件依赖的复杂性。
- 代码里设定了 60 秒的硬超时防呆熔断。并在结束时推流 `event: complete` 以通知前端掐断连接防泄漏。

### 2.4 内核池断电修复 (Crash Recovery)
系统必须具备反脆弱性。我在并发调度池源头 `scheduler/worker.go` 中，增补了一个**孤儿防漏池回收函数 (`recoverOrphanedSubmissions`)**。
如果在大量学生代码正跑在 Docker 沙盒内时，机房遭遇断电停机或主程序崩溃：
- 当 Go 程序再次重启初始化时（先于 Worker 拉起），它会火线清扫数据库，将大量意外永远卡死在 `RUNNING` 态的幽灵试卷，重新降级为 `PENDING`。
- 此时 Worker 队列苏醒，它们顺理成章地被重新接管进入二次评测，避免了数据库“烂尾案”发生。

## 3. 测试与总结
我们编写了 `handler_test.go` 并利用原生的 `httptest` 进行虚拟拨号 HTTP 请求注入：
- 虚拟注册了用户进行拦截测试。
- 成功捕获到了携带状态字段的 `ProblemResponse` 和接收判题任务。
配合执行 `go test ./...` 会发现底层沙盒、仓储到最顶层 Handlers 完美绿灯通过。后端的使命已全部完美交付。

下一阶段，我们将使用 TypeScript 等前端语言建立最终客户端，结束这场史诗般的从 DB 到客户端的垂直穿透开发。
