# Project Specification: PtLPOJ 登录与新人导引体验优化 (Login & Onboarding UX)

## 1. 项目背景 (Context)
虽然现有的邮箱 OTP 登录机制在安全性上达标，但在**易用性**与**用户转化**上存在明显短板：
- **发现感低**：新用户安装后，侧边栏为空且无操作指引。
- **操作链路碎**：需通过 `Ctrl+Shift+P` 搜寻命令，获取 OTP 需频繁切换窗口，导致输入焦点丢失。
- **反馈缺失**：登录成功后缺乏明确的下一步指引，导致用户迷茫。

## 2. 核心目标 (Objectives)
通过“所见即所得”的引导设计，将登录流程整合进一个聚焦的 Webview 中，并在登录成功后实现关键页面的自动跳转。

## 3. 主要功能需求 (Requirements)

### 3.1 侧边栏“零配置”登录引导 (Empty State Guide)
- **痛点**：安装后侧边栏空空如也。
- **方案**：
  - 当检测到未登录状态时，题库侧边栏应显示一条醒目的引导节点：“$(sign-in) 点击此处登录 PtLPOJ”。
  - 点击该节点直接触发登录流程。

### 3.2 聚焦式 Webview 登录窗 (Focused Login Webview)
- **痛点**：原生 InputBox 交互受限，切换窗口易丢失焦点。
- **方案**：
  - 开发一个极简但美观的 `LoginView` Webview。
  - **分步式交互**：
    1. 用户输入邮箱 -> 点击“获取验证码” -> 页面平滑过渡到验证码输入。
    2. 用户在 Webview 内输入验证码，并在输入完成后（或点击提交）自动执行验证。
  - **窗口置顶性**：Webview 面板能更好地维持用户的视觉专注。

### 3.3 登录成功自动转跳 (Auto-Onboarding Flow)
- **方案**：
  - 验证成功后，插件应**自动关闭**登录窗并**自动打开** `Dashboard` (仪表盘) 页面。
  - Dashboard 顶部的“新人指导”将接力后续的任务指引（如何选题、如何提交）。

## 4. 技术计划 (Execution Plan)

- [ ] **Phase 8.1**: 维护 `treeProvider` 的未登录占位节点。
- [ ] **Phase 8.2**: 开发 `LoginView` 核心逻辑与 HTML 模板（采用 HSL 配色与 Webview 变量）。
- [ ] **Phase 8.3**: 修改 `extension.ts` 中的命令链，实现 `Login -> Dashboard` 的自动导流。

---
*Refined based on real-world UX friction feedback.*
