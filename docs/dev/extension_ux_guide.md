# PtLPOJ Extension UI Customization Guide

This guide explains how to add or modify UI features in the sidebar and dashboard.

## 1. Adding a New Filter/Sorter
1. **treeProvider.ts**:
    - Add a new state variable (e.g., `filterCategory`).
    - Update `getChildren` logic to apply this filter.
    - Add a setter method that calls `this.refresh()`.
2. **extension.ts**:
    - Add a new `vscode.window.showQuickPick` option in `setFilterSortCommand`.

## 2. Adding Dashboard Cards
1. **stats_handler.go** (Backend):
    - Update the `UserStats` struct and populate the new field with SQL aggregation.
2. **dashboardView.ts** (Frontend):
    - Update the HTML template in `_getHtmlForWebview` to include a new `.card` div.

## 3. Search Implementation
The search functionality uses `vscode.window.showInputBox` to capture user input, which is then passed to `treeProvider.setSearchQuery`. The TreeView refreshes automatically via the `onDidChangeTreeData` event.

## 4. Local Testing with Doctest
The extension automatically detects the presence of `doctest` by scanning for `import doctest` or `doctest.testmod` in the source code. If found, it switches the execution command from a standard python run to `python -m doctest -v`.
