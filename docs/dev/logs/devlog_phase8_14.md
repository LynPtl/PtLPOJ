# Phase 8 - 14 开发日志：体系完备化与管理中台演进 (v0.2.0 - v0.3.1)

## 1. 阶段概述
自 Phase 7 完成核心题库渲染与统计闭环后，从 v0.2.0 起的开发重心转向了**系统可用性、分发自动化**及**后台架构解耦**。这使得 PtLPOJ 从一个基于本地原型的实验性项目，演进为支持大规模教学分发与非侵入式管理的成熟体系。

## 2. 核心架构演变

### 2.1 登录导流闭环 (Account Onboarding) (v0.2.0 - v0.2.1)
- **挑战**：原版登录属于隐藏极深的后门命令，不利于新人触达。
- **改建**：重组侧边栏 TreeView，在未命中 JWT 状态时动态渲染首行 `Sign-In` 指引节点；设计极简的 `LoginView` 面板接管 OTP 流，并实现了**登入即跳 Dashboard** 的转化漏斗。
- **闭环**：在状态栏注入了鉴权信息与下方的“登出断接” (Logout) 指令。

### 2.2 构建自动化流水线 (CI/CD) (v0.2.0)
- **痛点**：跨架构的 `vsce package` 人工耗时且依赖环境节点版本。
- **落地**：在 `.github/workflows/` 下接管部署逻辑，实装 Node.js 20+ 构建镜像，完成了 `client` 静态预演验证。

### 2.3 脱离单实例部署限制：配置解耦 (Config UX) (v0.2.2 - v0.2.3)
- **改进**：在插件与客户端间实装了全域配置热更新体系 `vscode.workspace.getConfiguration('ptlpoj')`。
- **交互**：为 Dashboard 与 Login 页面增加显式的 `serverUrl` 修改齿轮。使得系统具备了接入中心化公有判题服务器的能力。

### 2.4 管理与控制系统重构 (Admin Bulk Engine) (v0.3.0 - v0.3.1)
这段开发构成了本周期最庞大的代码工程，分治于前后两端：

**后端基建：**
- **轻量级鉴权**：规避复杂的 RBAC，设计基于 `OS.Getenv("PTLPOJ_ADMIN_TOKEN")` 并带 `ptlpoj_default_admin` 自动缺省降级的一级管理屏障。
- **自动化入库 (AST 解析)**：接入 `parse_worker.py`，取代了以往需要维护人员手动编写 Markdown 题面和 `tests.txt` 用例的传统方式。现在仅需一个包含 `def func` + `doctest` 的标准 Python 原文件，系统即可借助 AST (抽象语法树) 将其实时解析为脚手架代码、Markdown 题面与隐藏测试用例 (Hidden Tests) 并入库。

**前端重构：**
- **通信机制优化**：由于 Webview 沙箱环境对 HTML `<input type="file">` 获取绝对路径存在限制，本方案通过 `vscode.postMessage` 信号机制，由扩展宿主调用原生 API `vscode.window.showOpenDialog({canSelectMany: true})` 挂载原生批量文件选择器。
- **高吞吐量与容错设计**：白名单输入器引入了正则表达式 `[,;\n ]+` 进行分词，支持通过 Excel 整列复制粘贴进行批量下发，并有效过滤无效数据。多文件上传流程中引入了分批轮询及进度条（withProgress），并提供完善的失败拦截与汇总报告。

## 3. 技术难点与化解

| 异常现象 | 底层原因分析 | 终极解决方案 |
| :--- | :--- | :--- |
| **登录 Dashboard 跳动刷新报错 401** | VS Code Webview 沙箱异步机制下由于 JWT 提取落后于 React 请求时序 | 封装强屏障等待逻辑或将 Server Token 从 Cookie 隔离转入 Secret Keychain 读取 |
| **Webview 按钮上传无反应** | 现代浏览（WebView/Electron）处于安全原因将 file object 中原本的 `path`/`webkitRelativePath` 混淆为空字符串或 Fakepath | 后端不收 ArrayBuffer，前端全面切入 `VS Code Native Dialog` |

## 4. 交付总结
此里程碑为 PtLPOJ 提供了完善的行政管理支持，将出题与批量添加用户的流程转化为标准化的可视化操作。下一步系统的开发重点将回归教学与竞技本身（例如排行榜、组队匹配功能）。
