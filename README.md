# PtLPOJ (Python Training & Learning Platform Online Judge)

[![Platform](https://img.shields.io/badge/Platform-VS%20Code-blue.svg)](https://code.visualstudio.com/)
[![Language](https://img.shields.io/badge/Language-Go%20%7C%20TypeScript-00ADD8.svg)](#)
[![Docker](https://img.shields.io/badge/Sandbox-Docker-2496ED.svg)](#)

PtLPOJ (Python Training & Learning Platform Online Judge) 是一款轻量级、安全的在线评测系统，专为局域网内的内部 Python 培训和教学场景设计。

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
