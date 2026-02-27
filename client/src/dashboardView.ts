import * as vscode from 'vscode';
import axios from 'axios';

export class DashboardViewPanel {
    public static currentPanel: DashboardViewPanel | undefined;
    private readonly _panel: vscode.WebviewPanel;
    private readonly _extensionUri: vscode.Uri;
    private _disposables: vscode.Disposable[] = [];

    public static createOrShow(extensionUri: vscode.Uri, serverUrl: string, token: string) {
        if (DashboardViewPanel.currentPanel) {
            DashboardViewPanel.currentPanel._panel.reveal(vscode.ViewColumn.One);
            DashboardViewPanel.currentPanel.refresh(serverUrl, token);
            return;
        }

        const panel = vscode.window.createWebviewPanel(
            'ptlpojDashboard',
            'PtLPOJ Dashboard',
            vscode.ViewColumn.One,
            {
                enableScripts: true,
                localResourceRoots: [extensionUri]
            }
        );

        DashboardViewPanel.currentPanel = new DashboardViewPanel(panel, extensionUri, serverUrl, token);
    }

    private constructor(panel: vscode.WebviewPanel, extensionUri: vscode.Uri, serverUrl: string, token: string) {
        this._panel = panel;
        this._extensionUri = extensionUri;

        this.refresh(serverUrl, token);
        this._panel.onDidDispose(() => this.dispose(), null, this._disposables);

        this._panel.webview.onDidReceiveMessage(
            message => {
                switch (message.command) {
                    case 'diffSubmission':
                        vscode.commands.executeCommand('ptlpoj.diffSubmission', message.problemId, message.code);
                        return;
                }
            },
            null,
            this._disposables
        );
    }

    public async refresh(serverUrl: string, token: string) {
        this._panel.webview.html = this._getLoadingHtml();
        try {
            const response = await axios.get(`${serverUrl}/user/stats`, {
                headers: { 'Authorization': `Bearer ${token}` }
            });
            this._panel.webview.html = this._getHtmlForWebview(response.data);
        } catch (error: any) {
            this._panel.webview.html = `<h1>Error loading dashboard</h1><p>${error.message}</p>`;
        }
    }

    public dispose() {
        DashboardViewPanel.currentPanel = undefined;
        this._panel.dispose();
        while (this._disposables.length) {
            const x = this._disposables.pop();
            if (x) x.dispose();
        }
    }

    private _getLoadingHtml() {
        return `<html><body style="background-color: var(--vscode-editor-background); color: var(--vscode-editor-foreground);"><div style="display:flex;justify-content:center;align-items:center;height:100vh;"><h3>Loading Dashboard Statistics...</h3></div></body></html>`;
    }

    private _getHtmlForWebview(stats: any) {
        const recentRows = stats.recent_submissions?.map((s: any) => `
            <tr>
                <td>${new Date(s.CreatedAt).toLocaleString()}</td>
                <td>Problem ${s.ProblemID}</td>
                <td class="status-${s.Status.toLowerCase()}">${s.Status}</td>
                <td>${s.ExecutionTimeMs}ms</td>
                <td>
                    <button class="btn-small" onclick="diffSubmission(${s.ProblemID}, ${JSON.stringify(s.Code).replace(/"/g, '&quot;')})">Diff</button>
                </td>
            </tr>
        `).join('') || '<tr><td colspan="5">No recent submissions</td></tr>';

        return `<!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="UTF-8">
            <style>
                body {
                    font-family: var(--vscode-font-family);
                    padding: 30px;
                    color: var(--vscode-editor-foreground);
                    background-color: var(--vscode-editor-background);
                }
                .banner {
                    background: linear-gradient(135deg, var(--vscode-textLink-foreground) 0%, var(--vscode-button-background) 100%);
                    color: white;
                    padding: 20px 30px;
                    border-radius: 12px;
                    margin-bottom: 30px;
                    display: flex;
                    justify-content: space-between;
                    align-items: center;
                    box-shadow: 0 4px 15px rgba(0,0,0,0.2);
                }
                .banner-text h2 { border: none; margin: 0; padding: 0; color: white;}
                .banner-text p { margin: 5px 0 0 0; opacity: 0.9; }
                .grid {
                    display: grid;
                    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
                    gap: 20px;
                    margin-bottom: 40px;
                }
                .card {
                    background-color: var(--vscode-editorWidget-background);
                    border: 1px solid var(--vscode-widget-border);
                    border-radius: 8px;
                    padding: 24px;
                    text-align: center;
                    box-shadow: 0 4px 6px rgba(0,0,0,0.1);
                    transition: transform 0.2s;
                }
                .card:hover {
                    transform: translateY(-5px);
                    border-color: var(--vscode-focusBorder);
                }
                .card-value {
                    font-size: 36px;
                    font-weight: bold;
                    display: block;
                    margin-bottom: 8px;
                    color: var(--vscode-textLink-foreground);
                }
                .card-label {
                    font-size: 14px;
                    color: var(--vscode-descriptionForeground);
                    text-transform: uppercase;
                    letter-spacing: 1px;
                }
                h2 {
                    border-bottom: 1px solid var(--vscode-widget-border);
                    padding-bottom: 10px;
                    margin-top: 40px;
                }
                table {
                    width: 100%;
                    border-collapse: collapse;
                    margin-top: 20px;
                }
                th {
                    text-align: left;
                    padding: 12px;
                    border-bottom: 2px solid var(--vscode-widget-border);
                    color: var(--vscode-descriptionForeground);
                }
                td {
                    padding: 12px;
                    border-bottom: 1px solid var(--vscode-widget-border);
                }
                .status-ac { color: var(--vscode-testing-iconPassed); font-weight: bold; }
                .status-wa, .status-tle, .status-re { color: var(--vscode-testing-iconFailed); }
                .status-pending, .status-running { color: var(--vscode-progressBar-background); }
                .btn-small {
                    background-color: var(--vscode-button-secondaryBackground);
                    color: var(--vscode-button-secondaryForeground);
                    border: none;
                    padding: 4px 8px;
                    border-radius: 3px;
                    cursor: pointer;
                    font-size: 11px;
                }
                .btn-small:hover { background-color: var(--vscode-button-secondaryHoverBackground); }
                .onboarding {
                   margin-top: 40px;
                   padding: 20px;
                   background-color: var(--vscode-editor-inactiveSelectionBackground);
                   border-radius: 8px;
                   border-left: 4px solid var(--vscode-focusBorder);
                }
            </style>
        </head>
        <body>
            <div class="banner">
                <div class="banner-text">
                    <h2>Daily Challenge: Two Sum</h2>
                    <p>Level: Easy • Solve it to earn 10 points!</p>
                </div>
                <button class="btn-small" style="padding: 10px 20px; font-size: 14px; font-weight: bold;">Go Solve</button>
            </div>

            <div class="grid">
                <div class="card">
                    <span class="card-value">${stats.total_submissions}</span>
                    <span class="card-label">Total Submissions</span>
                </div>
                <div class="card">
                    <span class="card-value">${stats.ac_count}</span>
                    <span class="card-label">Solved (AC)</span>
                </div>
                <div class="card">
                    <span class="card-value">${stats.unique_problems_solved}</span>
                    <span class="card-label">Unique Problems</span>
                </div>
                <div class="card">
                    <span class="card-value">${stats.total_submissions > 0 ? Math.round((stats.ac_count / stats.total_submissions) * 100) : 0}%</span>
                    <span class="card-label">Success Rate</span>
                </div>
            </div>

            <h2>🕗 Recent Activity</h2>
            <table>
                <thead>
                    <tr>
                        <th>Time</th>
                        <th>Problem</th>
                        <th>Status</th>
                        <th>Runtime</th>
                        <th>Action</th>
                    </tr>
                </thead>
                <tbody>
                    ${recentRows}
                </tbody>
            </table>

            <div class="onboarding">
                <h3>New here? 💡</h3>
                <p>Welcome to PtLPOJ! To get started:
                   1. Select a problem from the left sidebar.
                   2. Write your solution in the Python editor.
                   3. Use <b>CodeLens</b> (Run Test/Submit) at the top of your file for quick feedback!
                </p>
            </div>

            <script>
                const vscode = acquireVsCodeApi();
                function diffSubmission(problemId, code) {
                    vscode.postMessage({ command: 'diffSubmission', problemId, code });
                }
            </script>
        </body>
        </html>`;
    }
}
