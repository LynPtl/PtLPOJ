# Project Specification: PtLPOJ UI/UX Optimization

## 1. 项目背景与目标 (Project Context & Objectives)

在 Phase 1-6 中，我们已经完成了 PtLPOJ 的核心功能建设，包括沙盒评测引擎、JWT 鉴权、插件基础交互等。目前系统处于“功能可用但体验简陋”的 Demo 阶段。

本阶段（UI/UX 优化）的目标是将 PtLPOJ 打造为一个**沉浸式、高反馈、具有现代审美**的 VS Code 插件。我们将抛弃简单的原生 UI，引入 Webview 渲染技术，提供类似 LeetCode 插件但更轻量、更聚焦的内部工具体验。

## 2. 核心需求描述 (Core Requirements)

### 2.1 仪表盘中心 (Home Dashboard)
* **需求痛点**：用户登录后缺乏全局视角，不知道自己做了多少题，正确率如何。
* **主要功能**：
    * **个人概览**：展示已解决、已尝试、正确率等核心统计指标。
    * **最近活动**：列出最近的 5 次提交，并支持一键跳转。
    * **每日任务/随机题目**：通过“今日题目”横幅引导用户保持练习。
    * **视觉要求**：使用卡片布局、渐变色装饰，营造“仪表盘”的感觉。

### 2.2 富文本题目视图 (Rich Problem View)
* **需求痛点**：目前的侧边预览 Markdown + Python 文件的模式分散了注意力，且不支持复杂的题目排版。
* **主要功能**：
    * **全能排版渲染 (New)**：使用 Webview 搭配进阶 Markdown 引擎（如 `markdown-it` 加载项），**无缝支持**以下高级格式：
        * LaTeX 数学公式（集成 KaTeX 或 MathJax，支持 `$` 和 `$$` 语法）。
        * 复杂的数据表格（GFM Schema）。
        * 带语言标识的语法高亮代码块。
    * **状态联动**：题目视图应直接显示当前题目是否已通过。
    * **双重提交入口**：
        1. **编辑器 CodeLens**：在 `Solution_*.py` 代码首行上方显示“Run Test | Submit”交互链接（类似 LeetCode），消除快捷键记忆负担，提供最快捷的编码期操作。
        2. **Webview 按钮**：在题目视图底部或固定悬浮位提供显形的“Submit”和“Run Local Test”按钮，方便阅读完题目后顺滑衔接操作。
    * **本地测试集成**：允许用户在提交到沙盒前，利用本地环境快速验证 `doctest`。

### 2.3 实时评测反馈与历史 (Live Feedback & History)
* **需求痛点**：提交后“Judging...”提示过于生硬，且无法方便地回顾之前的思路。
* **主要功能**：
    * **进度可视化**：在题目视图或 Dashboard 中展示评测条。
    * **分步反馈**：利用 SSE 推送，显示“Environment Ready” -> “Running Case 1/10” -> “Analyzing Result”等状态。
    * **提交记录对比 (New)**：支持查看历史提交的代码，并提供 Diff 视图将其与当前编辑器中的代码进行对比。
    * **结果卡片**：AC 时弹出精致的庆祝弹窗或组件；WA 时提供清晰的代码 Diff 或错误定位。

### 2.4 其他全局体验增强 (Global UX Enhancements)
* **主题完美适配**：Webview CSS 需自动同步 VS Code 主题变量（`--vscode-*`），确保在深色/浅色/高对比度模式下均有极佳表现。
* **题目筛选与搜索**：在侧边栏 TreeView 顶部增加搜索框或筛选器（按难度、按状态）。
* **新手引导 (Onboarding)**：Dashboard 提供“快速开始”指南，引导新用户完成登录和首道题目练习。

## 3. 技术路线与架构 (Tech Stack & Architecture)

### 3.1 表现层 (Client Side)
* **VS Code Webview API**：用于承载复杂的 HTML/CSS 界面。
* **Vanilla CSS / Modern Typography**：遵循 Web 应用开发规范，使用丰富的配色方案（HSL）和微动作。
* **Message Passing**：实现 Webview 界面与插件核心逻辑的双向通信。

### 3.2 数据层 (Server Side)
* **统计 API**：新增 `/api/user/stats` 接口，聚合计算用户数据。
* **SSE 增强**：优化推送消息的粒度，支持发送更详细的阶段性状态。

## 4. UI 设计原则 (Design Principles)

1.  **WOW Factor**：初次打开仪表盘应让用户感到视觉上的惊艳。
2.  **响应式布局**：完美适配 VS Code 侧边抽屉或编辑器主视图的不同宽度。
3.  **品牌一致性**：使用统一的调色板（建议深紫/青色系）和图标集。
4.  **低阻力交互**：减少配置操作，所有核心动作（登录、题目切换、提交）应在一两次点击内完成。

## 5. 任务优先级与执行计划 (Execution Plan)

> **原则**：综合考虑开发难度（Cost）与用户体验收益（ROI）。优先落实与“做题”最紧密结合的核心交互，再逐步完善全局仪表盘与高级特性。

### Phase 7.1: 快速交互赢取 (Quick Wins - CodeLens)
* **目标**: 消除快捷键痛点，以极低的成本获得巨大的体验提升。
* **任务**:
  - [ ] 开发 `PtLpoCodeLensProvider`。
  - [ ] 在 `Solution_*.py` 注入 `Run Test` 和 `Submit` 入口并绑定现有命令。

### Phase 7.2: 核心题目体验重构 (Core Problem View)
* **目标**: 解决 Markdown 原生预览体验差、缺乏提交反馈的核心痛点，奠定后续 Webview 基础。
* **任务**:
  - [ ] 搭建基础 Webview 框架与消息通信机制。
  - [ ] 引入 `markdown-it` 系列增强解析器。
  - [ ] 实现 `ProblemView`，集成题目描述与底部/悬浮的独立提交按钮。
  - [ ] 将现有基于 SSE 的“右下角弹窗评测反馈”无缝迁移为 Webview 内的实时进度条展示。

### Phase 7.3: 数据流栈与控制中心 (Data & Dashboard)
* **目标**: 补安全局概览，增强产品的完成度。
* **任务**:
  - [ ] (Server) 新增聚合统计逻辑与 `/api/user/stats` 接口。
  - [ ] (Client) 开发带有绚丽 UI/CSS 动效的 Dashboard Webview。
  - [ ] (Client) 将 Dashboard 与侧边栏的入口进行联动。

### Phase 7.4: 进阶体验特性 (Advanced Polishing)
* **目标**: 打通全链路闭环，达到生产级插件的体验标准。
* **任务**:
  - [ ] (Client) 开发代码内容的历史 Diff 对比试图。
  - [ ] (Client) 侧边栏 TreeView 增加按难度搜索与状态过滤功能。
  - [ ] (Client) 确保 Webview 全局主题自动桥接 VS Code 变量表现完美。
  - [ ] (Server/Client) 实现本地快速运行 `doctest` 的无缝集成。

---
*Generated based on user requirements for UI/UX hardening.*
