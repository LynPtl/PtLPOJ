# PtLPOJ (Personal Local Python Online Judge)

[![Platform](https://img.shields.io/badge/Platform-VS%20Code-blue.svg)](https://code.visualstudio.com/)
[![Language](https://img.shields.io/badge/Language-Go%20%7C%20TypeScript-00ADD8.svg)](#)
[![Docker](https://img.shields.io/badge/Sandbox-Docker-2496ED.svg)](#)

PtLPOJ 是一款专为内部团队、班级和小群体设计的**重后端、轻前端的沉浸式代码评测平台**。

它摒弃了传统的网页端刷题体验，将所有操作（发题、做题、评测、分析）深度整合进 Visual Studio Code 插件中。同时，通过 Go 语言编写的高并发判题内核与严格的安全沙盒机制，确保每一份不受信代码都能在极少资源消耗的情况下被安全、快速地验证。

---

## 🚀 核心特性 (Features)

*   **沉浸式体验**: 直接在 VS Code 编辑器中阅题、编写、运行与提交，告别网页端没有智能提示的痛苦。
*   **安全硬化沙盒**: 基于 Docker Engine API 的极简容器，结合 cgroups (CPU/RAM)、断网、User Namespace 降权多重防护，根绝恶意攻击与内核提权。
*   **高并发架构**: Go 协程并发拉起沙盒引擎，结合 SQLite WAL 模式优化写锁，支持强烈的提交尖峰。
*   **SSE 实时反馈**: 从排队、准备环境、分步运行到最终结果（AC/WA/TLE/RE/OLE），进度实时弹窗通知。
*   **无密码验证**: 面向团队的极简邮箱 OTP 验证 + 动态 JWT 会话鉴权。

---

## 📂 目录结构 (Directory Structure)

```text
PtLPOJ/
├── client/                 # VS Code 插件源码 (TypeScript)
│   ├── src/                # 插件核心逻辑 (TreeProvider, Commands等)
│   └── package.json        # 插件定义与贡献点
├── server/                 # 评测服务端源码 (Go)
│   ├── api/                # RESTful API 处理器与路由
│   ├── sandbox/            # Docker 沙盒编排与资源限制逻辑
│   ├── scheduler/          # 判题任务调度中心
│   ├── storage/            # SQLite 及文件系统持久化层
│   └── main.go             # 服务端入口
├── docs/                   # 项目文档
│   ├── dev/                # 开发日志与历史协议 (Internal)
│   ├── architecture_design.md
│   ├── deployment_guide.md
│   └── user_manual.md
├── data/                   # [运行期生成] 存放题库、用户数据与 SQLite 文件
└── README.md
```

---

## 📚 文档指南 (Documentation)

为了方便后续的接手与二次开发，针对不同角色备有以下专有文档：

### 👉 对于想要了解系统设计的开发者
*   [系统架构白皮书 (Architecture Design)](docs/architecture_design.md) - 详细讲解多层组件图、鉴权流与沙盒安全限制原理。

### 👉 对于想要部署服务端的运维/教师
*   [部署与运维指南 (Deployment Guide)](docs/deployment_guide.md) - 涵盖 Docker 与 Go 环境拉取、服务器运行、端口配置及 SQLite 数据备份。

### 👉 对于最终刷题的学生/使用者
*   [用户操作手册 (User Manual)](docs/user_manual.md) - 从插件绑定、OTP 登录到一键 `Ctrl+Alt+S` 提交流程的完整图文引导。

---

## 🗺️ 演进路线规划 (Roadmap)

我们遵循透明的里程碑计划推进功能（详情请参阅项目内的独立 mission 文档）：

*   [x] **Phase 1-6**: 核心骨架（鉴权、沙盒评测、数据通信、基础插件端对接）
*   [x] **Phase 9**: 系统全面文档化 (Architecture & User Manuals)
*   [ ] **Phase 7 (Next)**: UI/UX 极致体验升级 (CodeLens 左侧边距按键、动态 Dashboards) -> 参阅 `docs/dev/mission_UIUX.md`
*   [ ] **Phase 8**: 教师端 API 及动态题目分类检索拓展 -> 参阅 `docs/dev/mission_apis.md`

## 🛠️ 技术栈速览

*   **后端调度与 API Gateway**: Go 1.21+, GORM, SQLite(WAL).
*   **VS Code 插件端**: TypeScript, Webview API.
*   **沙盒执行体**: `python:3.11-alpine` (Docker API SDK v43+).

---
*Created by the PtLPOJ Development Team.*
