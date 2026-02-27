import * as vscode from 'vscode';
import axios from 'axios';

function getServerUrl(): string {
    return vscode.workspace.getConfiguration('ptlpoj').get<string>('serverUrl') || 'http://localhost:8080/api';
}

export class LoginViewPanel {
    public static currentPanel: LoginViewPanel | undefined;
    private readonly _panel: vscode.WebviewPanel;
    private readonly _extensionUri: vscode.Uri;
    private _disposables: vscode.Disposable[] = [];

    public static createOrShow(extensionUri: vscode.Uri) {
        const column = vscode.window.activeTextEditor ? vscode.window.activeTextEditor.viewColumn : undefined;

        if (LoginViewPanel.currentPanel) {
            LoginViewPanel.currentPanel._panel.reveal(column);
            return;
        }

        const panel = vscode.window.createWebviewPanel(
            'ptlpojLogin',
            'PtLPOJ Login',
            column || vscode.ViewColumn.One,
            {
                enableScripts: true,
                retainContextWhenHidden: true,
                localResourceRoots: [extensionUri]
            }
        );

        LoginViewPanel.currentPanel = new LoginViewPanel(panel, extensionUri);
    }

    private constructor(panel: vscode.WebviewPanel, extensionUri: vscode.Uri) {
        this._panel = panel;
        this._extensionUri = extensionUri;

        this._panel.webview.html = this._getHtmlForWebview();
        this._panel.onDidDispose(() => this.dispose(), null, this._disposables);

        this._panel.webview.onDidReceiveMessage(
            async message => {
                switch (message.command) {
                    case 'requestOtp':
                        await this._handleRequestOtp(message.email);
                        return;
                    case 'verifyOtp':
                        await this._handleVerifyOtp(message.email, message.otp);
                        return;
                    case 'updateConfig':
                        vscode.workspace.getConfiguration('ptlpoj').update('serverUrl', message.url, vscode.ConfigurationTarget.Global).then(() => {
                            this._panel.webview.postMessage({ type: 'configUpdated' });
                            // Trigger refresh of tree view
                            vscode.commands.executeCommand('ptlpoj.refreshProblems');
                        });
                        return;
                }
            },
            null,
            this._disposables
        );
    }

    public dispose() {
        LoginViewPanel.currentPanel = undefined;
        this._panel.dispose();
        while (this._disposables.length) {
            const x = this._disposables.pop();
            if (x) {
                x.dispose();
            }
        }
    }

    private async _handleRequestOtp(email: string) {
        try {
            await axios.post(`${getServerUrl()}/auth/login`, { email });
            this._panel.webview.postMessage({ type: 'otpSent', email });
        } catch (err: any) {
            this._panel.webview.postMessage({ type: 'error', message: err.response?.data?.error || err.message });
        }
    }

    private async _handleVerifyOtp(email: string, otp: string) {
        try {
            const res = await axios.post(`${getServerUrl()}/auth/verify`, { email, code: otp });
            const token = res.data.token;

            // Emit success to extension to store token
            vscode.commands.executeCommand('ptlpoj.completeLogin', token);

            this.dispose();
        } catch (err: any) {
            this._panel.webview.postMessage({ type: 'error', message: err.response?.data?.error || err.message });
        }
    }

    private _getHtmlForWebview() {
        return `
            <!DOCTYPE html>
            <html lang="en">
            <head>
                <meta charset="UTF-8">
                <style>
                    body {
                        font-family: var(--vscode-font-family);
                        padding: 40px;
                        color: var(--vscode-foreground);
                        background-color: var(--vscode-editor-background);
                        display: flex;
                        flex-direction: column;
                        align-items: center;
                        justify-content: center;
                        height: 85vh;
                    }
                    .container {
                        max-width: 400px;
                        width: 100%;
                        background: var(--vscode-sideBar-background);
                        padding: 30px;
                        border-radius: 8px;
                        box-shadow: 0 4px 15px rgba(0,0,0,0.3);
                        border: 1px solid var(--vscode-widget-border);
                    }
                    h2 { color: var(--vscode-textLink-foreground); margin-bottom: 20px; text-align: center; }
                    .input-group { margin-bottom: 20px; }
                    label { display: block; margin-bottom: 8px; font-weight: bold; font-size: 13px; }
                    input {
                        width: 100%;
                        padding: 10px;
                        box-sizing: border-box;
                        background: var(--vscode-input-background);
                        color: var(--vscode-input-foreground);
                        border: 1px solid var(--vscode-input-border);
                        border-radius: 4px;
                        outline: none;
                    }
                    input:focus { border-color: var(--vscode-focusBorder); }
                    button {
                        width: 100%;
                        padding: 12px;
                        background: var(--vscode-button-background);
                        color: var(--vscode-button-foreground);
                        border: none;
                        border-radius: 4px;
                        cursor: pointer;
                        font-weight: bold;
                        transition: opacity 0.2s;
                    }
                    button:hover { background: var(--vscode-button-hoverBackground); }
                    button:disabled { opacity: 0.5; cursor: not-allowed; }
                    .feedback { margin-top: 15px; font-size: 12px; text-align: center; min-height: 1.5em; }
                    .feedback-error { color: var(--vscode-errorForeground); }
                    .feedback-success { color: var(--vscode-testing-iconPassed); }
                    
                    .server-config {
                        margin-top: 25px;
                        border-top: 1px solid var(--vscode-widget-border);
                        padding-top: 15px;
                        text-align: center;
                    }
                    .link-btn {
                        background: none;
                        border: none;
                        color: var(--vscode-textLink-foreground);
                        cursor: pointer;
                        font-size: 11px;
                        text-decoration: underline;
                        padding: 0;
                        width: auto;
                    }
                    #configPanel {
                        text-align: left;
                        background: var(--vscode-editor-background);
                        padding: 10px;
                        border-radius: 4px;
                        border: 1px solid var(--vscode-widget-border);
                    }
                    #otpStep { display: none; }
                </style>
            </head>
            <body>
                <div class="container">
                    <h2>PtLPOJ 登录</h2>
                    
                    <div id="emailStep">
                        <div class="input-group">
                            <label>邮箱地址 (Email)</label>
                            <input type="email" id="email" placeholder="you@example.com" />
                        </div>
                        <button id="btnSend">发送验证码 (Request OTP)</button>
                    </div>

                    <div id="otpStep">
                        <p style="font-size: 12px; opacity: 0.8; text-align: center;" id="sentMsg"></p>
                        <div class="input-group">
                            <label>6 位验证码 (OTP)</label>
                            <input type="text" id="otp" maxlength="6" placeholder="000000" />
                        </div>
                        <button id="btnVerify">立即登录 (Login)</button>
                        <p style="text-align: center; font-size: 11px; margin-top: 15px;">
                            <a href="#" id="backLink" style="color: var(--vscode-textLink-foreground); text-decoration: none;">← 修改邮箱</a>
                        </p>
                    </div>

                    <div class="feedback" id="feedbackArea"></div>

                    <div class="server-config">
                        <button class="link-btn" id="toggleConfig">⚙️ 修改服务器端口/地址 (Server Config)</button>
                        <div id="configPanel" style="display: none; margin-top: 10px;">
                            <label style="font-size: 11px;">API Base URL:</label>
                            <input type="text" id="serverUrlInput" style="font-size: 11px; padding: 5px;" value="${getServerUrl()}">
                            <button id="saveConfig" style="margin-top: 10px; padding: 6px; font-size: 11px; background: var(--vscode-button-secondaryBackground); color: var(--vscode-button-secondaryForeground);">保存并刷新 (Save & Refresh)</button>
                        </div>
                    </div>
                </div>

                <script>
                    const vscode = acquireVsCodeApi();
                    const emailStep = document.getElementById('emailStep');
                    const otpStep = document.getElementById('otpStep');
                    const feedbackArea = document.getElementById('feedbackArea');
                    const btnSend = document.getElementById('btnSend');
                    const btnVerify = document.getElementById('btnVerify');
                    let currentEmail = '';

                    document.getElementById('btnSend').onclick = () => {
                        const email = document.getElementById('email').value.trim();
                        if (!email) return;
                        currentEmail = email;
                        btnSend.disabled = true;
                        feedbackArea.className = 'feedback';
                        feedbackArea.innerText = '正在发送请求...';
                        vscode.postMessage({ command: 'requestOtp', email });
                    };

                    document.getElementById('btnVerify').onclick = () => {
                        const otp = document.getElementById('otp').value.trim();
                        if (otp.length !== 6) return;
                        btnVerify.disabled = true;
                        feedbackArea.className = 'feedback';
                        feedbackArea.innerText = '正在验证中...';
                        vscode.postMessage({ command: 'verifyOtp', email: currentEmail, otp });
                    };

                    document.getElementById('backLink').onclick = (e) => {
                        e.preventDefault();
                        otpStep.style.display = 'none';
                        emailStep.style.display = 'block';
                        btnSend.disabled = false;
                        feedbackArea.innerText = '';
                    };

                    document.getElementById('toggleConfig').onclick = () => {
                        const p = document.getElementById('configPanel');
                        p.style.display = p.style.display === 'none' ? 'block' : 'none';
                    };

                    document.getElementById('saveConfig').onclick = () => {
                        const newUrl = document.getElementById('serverUrlInput').value.trim();
                        feedbackArea.className = 'feedback';
                        feedbackArea.innerText = '正在保存设置...';
                        vscode.postMessage({ command: 'updateConfig', url: newUrl });
                    };

                    window.addEventListener('message', event => {
                        const message = event.data;
                        switch (message.type) {
                            case 'otpSent':
                                emailStep.style.display = 'none';
                                otpStep.style.display = 'block';
                                document.getElementById('sentMsg').innerText = '验证码已发送至: ' + currentEmail;
                                feedbackArea.className = 'feedback feedback-success';
                                feedbackArea.innerText = '验证码已发送，请查收控制台/邮件';
                                break;
                            case 'error':
                                btnSend.disabled = false;
                                btnVerify.disabled = false;
                                feedbackArea.className = 'feedback feedback-error';
                                feedbackArea.innerText = '失败: ' + message.message;
                                break;
                            case 'configUpdated':
                                feedbackArea.className = 'feedback feedback-success';
                                feedbackArea.innerText = '服务器地址已更新！';
                                break;
                        }
                    });
                </script>
            </body>
            </html>
        `;
    }
}
