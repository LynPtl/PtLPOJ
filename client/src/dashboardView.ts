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
        return `<html><body><div style="display:flex;justify-content:center;align-items:center;height:100vh;"><h3>Loading Dashboard Statistics...</h3></div></body></html>`;
    }

    private _getHtmlForWebview(stats: any) {
        const recentRows = stats.recent_submissions?.map((s: any) => `
            <tr>
                <td>${new Date(s.CreatedAt).toLocaleString()}</td>
                <td>Problem ${s.ProblemID}</td>
                <td class="status-${s.Status.toLowerCase()}">${s.Status}</td>
                <td>${s.ExecutionTimeMs}ms</td>
            </tr>
        `).join('') || '<tr><td colspan="4">No recent submissions</td></tr>';

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
            </style>
        </head>
        <body>
            <h1>🚀 Your Learning Progress</h1>
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
                    </tr>
                </thead>
                <tbody>
                    ${recentRows}
                </tbody>
            </table>
        </body>
        </html>`;
    }
}
