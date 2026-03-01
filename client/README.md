# PtLPOJ 客户端 (VS Code 增强扩展)

这是专为 PtLPOJ (Ptlantern's Learning Platform Online Judge) 打造的 Visual Studio Code 原生客户端环境。

## 🌟 核心特性

- **动态 OTP 身份鉴权**: 采用无密码的安全邮箱验证流，强效保护教学资源与学生隐私。
- **题目导航与资源库**: 通过内置的侧边栏视图，直观浏览、检索与管理您的算法题库。
- **沉浸式刷题空间 (Immersive Mode)**: 告别浏览器来回切换。插件将自动为您分配双栏工作区布局：左侧渲染精美的 Markdown 题面，右侧提供 Python 代码智能补全编辑区。
- **即时评测反馈引擎**: 基于 SSE 流媒体技术推送实时的云端沙盒评测进度（支持 CodeLens 一键点击 “Submit to Sandbox” ）。
- **去中心化服务器接入**: 支持动态绑定不同的 PtLPOJ 后端评判节点集群。
- **自动化管理控制台**: 高度集成的管理员可视化后台。支持授课教师通过 Excel 粘贴等方式大批量导入学生白名单；并在安全受控的环境内调用操作系统的原生文件选择器，利用 AST (抽象语法树) 智能并发生成与上传 Python 题库。

## 🚀 快速安装与接入

1. 在 VS Code 的扩展管理器 (Extensions View) 右上角点击 `...` -> `Install from VSIX...`，选择本地编译好的安装包进行安装。
2. 确保您（或您的老师）部署的 PtLPOJ 服务端进程已经上线并可通过局域网/公网访问。
3. 如果未运行在默认的 `localhost:8080` 环境下，请打开 VS Code 设置（使用快捷键 `Ctrl+,`），搜索 `ptlpoj.serverUrl` 选项，并将其修改为目标 API 网关地址。
4. 点击侧边活动栏出现的 PtLPOJ 图标，即可开启验证登入！
