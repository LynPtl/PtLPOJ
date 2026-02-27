# 指南：UI/UX 贡献与扩展说明 (UIUX_CONTRIBUTING_GUIDE)

## 1. 愿景
PtLPOJ 的 UI 设计原则是 **极致反馈、沉浸体验**。任何 UI 的变更都应以“减少用户点击次数”和“增强评测透明度”为先。

## 2. 扩展现有 TreeView
若需要为左侧题目树增加新的过滤逻辑（例如：按题目来源过滤）：

1. **State 管理**：在 `treeProvider.ts` 中新增状态变量，并实现对应的 Setter（确保注火 `_onDidChangeTreeData` 事件）。
2. **命令注册**：在 `extension.ts` 中注册对应的指令，并通过 `vscode.window.showQuickPick` 触发。
3. **图标配置**：新功能图标应从 [Material Icons](https://marella.me/material-design-icons/demo/font/) 或 VS Code 内置图标集中选取。

## 3. 增强 Dashboard 界面
Dashboard 是一个基于 Vanilla CSS 的响应式单页应用。
- **模板更新**：在 `dashboardView.ts` 的 `_getHtmlForWebview` 方法中添加 HTML 结构。
- **动效规范**：
  - 微交互：所有的卡片悬浮（Hover）均应有 `transform: translateY` 动画。
  - 颜色：背景渐变色应控制在 135deg 到 180deg 之间，使用 HSL 体系以保证色彩柔和。

## 4. 本地测试逻辑优化
目前的 `Run Local Test` 已支持 `doctest` 自动识别。扩展该逻辑时请注意：
- **无依赖性**：本地运行不应强制依赖 Python 以外的第三方评测工具，以保证开箱即用。
- **环境隔离**：确保本地执行命令时带上引号以处理包含空格的路径。

## 5. 代码贡献检查单
- [ ] 是否在所有 Webview 元素中使用了系统主题变量？
- [ ] 是否为新增的按钮添加了 `aria-label` 辅助说明？
- [ ] 关键的 UI 状态切换（如评测结束）是否有视觉动效反馈？
