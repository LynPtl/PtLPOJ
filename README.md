# PtLPOJ (Ptlantern's Learning Platform Online Judge)

[![Platform](https://img.shields.io/badge/Platform-VS%20Code-blue.svg)](https://code.visualstudio.com/)
[![Language](https://img.shields.io/badge/Language-Go%20%7C%20TypeScript-00ADD8.svg)](#)
[![Docker](https://img.shields.io/badge/Sandbox-Docker-2496ED.svg)](#)
[![zread](https://img.shields.io/badge/Ask_Zread-_.svg?style=flat&color=00b0aa&labelColor=000000&logo=data%3Aimage%2Fsvg%2Bxml%3Bbase64%2CPHN2ZyB3aWR0aD0iMTYiIGhlaWdodD0iMTYiIHZpZXdCb3g9IjAgMCAxNiAxNiIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTQuOTYxNTYgMS42MDAxSDIuMjQxNTZDMS44ODgxIDEuNjAwMSAxLjYwMTU2IDEuODg2NjQgMS42MDE1NiAyLjI0MDFWNC45NjAxQzEuNjAxNTYgNS4zMTM1NiAxLjg4ODEgNS42MDAxIDIuMjQxNTYgNS42MDAxSDQuOTYxNTZDNS4zMTUwMiA1LjYwMDEgNS42MDE1NiA1LjMxMzU2IDUuNjAxNTYgNC45NjAxVjIuMjQwMUM1LjYwMTU2IDEuODg2NjQgNS4zMTM1NiAxLjYwMDEgNC45NjE1NiAxLjYwMDFaIiBmaWxsPSIjZmZmIi8%2BCjxwYXRoIGQ9Ik00Ljk2MTU2IDEwLjM5OTlIMi4yNDE1NkMxLjg4ODEgMTAuMzk5OSAxLjYwMTU2IDEwLjY4NjQgMS42MDE1NiAxMS4wMzk5VjEzLjc1OTlDMS42MDE1NiAxNC4xMTM0IDEuODg4MSAxNC4zOTk5IDIuMjQxNTYgMTQuMzk5OUg0Ljk2MTU2QzUuMzE1MDIgMTQuMzk5OSA1LjYwMTU2IDE0LjExMzQgNS42MDE1NiAxMy43NTk5VjExLjAzOTlDNS42MDE1NiAxMC42ODY0IDUuMzE1MDIgMTAuMzk5OSA0Ljk2MTU2IDEwLjM5OTlaIiBmaWxsPSIjZmZmIi8%2BCjxwYXRoIGQ9Ik0xMy43NTg0IDEuNjAwMUgxMS4wMzg0QzEwLjY4NSAxLjYwMDEgMTAuMzk4NCAxLjg4NjY0IDEwLjM5ODQgMi4yNDAxVjQuOTYwMUMxMC4zOTg0IDUuMzEzNTYgMTAuNjg1IDUuNjAwMSAxMS4wMzg0IDUuNjAwMUgxMy43NTg0QzE0LjExMTkgNS42MDAxIDE0LjM5ODQgNS4zMTM1NiAxNC4zOTg0IDQuOTYwMVYyLjI0MDFDMTQuMzk4NCAxLjg4NjY0IDE0LjExMTkgMS42MDAxIDEzLjc1ODQgMS42MDAxWiIgZmlsbD0iI2ZmZiIvPgo8cGF0aCBkPSJNNCAxMkwxMiA0TDQgMTJaIiBmaWxsPSIjZmZmIi8%2BCjxwYXRoIGQ9Ik00IDEyTDEyIDQiIHN0cm9rZT0iI2ZmZiIgc3Ryb2tlLXdpZHRoPSIxLjUiIHN0cm9rZS1saW5lY2FwPSJyb3VuZCIvPgo8L3N2Zz4K&logoColor=ffffff)](https://zread.ai/LynPtl/PtLPOJ)

PtLPOJ (Ptlantern's Learning Platform Online Judge) 是一款轻量级、安全的在线评测系统，专为局域网内的内部 Python 培训和教学场景设计。

它采用整洁的客户端-服务器架构，摒弃了传统的网页端刷题体验，将所有操作（题目浏览、代码编写、评测反馈）深度整合进 Visual Studio Code 插件中。通过 Go 编写的高并发判题内核与基于 Docker 的沙盒技术，确保每一份不受信代码都能在隔离、受限的环境中安全、快速地得到验证。

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
│   ├── src/
│   │   ├── extension.ts    # 插件入口，处理指令、登录与提交
│   │   └── treeProvider.ts # 定义侧边栏题目树状视图
│   ├── resources/          # 插件图标及静态资源
│   ├── package.json        # 插件清单与 VS Code 贡献点配置
│   └── tsconfig.json       # TypeScript 编译配置
├── server/                 # Go 评测服务端源码
│   ├── api/                # RESTful API 路由与 Handler 逻辑
│   ├── auth/               # OTP 生成及 JWT 校验模块
│   ├── middleware/         # 限流 (Rate Limit) 与身份鉴权中间件
│   ├── models/             # GORM 数据库实体模型
│   ├── sandbox/            # Docker 容器编排与安全性硬化
│   ├── scheduler/          # 多协程判题任务调度中心
│   ├── storage/            # 仓库层 (Repository) 与 SQLite 初始化
│   └── main.go             # 服务端入口程序
├── data/                   # [运行期数据] 
│   ├── ptlpoj_dev.db       # SQLite 数据库文件 (WAL 模式)
│   └── problems/           # 题目元数据与隐藏测试用例 (tests.txt)
├── docs/                   # 项目文档
│   ├── dev/                # 内部开发日志与历史 Phase 协议
│   ├── architecture_design.md
│   ├── deployment_guide.md
│   └── user_manual.md
├── tools/                  # 辅助工具脚本 (题目导入、配置生成等)
└── README.md
```

---

## 📚 文档指南 (Documentation)

为了方便后续的接手与二次开发，针对不同角色备有以下专有文档：

### 👉 对于想要了解系统设计的开发者
*   [系统架构白皮书 (Architecture Design)](docs/architecture_design.md) - 详细讲解多层组件图、鉴权流与沙盒安全限制原理。

### 👉 对于运维、教师及普通使用者
*   [全量部署与快速使用指南 (Full Guide)](docs/PtLPOJ_Full_Guide_ZH.md) - **推荐阅读**。涵盖后端 Docker 部署、插件 VSIX 安装以及如何通过 VS Code 设置连接自定义服务器。
*   [用户操作手册 (User Manual)](docs/user_manual.md) - 针对学生的纯使用流程指引。

---

## 🗺️ 演进路线规划 (Roadmap)

我们遵循透明的里程碑计划推进功能（详情请参阅项目内的独立 mission 文档）：

*   [x] **Phase 10-11**: 登录/退出 UX 闭环与 GitHub CI 自动化分发，支持自定义 Server URL。

## 🛠️ 技术栈速览

*   **后端调度与 API Gateway**: Go 1.21+, GORM, SQLite(WAL).
*   **VS Code 插件端**: TypeScript, Webview API.
*   **沙盒执行体**: `python:3.11-alpine` (Docker API SDK v43+).

---
*Created by the PtLPOJ Development Team.*
