# PtLPOJ 开发日志 - Phase 5: VS Code 插件开发 (Client Application)

**日期**: 2026-02-24
**阶段**: Phase 5 (VS Code Client Application)
**状态**: ✅ 已完成

## 1. 原生 VS Code 插件架构
我们采用了纯血的 TypeScript + VS Code 原生 API 体系，没有引入任何 Webview 黑盒，也没有引入 React 等前端框架渲染，因为这样能够做到：
- 极其节省系统资源，响应速度可以达到毫秒级。
- 完美融入 VS Code 自带的 Dark/Light 主题。
- 与用户本地的文件系统、编辑窗口拥有最深刻的同构。

项目通过 `package.json` 中的 `contributes` 字段接管了左侧的活动栏 (Activity Bar)，并在其中注入了我们的 `Problem Sets`。

## 2. 核心交互链路实现

### 2.1 OTP 登录鉴权流与凭据管理
在 `ptlpoj.login` 中实现了分步引导：
1. 弹出 `showInputBox` 请求用户输入授权邮箱。
2. 内部发出 Axios 请求向服务器索要一次性验证码，服务器控制台将模拟拦截发送过程并显示。
3. 再次弹出输入框索要 OTP 码。
4. 获取 JWT 后，严格通过 `context.secrets.store()` 脱水保存在用户的操作系统的 Credential Manager 里面（防止本地黑客偷窃 `.json`）。

### 2.2 TreeView 侧边栏与数据渲染
`PtLpoTreeProvider` 实现了根据后端 `/api/problems` 接口的状态数组进行树状图编排：
- 成功打通过的节点（`AC`）展现出耀眼的**绿色勾号** `testing.iconPassed`。
- 回答错误的节点（`WA`）展现出红色的 `error` 图标。
- 只有鼠标点击刷新或者重新激活时才会懒加载，避免了在编辑阶段不必要的轮询。

### 2.3 工作区注入与双屏展示
核心绝技：`ptlpoj.openProblem` 指令。
当你在侧边栏点击一道题：
1. 它会自动拉取那道题的独立 Markdown 题眼描述以及基础函数的坑位 `scaffold.py`。
2. 将这俩文件瞬间实体化下载注入到当前用户打开的工作区目录。
3. 通过原生的 `markdown.showPreviewToSide` 在右侧分屏展开试卷。
4. 调用 `workspace.openTextDocument` 在左侧打开刚才释放出的 `.py` 进行焦点唤醒。这实现了零配置的沉浸式刷题方案。

### 2.4 一键提交与 SSE 原生监听流
- 绑定了原生的 `ctrl+alt+s` 快捷键，当且仅当光标停留在 python 文件时才可以触发打包。
- 调用 `axios.post` 将代码送去服务器排队 `PENDING` 并拿到判题 ID。
- **SSE 流无缝接管**：接着使用 Node.js 最底层的 `http.get` 和 `text/event-stream` Headers 实现了跟服务器的单向长连接。每隔 500ms 服务端发来最新的战报数据流。
- 当命中 `RUNNING` 的非终态时忽略，而只要状态变更到 `AC`, `WA` 或是 `TLE` 的任一终态：直接调用 `vscode.window.showInformationMessage` 弹出气泡包裹，里面含有精确的程序执行毫秒和运行峰值内存。至此完成了从写代码到最终看结果的完美闭环。

## 3. 下一步总结
目前整个扩展的核心链路（Auth -> Fetch -> Write -> Edit -> Submit -> Polling）极其通顺。
配合我们坚如磐石的 Go 后端，在接下来的 Phase 6，我们就可以启动前后端一体的双端联调大阅兵了！
