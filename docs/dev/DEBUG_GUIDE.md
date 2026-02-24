# PtLPOJ 插件安装与调试指南 (Manual Testing Guide)

本指南将引导您在本地安装并测试刚刚构建完成的 `ptlpoj-client-0.1.0.vsix` 插件包。

## 1. 准备工作
在测试插件之前，请确保您的后端服务器正在运行：
```bash
cd /home/pt/PtLPOJ/server
go run main.go
```
*后端服务器默认运行在 `http://localhost:8080`*

## 2. 安装方法 (VSIX 侧载)
1. 打开您的 **VS Code**。
2. 进入 **扩展 (Extensions)** 视图（快捷键 `Ctrl+Shift+X`）。
3. 点击右上角的 **“...”**（视图和更多操作）菜单。
4. 选择 **从 VSIX 安装... (Install from VSIX...)**。
5. 在文件选择器中导航到：`/home/pt/PtLPOJ/client/ptlpoj-client-0.1.0.vsix`。
6. 安装完成后，您可能需要点击“重新加载 (Reload)”或重启 VS Code。

## 3. 测试功能流程

### Step 3.1: 启动与登录
- 在 VS Code 左侧活动栏，您会看到一个名为 **PtLPOJ** 的图标（一个类似于叠放层的图标）。点击它。
- 点击下方的 **PtLPOJ: Login via OTP**。
- 按提示输入邮箱（例如 `ptlantern@gmail.com`），然后输入在后端控制台打印出的 6 位 OTP 验证码。
- 登录成功后，左下角状态栏会显示 `$(check) PtLPOJ: Logged In`。

### Step 3.2: 浏览与同步题目
- 在左侧面板的 **Problem Sets** 视图中，点击右上角的 **刷新** 图标。
- 您应该能看到题库列表（如 1001, 1002 等）。

### Step 3.3: 沉浸式刷题
- **点击** 列表中的某道题（例如 1001）。
- **预期结果**：插件会自动下载该题的 Markdown 文档和 Python 模板到您的当前文件夹，并自动开启左右分屏。

### Step 3.4: 提交与 SSE 监听
- 在生成的 `.py` 文件中编写代码。
- 按下快捷键 `Ctrl + Alt + S` (Mac 为 `Cmd + Alt + S`) 发起提交。
- **预期结果**：右下角弹出“Judging...”气泡。几秒钟后，您将收到最终判题结果的弹窗通知。

## 4. 调试与排错 (Debug Tips)
如果您遇到问题，可以使用以下方法查看底层日志：
1. **开发者工具**: 点击 VS Code 菜单：`帮助 (Help)` -> `切换开发人员工具 (Toggle Developer Tools)`。在 **Console** 标签页可以看到插件的所有日志和 Axios 请求报错。
2. **输出面板**: 在 VS Code 底部面板切换到 **输出 (Output)** 标签页，在下拉菜单中选择 **Log (Extension Host)**。
