# PtLPOJ (Ptlantern's Learning Platform Online Judge)

[![Platform](https://img.shields.io/badge/Platform-VS%20Code-blue.svg)](https://code.visualstudio.com/)
[![Language](https://img.shields.io/badge/Language-Go%20%7C%20TypeScript-00ADD8.svg)](#)
[![Docker](https://img.shields.io/badge/Sandbox-Docker-2496ED.svg)](#)
[![zread](https://img.shields.io/badge/Ask_Zread-_.svg?style=flat&color=00b0aa&labelColor=000000&logo=data%3Aimage%2Fsvg%2Bxml%3Bbase64%2CPHN2ZyB3aWR0aD0iMTYiIGhlaWdodD0iMTYiIHZpZXdCb3g9IjAgMCAxNiAxNiIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTQuOTYxNTYgMS42MDAxSDIuMjQxNTZDMS44ODgxIDEuNjAwMSAxLjYwMTU2IDEuODg2NjQgMS42MDE1NiAyLjI0MDFWNC45NjAxQzEuNjAxNTYgNS4zMTM1NiAxLjg4ODEgNS42MDAxIDIuMjQxNTYgNS42MDAxSDQuOTYxNTZDNS4zMTUwMiA1LjYwMDEgNS42MDE1NiA1LjMxMzU2IDUuNjAxNTYgNC45NjAxVjIuMjQwMUM1LjYwMTU2IDEuODg2NjQgNS4zMTM1NiAxLjYwMDEgNC45NjE1NiAxLjYwMDFaIiBmaWxsPSIjZmZmIi8%2BCjxwYXRoIGQ9Ik00Ljk2MTU2IDEwLjM5OTlIMi4yNDE1NkMxLjg4ODEgMTAuMzk5OSAxLjYwMTU2IDEwLjY4NjQgMS42MDE1NiAxMS4wMzk5VjEzLjc1OTlDMS42MDE1NiAxNC4xMTM0IDEuODg4MSAxNC4zOTk5IDIuMjQxNTYgMTQuMzk5OUg0Ljk2MTU2QzUuMzE1MDIgMTQuMzk5OSA1LjYwMTU2IDE0LjExMzQgNS42MDE1NiAxMy43NTk5VjExLjAzOTlDNS42MDE1NiAxMC42ODY0IDUuMzE1MDIgMTAuMzk5OSA0Ljk2MTU2IDEwLjM5OTlaIiBmaWxsPSIjZmZmIi8%2BCjxwYXRoIGQ9Ik0xMy43NTg0IDEuNjAwMUgxMS4wMzg0QzEwLjY4NSAxLjYwMDEgMTAuMzk4NCAxLjg4NjY0IDEwLjM5ODQgMi4yNDAxVjQuOTYwMUMxMC4zOTg0IDUuMzEzNTYgMTAuNjg1IDUuNjAwMSAxMS4wMzg0IDUuNjAwMUgxMy43NTg0QzE0LjExMTkgNS42MDAxIDE0LjM5ODQgNS4zMTM1NiAxNC4zOTg0IDQuOTYwMVYyLjI0MDFDMTQuMzk4NCAxLjg4NjY0IDE0LjExMTkgMS42MDAxIDEzLjc1ODQgMS42MDAxWiIgZmlsbD0iI2ZmZiIvPgo8cGF0aCBkPSJNNCAxMkwxMiA0TDQgMTJaIiBmaWxsPSIjZmZmIi8%2BCjxwYXRoIGQ9Ik00IDEyTDEyIDQiIHN0cm9rZT0iI2ZmZiIgc3Ryb2tlLXdpZHRoPSIxLjUiIHN0cm9rZS1saW5lY2FwPSJyb3VuZCIvPgo8L3N2Zz4K&logoColor=ffffff)](https://zread.ai/LynPtl/PtLPOJ)

PtLPOJ (Ptlantern's Learning Platform Online Judge) 是一款轻量级、安全的在线评测系统，专为局域网内的内部 Python 培训和教学场景设计。

它采用整洁的客户端-服务器架构，摒弃了传统的网页端刷题体验，将所有操作（题目浏览、代码编写、评测反馈）深度整合进 Visual Studio Code 插件中。通过 Go 编写的高并发判题内核与基于 Docker 的沙盒技术，确保每一份不受信代码都能在隔离、受限的环境中安全、快速地得到验证。

---

*   **沉浸式体验**: 直接在 VS Code 编辑器中阅题、编写、运行与提交，告别网页端 IDE 切换的体验割裂问题。
*   **安全强化的沙盒环境**: 基于 Docker Engine API 编排极简容器，结合 cgroups (限制 CPU/RAM 资源)、网络阻断、User Namespace 权限降级等多重安全防护策略，有效防范恶意代码攻击与内核越权。
*   **自动化题目解析**: 基于 Python AST（抽象语法树）实现的原生代码全自动拆解流程。系统通过对单一 Python 源码文件进行解析，自动剥离生成代码脚手架、题目 Markdown 描述与隐藏测试用例记录，降低题目维护成本。
*   **可视化管理控制台**: 在 VS Code 内置原生管理面板，提供带有进度反馈的防并发用户白名单导入机制，及具备批量并发上传能力的题目管理引擎。
*   **高并发架构支撑**: 基于 Go 语言的协程模型并发调度沙盒生命周期，结合 SQLite 的 WAL (Write-Ahead Logging) 模式优化写入锁机制，确保在密集提交突发情况下的系统稳定性。
*   **无密码身份鉴权**: 采用面向团队的邮箱 OTP 验证机制与动态 JWT 会话鉴权体系，确保登录状态的长期安全性与权限隔离。

---

## 📂 目录结构 (Directory Structure)

```text
PtLPOJ/
├── client/                 # VS Code 插件源码 (TypeScript)
│   ├── src/
│   │   ├── extension.ts    # 插件入口，提供原生 FilePicker 等拦截挂载
│   │   ├── adminView.ts    # 原生批量题库管理与白名单视窗
│   │   └── treeProvider.ts # 定义侧边栏题目树状导航视图
├── server/                 # Go 评测与管理中台源码
│   ├── api/                # RESTful API (含 /api/admin/* 控制权分离)
│   ├── sandbox/            # Docker 容器编排与安全拦截隔离网
│   ├── scheduler/          # 多协程判题任务并发调度中心
│   └── main.go             # 服务端入口程序
├── tools/                  # 题库智能解析引擎 (包含 AST parse_worker.py)
├── docs/                   # 详实的部署指南与核心架构协议
│   ├── dev/                # 内部开发日志与历史阶段任务清单
│   ├── architecture_design.md
│   ├── PtLPOJ_Full_Guide_ZH.md 
│   └── user_manual.md
```

---

## 📚 文档指南 (Documentation)

为了方便后续的接手与二次开发，针对不同角色备有以下专有文档：

### 👉 对于想要了解系统设计的开发者
*   [系统架构白皮书 (Architecture Design)](docs/architecture_design.md) - 详细讲解多层组件图、鉴权流与沙盒安全限制原理。
*   [中台与插件交互协议 (Admin Control Spec)](docs/dev/missions/mission_admin.md) - 深入解析批量化插件与解析器如何交互。

### 👉 对于运维、教师及普通使用者
*   [全量部署与快速使用指南 (Full Guide)](docs/PtLPOJ_Full_Guide_ZH.md) - **推荐阅读**。涵盖后端部署、配置修改、后台系统以及白名单如何加。
*   [用户操作手册 (User Manual)](docs/user_manual.md) - 针对纯测试、纯刷题学生的使用流程。

---

## 🗺️ 演进路线规划 (Roadmap)

我们遵循透明的里程碑计划推进功能：

*   **Phase 10-12**: 身份鉴权与会话生命周期闭环、引入 GitHub CI 自动化发布工作流，支持动态扩展的自定义 Server URL 配置解耦。(v0.2.x)
*   [x] **Phase 13-14**: 可视化管理后台 (Admin Control Panel)。突破单纯的 Webview 沙箱限制，接入操作系统级文件选取交互架构与正则分词并发执行库，支持白名单用户管理与海量 `.py` 源文件的秒级录入机制。(v0.3.x)
*   [x] **Phase 15**: 队列架构重构与健康度优化。Go Channel 替代 SQL 轮询、Problem 缓存 mtime 自动 reload、Rate Limiter LRU、Graceful Shutdown、Worker Panic Recovery、TLE 判断改用 CPU 时间等改进。(v0.3.2)
*   [ ] **Phase 16+**: 社交与排行榜体系探索...

## 🛠️ 技术栈速览

*   **中间件层**: Go 1.25+, GORM, SQLite (WAL).
*   **视图层 UI**: VS Code Extension API, TypeScript, WebView Native Bridge.
*   **编译挂载体**: Python AST Compiler Engine, Docker API SDK v43+.

---
*Created by the PtLPOJ Development Team.*
