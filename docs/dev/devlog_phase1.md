# PtLPOJ 开发日志 - Phase 1: 数据层与核心架构设计

**日期**: 2026-02-23
**阶段**: Phase 1 (Data Layer)
**状态**: ✅ 已完成

## 1. 阶段目标回顾
本阶段核心目标是为“局域网轻量级 Python 专属 OJ”搭建第一块基石：**底层数据模型与静态存储层**。包括解析题目数据源、初始化 Go 后端、构建 SQLite 关系型存储以及文件系统访问逻辑。

## 2. 关键架构选型与决策

### 2.1 零侵入的 “智能题目解析器”
- **背景**: 老师们提供的 `sample_questions` 全都是自带 `doctest`（`>>> f(1)`）和包含 `#sample answer` 答案的完整 `.py` 文件。
- **挑战**: 如果要将这些文件直接暴露给学生，会泄露测试用例和答案。如果要手动改写成标注的 OJ 数据，工作量极大。
- **决策**: 我们开发了 `tools/parse_questions.py`。该脚本**通过 AST 解析**源文件，自动抽离出答案源码，保留顶部的函数签名（`Signature`）作为脚手架 (`Scaffold`) 发给前端。
- **数据脱敏创新 (Fractional Split)**: 针对隐藏用例的诉求，脚本会**动态按比例（约 60%）截取**前几个用例作为 Public Tests 下发给前后端，把后几个大压力/边界用例作为 Hidden Tests 隔离在服务端沙盒。这样完美解决了“光看参数不懂题意，看全参数又容易被枚举作弊”的教学痛点。

### 2.2 Go 后端与高并发 SQLite
- **背景**: 需要一个极低部署成本，但又能在学生集中交卷时扛住一定并发的关系型数据库。
- **决策**: 选用 SQLite + GORM 取代 MySQL。
- **防死锁优化**: 在连接字符串强制开启了 WAL (Write-Ahead Logging) 模式 (`?_journal_mode=WAL&_busy_timeout=5000`)。由于 Docker 异步判卷 (Phase 2) 完成时会有很多协程同时向 DB 写入判卷结果（AC/WA），开启 WAL 可以避免致命的 `Database is locked` 错误。

### 2.3 严谨的 Submission 性能与防泄漏指标
在对齐最新的生命周期 (`lifecircle.md`) 需求时，我们对核心表 `Submission` 做了一次紧急且关键的扩充：
- **新增了 3 个关键考核字段**: `ExecutionTimeMs` (执行耗时), `MemoryPeakKb` (内存峰值), `FailedAtCase` (出错的测试点编号)。
- **安全防线**: 无论学生代码因为什么报错（WA / RE），我们的 `Message` 和展示逻辑**绝对不包含**失败用例的具体输入输出。仅通过 `FailedAtCase = N` 来提示学生是在第几个隐藏测试点倒下的，从根源上斩断面向用例编程 (Hardcoding)。

## 3. 产出物清单
1. **工具脚本**: `tools/parse_questions.py` -> 可重复执行的 CI/CD 题目生成器。
2. **静态题库**: `/data/problems/*` -> 生成出的 `problems.json`、`scaffold.py`、`problem.md` 及物理隔离的 `tests.txt`。
3. **ORM 模型**: `server/models/user.go`, `submission.go`, `problem.go`。
4. **数据访问层**: `server/storage/db.go`, `repository.go`, `problem_repo.go`。
5. **单元测试**: `server/storage/db_test.go`, `problem_repo_test.go` -> 已顺利通过。

## 4. 下一步计划 (Phase 2 Preview)
阶段 1 的坚实基础为接下来最硬核的 **阶段 2: 沙盒评测引擎开发 (Sandbox Engine)** 铺平了道路。
接下来我们将编写 Go 协程（Worker），接入 Docker Engine API，动态拉起带 `cgroups` 限制和无网络环境的极轻量 `alpine-python3` 容器，静默执行带隐式 `doctest` 注入的安全评测。
