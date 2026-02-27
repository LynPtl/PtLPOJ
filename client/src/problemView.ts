import * as vscode from 'vscode';

export class ProblemViewPanel {
    public static currentPanel: ProblemViewPanel | undefined;
    private readonly _panel: vscode.WebviewPanel;
    private readonly _extensionUri: vscode.Uri;
    private _problemId: number;
    private _disposables: vscode.Disposable[] = [];

    public static createOrShow(extensionUri: vscode.Uri, problemData: any, problemId: number) {
        const column = vscode.window.activeTextEditor ? vscode.window.activeTextEditor.viewColumn : undefined;

        if (ProblemViewPanel.currentPanel) {
            ProblemViewPanel.currentPanel._panel.reveal(vscode.ViewColumn.Two);
            ProblemViewPanel.currentPanel.update(problemData, problemId);
            return;
        }

        const panel = vscode.window.createWebviewPanel(
            'ptlpojProblemView',
            `Problem ${problemId}`,
            vscode.ViewColumn.Two,
            {
                enableScripts: true,
                localResourceRoots: [vscode.Uri.joinPath(extensionUri, 'node_modules')]
            }
        );

        ProblemViewPanel.currentPanel = new ProblemViewPanel(panel, extensionUri, problemData, problemId);
    }

    private constructor(panel: vscode.WebviewPanel, extensionUri: vscode.Uri, problemData: any, problemId: number) {
        this._panel = panel;
        this._extensionUri = extensionUri;
        this._problemId = problemId;

        this.update(problemData, problemId);
        this._panel.onDidDispose(() => this.dispose(), null, this._disposables);

        this._panel.webview.onDidReceiveMessage(
            message => {
                switch (message.command) {
                    case 'submitCode':
                        vscode.commands.executeCommand('ptlpoj.submitCode', this._problemId);
                        return;
                    case 'runLocalTest':
                        vscode.commands.executeCommand('ptlpoj.runTest', undefined, this._problemId);
                        return;
                }
            },
            null,
            this._disposables
        );
    }

    public update(problemData: any, problemId: number) {
        this._problemId = problemId;
        this._panel.title = `Problem ${problemId}`;
        this._panel.webview.html = this._getHtmlForWebview(this._panel.webview, problemData, problemId);
    }

    public postMessage(message: any) {
        this._panel.webview.postMessage(message);
    }

    public dispose() {
        ProblemViewPanel.currentPanel = undefined;
        this._panel.dispose();
        while (this._disposables.length) {
            const x = this._disposables.pop();
            if (x) {
                x.dispose();
            }
        }
    }

    private _getHtmlForWebview(webview: vscode.Webview, problemData: any, problemId: number) {
        const MarkdownIt = require('markdown-it');
        const markdownItKatex = require('markdown-it-katex');

        const md = new MarkdownIt({ html: true }).use(markdownItKatex);
        // The problem description payload seems to have either Title or problem details
        const titleStr = problemData.Problem?.Title || problemData.Title || `Problem ${problemId}`;
        const markdownSource = problemData.markdown || problemData.Description || 'No description provided.';
        const renderedMarkdown = md.render(markdownSource);

        const katexUri = webview.asWebviewUri(vscode.Uri.joinPath(this._extensionUri, 'node_modules', 'katex', 'dist', 'katex.min.css'));

        return `<!DOCTYPE html>
            <html lang="en">
            <head>
                <meta charset="UTF-8">
                <meta name="viewport" content="width=device-width, initial-scale=1.0">
                <title>${titleStr}</title>
                <link rel="stylesheet" href="${katexUri}">
                <style>
                    body {
                        font-family: var(--vscode-font-family);
                        padding: 0 20px 20px 20px;
                        line-height: 1.6;
                        color: var(--vscode-editor-foreground);
                        background-color: var(--vscode-editor-background);
                    }
                    /* Fixed Header for Progress */
                    #progress-container {
                        display: none;
                        position: sticky;
                        top: 0;
                        left: 0;
                        right: 0;
                        background: var(--vscode-editorWidget-background);
                        padding: 15px;
                        border-bottom: 2px solid var(--vscode-focusBorder);
                        z-index: 1001;
                        box-shadow: 0 2px 5px rgba(0,0,0,0.2);
                        margin: 0 -20px 20px -20px;
                    }
                    h1, h2, h3 { color: var(--vscode-editor-foreground); border-bottom: 1px solid var(--vscode-widget-border); padding-bottom: 5px;}
                    .action-bar {
                        position: fixed;
                        bottom: 0px;
                        left: 0;
                        right: 0;
                        padding: 10px 20px;
                        background-color: var(--vscode-editorWidget-background);
                        border-top: 1px solid var(--vscode-widget-border);
                        display: flex;
                        justify-content: flex-end;
                        gap: 10px;
                        box-shadow: 0 -2px 10px rgba(0,0,0,0.5);
                        z-index: 1000;
                    }
                    /* ... (rest of old CSS) ... */
                    .btn {
                        padding: 8px 16px;
                        border: none;
                        border-radius: 2px;
                        cursor: pointer;
                        font-weight: 500;
                        font-size: 13px;
                    }
                    .btn-primary {
                        background-color: var(--vscode-button-background);
                        color: var(--vscode-button-foreground);
                    }
                    .btn-primary:hover {
                        background-color: var(--vscode-button-hoverBackground);
                    }
                    .btn-secondary {
                        background-color: var(--vscode-button-secondaryBackground);
                        color: var(--vscode-button-secondaryForeground);
                    }
                    .btn-secondary:hover {
                        background-color: var(--vscode-button-secondaryHoverBackground);
                    }
                    .content {
                        margin-bottom: 80px; 
                    }
                    pre {
                        background-color: var(--vscode-textCodeBlock-background);
                        padding: 10px;
                        border-radius: 4px;
                        overflow-x: auto;
                    }
                    code {
                        font-family: var(--vscode-editor-font-family);
                    }
                    blockquote {
                        border-left: 4px solid var(--vscode-textLink-foreground);
                        padding: 0 15px;
                        margin-left: 0;
                        color: var(--vscode-textPreformat-foreground);
                    }
                    
                    table {
                        border-collapse: collapse;
                        width: 100%;
                        margin-bottom: 20px;
                    }
                    table, th, td {
                        border: 1px solid var(--vscode-widget-border);
                    }
                    th, td {
                        padding: 8px 12px;
                        text-align: left;
                    }
                    th {
                        background-color: var(--vscode-editorWidget-background);
                    }

                    .progress-bar-bg {
                        width: 100%;
                        height: 6px;
                        background-color: var(--vscode-scrollbarSlider-background);
                        border-radius: 3px;
                        overflow: hidden;
                        margin-bottom: 10px;
                    }
                    .progress-bar-fill {
                        height: 100%;
                        background-color: var(--vscode-button-background);
                        width: 0%;
                        transition: width 0.3s ease;
                    }
                    #progress-text {
                        font-size: 13px;
                        font-weight: bold;
                    }
                    .finished-success .progress-bar-fill {
                        background-color: var(--vscode-testing-iconPassed);
                    }
                    .finished-error .progress-bar-fill {
                        background-color: var(--vscode-testing-iconFailed);
                    }
                    .hint {
                        font-size: 11px;
                        color: var(--vscode-descriptionForeground);
                        margin-top: 5px;
                    }
                </style>
            </head>
            <body>
                <div id="progress-container">
                    <div class="progress-bar-bg">
                        <div class="progress-bar-fill" id="progress-fill"></div>
                    </div>
                    <div id="progress-text">Submitting...</div>
                </div>

                <div class="content">
                    <h1>${titleStr}</h1>
                    ${renderedMarkdown}
                </div>

                <div class="action-bar">
                    <div style="margin-right: auto; align-self: center;">
                        <span class="hint">Note: Add your own test calls to run local tests.</span>
                    </div>
                    <button class="btn btn-secondary" onclick="runLocalTest()">▶ Run Local Test</button>
                    <button class="btn btn-primary" onclick="submitCode()">☁ Submit to Sandbox</button>
                </div>

                <script>
                    const vscode = acquireVsCodeApi();
                    
                    function submitCode() {
                        vscode.postMessage({ command: 'submitCode' });
                    }
                    
                    function runLocalTest() {
                        vscode.postMessage({ command: 'runLocalTest' });
                    }

                    window.addEventListener('message', event => {
                        const message = event.data;
                        if (message.type === 'progress') {
                            const container = document.getElementById('progress-container');
                            const fill = document.getElementById('progress-fill');
                            const text = document.getElementById('progress-text');
                            
                            container.style.display = 'block';
                            container.className = '';
                            
                            if (message.state === 'finished') {
                                fill.style.width = '100%';
                                container.classList.add(message.success ? 'finished-success' : 'finished-error');
                                text.innerText = message.text;
                                // Keep it visible for 8 seconds after finish
                                setTimeout(() => { container.style.display = 'none'; }, 8000);
                            } else {
                                fill.style.width = message.percent + '%';
                                text.innerText = message.text;
                            }
                            // No need to scroll, it's sticky!
                        }
                    });
                </script>
            </body>
            </html>`;
    }
}
