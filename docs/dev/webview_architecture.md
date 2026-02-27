# PtLPOJ Webview Architecture & UX Implementation

This document details the technical implementation of the rich UI components in the PtLPOJ VS Code extension, focusing on Webview-based views like ProblemView and DashboardView.

## 1. Overview
We use the VS Code Webview API to provide a more dynamic and visually appealing experience than standard Markdown previews.

### Key Components:
- **ProblemViewPanel**: Renders problem descriptions with LaTeX and provides action buttons.
- **DashboardViewPanel**: Displays user statistics and recent activity using CSS Grid and Flexbox.

## 2. Rendering Engine
We use `markdown-it` in the extension host to pre-render Markdown into HTML, which is then injected into the Webview.

- **LaTeX Support**: Integrated via `markdown-it-katex`. CSS for KaTeX is loaded from `node_modules` via `asWebviewUri`.
- **Syntax Highlighting**: Handled by the Markdown engine for standard code blocks.

## 3. Communication Bridge (Message Passing)
Webview and Extension communicate via JSON messages.

### From Webview to Extension:
- `submitCode`: Triggers the submission process for the currently displayed problem.
- `runLocalTest`: Triggers local execution behavior.
- `diffSubmission`: Sends the submission ID and historical code to open a `vscode.diff` editor.

### From Extension to Webview:
- `progress`: SSE updates are formatted as progress messages (percent, text, state) to update the Webview's sticky progress bar.

## 4. Theme Adaptation
All Webview CSS uses VS Code's global CSS variables to ensure seamless transition between themes.

- **Background**: `var(--vscode-editor-background)`
- **Foreground**: `var(--vscode-editor-foreground)`
- **Accent**: `var(--vscode-textLink-foreground)`
- **Button**: `var(--vscode-button-background)`

## 5. Sticky Elements & Micro-animations
- **Sticky Progress Bar**: Fixed at the top of the Problem View during submission to keep feedback visible while scrolling.
- **Card Hover Effects**: Smooth scale and transform transitions in the Dashboard for a premium feel.
