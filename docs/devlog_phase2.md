# PtLPOJ 开发日志 - Phase 2: 沙盒评测与引擎执行层 (Sandbox Engine)

**日期**: 2026-02-24
**阶段**: Phase 2 (Sandbox Engine)
**状态**: ✅ 已完成

## 1. 本阶段核心突破
Phase 2 是本项目技术难度最高的核心组件。我们在没有任何现成评测机依赖的情况下，使用原生 Docker Go SDK 从零打造了一个并发安全、极度隔离的轻量级 Python 评测沙盒，并编写了与 Phase 1 数据库无缝对接的异步调度 Worker。

## 2. 独创的架构与安全设计

### 2.1 基于 Doctest 的运行时注入抽象
放弃了传统的 `.in / .out` 文件挂载方式，我们开发了 `sandbox.BuildExecutableCode` 单例注入器。
- 它将学生提交的代码与服务端私密的隐藏测试用例进行动态组装，并依赖 Python 内置的 `doctest` 进行静默的单元测试级别评估。
- 精准解析了 Python 带有 `Trying:` / `Failed example:` 的报错流，并将其提取返回为 `FailedAtCase` 这个整型指标。该设计彻底截断了底层长篇幅的报错栈输出，符合了 **“绝对隔离，仅反馈失败用例序号”** 的安全要求。

### 2.2 Docker 安全极值硬化 (Hardening)
在 `runner.go` 中创建沙盒时，我们制定了极尽苛刻的安全容器 `HostConfig / ContainerConfig` 配置，将隔离等级拉满以防止一切云端渗透可能：
1. **完全断网**: `NetworkDisabled: true`，禁止发包和反弹 Shell。测试用例 `import urllib` 直接被断头拦截 (`urllib.error.URLError`)。
2. **防提权逃跑**: `CapDrop: ["ALL"]` 移除一切额外特权；并且指定容器进程以 `User: "nobody"` 极低权限用户运行。
3. **彻底防写盘**: `ReadonlyRootfs: true` 锁定整个文件系统；通过 `Binds` 将用户生成的代码以只读 (`ro`) 挂载。
4. **资源封锁防炸弹**: 将运行时 `Memory` 硬约束在用户指定的峰值，并将 `PidsLimit` 置为 `20` 封锁无限进程 Fork Bomb，结合 Go 原生的超时协程 Context 主动 `SIGKILL` 猎杀无限超时死循环。

## 3. Worker 队列与防锁优化
编写了 `scheduler/worker.go` 模拟生产环境的 Job Poller。
为了应对突发规模提交（如全班打卡），Worker 使用了携程并发。利用 GORM 的 `WHERE id = ? AND status = PENDING` 乐观排他锁，保证无论启动多少个 Gorptide，绝不会有两个裁判长同时领取并执行同一份提交代码。

## 4. 产出物清单
1. 依赖库升级: `go mod` 平滑兼容 Docker API v1.44+。
2. `sandbox/client.go`: `InitDockerClient()` -> 环境感知器，负责拉取轻量级基础镜像。
3. `sandbox/injector.go`: -> `doctest` 组装器引擎。
4. `sandbox/runner.go`: `RunCode()` -> Docker SDK 发动机，挂载、隔离、捕获执行并回吐 OOM/TLE 等极端指标。
5. `scheduler/worker.go`: -> 并发安全调度器。
6. `sandbox/runner_test.go` / `scheduler/worker_test.go`: 涵盖四重极限测试用例（正常 AC，死循环 TLE，答案错 WA，无网络拦截 RE）。

## 5. 展望 Phase 3
评测核心组装完毕后，我们需要开始给整个应用披上一件“防盗外衣”。Phase 3 我们将全面进军基于无密码 OTP + JWT 的身份服务（Auth Node）开发。
