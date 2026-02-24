import * as vscode from 'vscode';
import axios from 'axios';
import * as fs from 'fs';
import * as path from 'path';
import { PtLpoTreeProvider, ProblemNode } from './treeProvider';
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

    // Command: Submit Code
    const submitCommand = vscode.commands.registerCommand('ptlpoj.submitCode', async () => {
        await submitCode(context, treeProvider);
    });

    context.subscriptions.push(loginCommand, refreshCommand, openProblemCommand, submitCommand);
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

        // Open MD preview and then Python source
        const mdUri = vscode.Uri.file(mdPath);
        await vscode.commands.executeCommand('markdown.showPreviewToSide', mdUri);

        const pyUri = vscode.Uri.file(pyPath);
        const doc = await vscode.workspace.openTextDocument(pyUri);
        await vscode.window.showTextDocument(doc, { viewColumn: vscode.ViewColumn.One });

    } catch (error: any) {
        vscode.window.showErrorMessage(`Failed to fetch problem metadata: ${error.message}`);
    }
}

async function submitCode(context: vscode.ExtensionContext, treeProvider: PtLpoTreeProvider) {
    const editor = vscode.window.activeTextEditor;
    if (!editor) {
        vscode.window.showErrorMessage('No active editor found!');
        return;
    }

    if (editor.document.languageId !== 'python') {
        vscode.window.showErrorMessage('Only Python code can be submitted!');
        return;
    }

    const code = editor.document.getText();
    const fileName = path.basename(editor.document.uri.fsPath);

    // Auto infer Problem ID from filename Solution_1001.py -> 1001
    const match = fileName.match(/Solution_(\d+)\.py/);
    let problemIdStr = match ? match[1] : undefined;

    if (!problemIdStr) {
        problemIdStr = await vscode.window.showInputBox({
            prompt: 'Enter Problem ID to submit against (e.g. 1001)',
            placeHolder: '1001'
        });
    }

    if (!problemIdStr) return;
    const problemId = parseInt(problemIdStr, 10);

    const token = await context.secrets.get(TOKEN_KEY);
    if (!token) {
        vscode.window.showErrorMessage('You must login first.');
        return;
    }

    // Save document before submitting
    await editor.document.save();

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
                        if (parsed.finished || (parsed.Status && parsed.Status !== 'PENDING' && parsed.Status !== 'RUNNING')) {
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
