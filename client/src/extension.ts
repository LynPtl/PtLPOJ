import * as vscode from 'vscode';
import axios from 'axios';
import * as fs from 'fs';
import * as path from 'path';
import { PtLpoTreeProvider, ProblemNode } from './treeProvider';
import { PtLpoCodeLensProvider } from './codeLensProvider';
import { ProblemViewPanel } from './problemView';
import { DashboardViewPanel } from './dashboardView';
import * as http from 'http';

const SERVER_URL = 'http://localhost:8080/api';
const TOKEN_KEY = 'ptlpoj_jwt_token';

let statusBarItem: vscode.StatusBarItem;

export function activate(context: vscode.ExtensionContext) {
    console.log('PtLPOJ Extension is now active!');

    // Status bar item to show current authentication state
    statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Left, 100);
    statusBarItem.command = 'ptlpoj.login';
    context.subscriptions.push(statusBarItem);
    updateStatusBar(context);

    // Register Sidebar Tree Provider
    const treeProvider = new PtLpoTreeProvider(context);
    vscode.window.registerTreeDataProvider('ptlpoj.problemsView', treeProvider);

    // Register CodeLens Provider for Python files
    const codeLensProvider = new PtLpoCodeLensProvider();
    vscode.languages.registerCodeLensProvider({ language: 'python', scheme: 'file' }, codeLensProvider);

    // Command: Login
    const loginCommand = vscode.commands.registerCommand('ptlpoj.login', async () => {
        await handleLogin(context);
        treeProvider.refresh();
    });

    // Command: Refresh Problems
    const refreshCommand = vscode.commands.registerCommand('ptlpoj.refreshProblems', () => {
        treeProvider.refresh();
    });

    // Command: Open Problem (From Tree View)
    const openProblemCommand = vscode.commands.registerCommand('ptlpoj.openProblem', async (node: ProblemNode) => {
        await openProblem(context, node);
    });

    // Command: Open Dashboard
    const openDashboardCommand = vscode.commands.registerCommand('ptlpoj.openDashboard', async () => {
        const token = await context.secrets.get(TOKEN_KEY);
        if (!token) {
            vscode.window.showErrorMessage('Please login first to view dashboard.');
            return;
        }
        DashboardViewPanel.createOrShow(context.extensionUri, SERVER_URL, token);
    });

    // Command: Submit Code
    const submitCommand = vscode.commands.registerCommand('ptlpoj.submitCode', async (problemIdFromWebview?: number) => {
        await submitCode(context, treeProvider, problemIdFromWebview);
    });

    // Command: Run Local Test
    const runTestCommand = vscode.commands.registerCommand('ptlpoj.runTest', async (uri?: vscode.Uri, problemIdFromWebview?: number) => {
        const terminal = vscode.window.activeTerminal || vscode.window.createTerminal('PtLPOJ Local Test');
        terminal.show();
        let targetUri = uri;

        if (!targetUri && problemIdFromWebview) {
            // Try to find the file in workspace if editor is not focused
            const workspaceFolders = vscode.workspace.workspaceFolders;
            if (workspaceFolders) {
                const pyPath = path.join(workspaceFolders[0].uri.fsPath, `Solution_${problemIdFromWebview}.py`);
                if (fs.existsSync(pyPath)) {
                    targetUri = vscode.Uri.file(pyPath);
                }
            }
        }

        if (!targetUri && vscode.window.activeTextEditor) {
            targetUri = vscode.window.activeTextEditor.document.uri;
        }

        if (targetUri) {
            await vscode.workspace.openTextDocument(targetUri).then(doc => doc.save());
            const isWin = process.platform === 'win32';
            // PowerShell fix: use semicolon or just pick one. For simplicity and win typicality:
            const cmd = isWin ? `python "${targetUri.fsPath}"` : `python3 "${targetUri.fsPath}" || python "${targetUri.fsPath}"`;
            terminal.sendText(cmd);
        } else {
            vscode.window.showErrorMessage('No file found to run test! Please open your solution file.');
        }
    });

    // Command: Set Filter and Sort
    const setFilterSortCommand = vscode.commands.registerCommand('ptlpoj.setFilterSort', async () => {
        const state = treeProvider.getFilterSortState();

        // Step 1: Pick Category
        const category = await vscode.window.showQuickPick(['Sort By', 'Filter by Status', 'Filter by Tag', 'Reset All'], {
            placeHolder: `Currently: Sort:${state.sort}, Status:${state.status}, Tag:${state.tag}`
        });

        if (!category) return;

        if (category === 'Reset All') {
            treeProvider.setFilterSort('ALL', 'ALL', 'id');
            return;
        }

        if (category === 'Sort By') {
            const pick = await vscode.window.showQuickPick(['ID (Default)', 'Status (AC first)', 'Difficulty'], { placeHolder: 'Sort by...' });
            if (!pick) return;
            const sort = pick === 'ID (Default)' ? 'id' : pick === 'Status (AC first)' ? 'status' : 'difficulty';
            treeProvider.setFilterSort(state.status, state.tag, sort as any);
        } else if (category === 'Filter by Status') {
            const pick = await vscode.window.showQuickPick(['ALL', 'AC', 'WA', 'TLE', 'RE', 'UNATTEMPTED'], { placeHolder: 'Filter by status...' });
            if (!pick) return;
            treeProvider.setFilterSort(pick, state.tag, state.sort);
        } else if (category === 'Filter by Tag') {
            const allTags = await treeProvider.getAllTags();
            const pick = await vscode.window.showQuickPick(['ALL', ...allTags], { placeHolder: 'Choose a tag...' });
            if (!pick) return;
            treeProvider.setFilterSort(state.status, pick, state.sort);
        }
    });

    context.subscriptions.push(loginCommand, refreshCommand, openProblemCommand, openDashboardCommand, submitCommand, runTestCommand, setFilterSortCommand);
}

async function handleLogin(context: vscode.ExtensionContext) {
    const email = await vscode.window.showInputBox({
        prompt: 'Enter your registered email for PtLPOJ',
        placeHolder: 'e.g., ptlantern@gmail.com'
    });

    if (!email) {
        return; // User canceled
    }

    try {
        await axios.post(`${SERVER_URL}/auth/login`, { email });
        vscode.window.showInformationMessage(`OTP requested. Check your email/console for ${email}`);

        const code = await vscode.window.showInputBox({
            prompt: `Enter the 6-digit OTP sent to ${email}`,
            placeHolder: '123456'
        });

        if (!code) {
            return;
        }

        const res = await axios.post(`${SERVER_URL}/auth/verify`, { email, code });
        const token = res.data.token;

        // Securely store token
        await context.secrets.store(TOKEN_KEY, token);
        vscode.window.showInformationMessage('Successfully logged into PtLPOJ!');
        updateStatusBar(context);

        // Auto-refresh problems after login
        vscode.commands.executeCommand('ptlpoj.refreshProblems');

    } catch (error: any) {
        vscode.window.showErrorMessage(`Login Failed: ${error.response?.data || error.message}`);
    }
}

async function openProblem(context: vscode.ExtensionContext, node: ProblemNode) {
    const token = await context.secrets.get(TOKEN_KEY);
    if (!token) {
        vscode.window.showErrorMessage('You must login first.');
        return;
    }

    // Determine a temp workspace or open in active workspace
    const workspaceFolders = vscode.workspace.workspaceFolders;
    if (!workspaceFolders || workspaceFolders.length === 0) {
        vscode.window.showErrorMessage('Please open a folder (Workspace) in VS Code first to download the problem.');
        return;
    }
    const currentFolder = workspaceFolders[0].uri.fsPath;

    try {
        const res = await axios.get(`${SERVER_URL}/problems/${node.problemId}`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });

        const data = res.data;

        // Write Markdown description
        const mdPath = path.join(currentFolder, `Problem_${node.problemId}.md`);
        fs.writeFileSync(mdPath, data.markdown);

        // Write Scaffold Python if not exists
        const pyPath = path.join(currentFolder, `Solution_${node.problemId}.py`);
        if (!fs.existsSync(pyPath)) {
            fs.writeFileSync(pyPath, data.scaffold);
        }

        // Open Python source on original view
        const pyUri = vscode.Uri.file(pyPath);
        const doc = await vscode.workspace.openTextDocument(pyUri);
        await vscode.window.showTextDocument(doc, { viewColumn: vscode.ViewColumn.One, preview: false });

        // Open Webview Problem View
        ProblemViewPanel.createOrShow(context.extensionUri, data, node.problemId);

    } catch (error: any) {
        vscode.window.showErrorMessage(`Failed to fetch problem metadata: ${error.message}`);
    }
}

async function submitCode(context: vscode.ExtensionContext, treeProvider: PtLpoTreeProvider, problemIdFromWebview?: number) {
    let editor = vscode.window.activeTextEditor;
    let code: string | undefined;
    let problemId: number | undefined = problemIdFromWebview;

    if (!editor && problemIdFromWebview) {
        // Fallback: try to find the solution file in workspace if Webview is focused
        const workspaceFolders = vscode.workspace.workspaceFolders;
        if (workspaceFolders) {
            const pyPath = path.join(workspaceFolders[0].uri.fsPath, `Solution_${problemIdFromWebview}.py`);
            if (fs.existsSync(pyPath)) {
                const doc = await vscode.workspace.openTextDocument(vscode.Uri.file(pyPath));
                code = doc.getText();
                await doc.save();
            }
        }
    } else if (editor) {
        if (editor.document.languageId !== 'python') {
            vscode.window.showErrorMessage('Only Python code can be submitted!');
            return;
        }
        code = editor.document.getText();
        const fileName = path.basename(editor.document.uri.fsPath);
        const match = fileName.match(/Solution_(\d+)\.py/);
        if (match && !problemId) {
            problemId = parseInt(match[1], 10);
        }
        await editor.document.save();
    }

    if (!code) {
        vscode.window.showErrorMessage('No active editor found or could not locate the solution file!');
        return;
    }

    if (!problemId) {
        const idStr = await vscode.window.showInputBox({
            prompt: 'Enter Problem ID to submit against (e.g. 1001)',
            placeHolder: '1001'
        });
        if (idStr) problemId = parseInt(idStr, 10);
    }

    if (!problemId) return;

    const token = await context.secrets.get(TOKEN_KEY);
    if (!token) {
        vscode.window.showErrorMessage('You must login first.');
        return;
    }

    try {
        const res = await axios.post(`${SERVER_URL}/submissions`,
            { problem_id: problemId, source_code: code },
            { headers: { 'Authorization': `Bearer ${token}` } }
        );

        const submissionId = res.data.submission_id;
        vscode.window.showInformationMessage(`Submission sent! ⏳ Judging...`);
        treeProvider.refresh(); // Status goes to Pending

        // Begin SSE Polling manually using Node's http driver to avoid heavy event-source libs
        monitorSubmissionSSE(submissionId, token, treeProvider);

    } catch (error: any) {
        vscode.window.showErrorMessage(`Submission Failed: ${error.response?.data || error.message}`);
    }
}

function monitorSubmissionSSE(subId: string, token: string, treeProvider: PtLpoTreeProvider) {
    const options = {
        hostname: 'localhost',
        port: 8080,
        path: `/api/submissions/${subId}/stream`,
        headers: {
            'Authorization': `Bearer ${token}`,
            'Accept': 'text/event-stream'
        }
    };

    http.get(options, (res) => {
        let buffer = '';
        res.on('data', (chunk) => {
            buffer += chunk.toString();
            // Parse Server-Sent Events lines
            const lines = buffer.split('\n');
            buffer = lines.pop() || '';

            for (const line of lines) {
                if (line.startsWith('data: ')) {
                    const dataStr = line.slice(6);
                    try {
                        const parsed = JSON.parse(dataStr);
                        const isFinished = parsed.finished || (parsed.Status && parsed.Status !== 'PENDING' && parsed.Status !== 'RUNNING');

                        if (ProblemViewPanel.currentPanel) {
                            let percent = 10;
                            let text = 'Waiting...';
                            if (parsed.Status === 'PENDING') { percent = 20; text = 'Queued in server...'; }
                            else if (parsed.Status === 'RUNNING') { percent = 60; text = 'Running in sandbox...'; }

                            if (isFinished) {
                                percent = 100;
                                text = `Finished: ${parsed.Status || 'Done'}. Time: ${parsed.ExecutionTimeMs || 0}ms, Memory: ${parsed.MemoryPeakKb || 0}KB`;
                                ProblemViewPanel.currentPanel.postMessage({
                                    type: 'progress',
                                    state: 'finished',
                                    percent,
                                    text,
                                    success: parsed.Status === 'AC'
                                });
                            } else {
                                ProblemViewPanel.currentPanel.postMessage({
                                    type: 'progress',
                                    state: 'running',
                                    percent,
                                    text
                                });
                            }
                        }

                        if (isFinished) {
                            // Hit terminal state!
                            const msg = `Judge Finished: ${parsed.Status || 'Done'}. Time: ${parsed.ExecutionTimeMs}ms. Memory: ${parsed.MemoryPeakKb}KB`;
                            vscode.window.showInformationMessage(msg, { modal: true });
                            treeProvider.refresh(); // Update the sidebar icon
                            res.destroy(); // End SSE stream gracefully
                        }
                    } catch (e) {
                        // ignore malformed chunks (e.g. just raw messages)
                    }
                }
            }
        });
    }).on('error', (err) => {
        vscode.window.showErrorMessage(`Connection to Server lost: ${err.message}`);
    });
}

async function updateStatusBar(context: vscode.ExtensionContext) {
    const token = await context.secrets.get(TOKEN_KEY);
    if (token) {
        statusBarItem.text = `$(check) PtLPOJ: Logged In`;
        statusBarItem.tooltip = 'Click to re-login';
    } else {
        statusBarItem.text = `$(x) PtLPOJ: Offline`;
        statusBarItem.tooltip = 'Click to log in via OTP';
    }
    statusBarItem.show();
}

export function deactivate() { }
