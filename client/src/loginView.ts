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

        this._update();
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

    private _update() {
        this._panel.webview.html = this._getHtmlForWebview();
    }

    private _getHtmlForWebview() {
        return `
            <!DOCTYPE html>
            <html lang="en">
            <head>
                <meta charset="UTF-8">
                <meta name="viewport" content="width=device-width, initial-scale=1.0">
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
                        height: 70vh;
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
                    .error { color: var(--vscode-errorForeground); font-size: 12px; margin-top: 10px; text-align: center; }
                    .success { color: var(--vscode-testing-iconPassed); font-size: 12px; margin-top: 10px; text-align: center; }
                    #otp-step { display: none; }
                </style>
            </head>
            <body>
                <div class="container">
                    <h2>PtLPOJ 登录</h2>
                    
                    <div id="email-step">
                        <div class="input-group">
                            <label>邮箱地址</label>
                            <input type="email" id="email" placeholder="you@example.com" />
                        </div>
                        <button id="btn-send">发送验证码</button>
                    </div>

                    <div id="otp-step">
                        <p style="font-size: 12px; opacity: 0.8; text-align: center;" id="sent-msg"></p>
                        <div class="input-group">
                            <label>6 位验证码</label>
                            <input type="text" id="otp" maxlength="6" placeholder="000000" />
                        </div>
                        <button id="btn-verify">立即登录</button>
                        <p style="text-align: center; font-size: 11px; margin-top: 15px;">
                            <a href="#" id="back-link" style="color: var(--vscode-textLink-foreground); text-decoration: none;">修改邮箱</a>
                        </p>
                    </div>

                    <div id="msg" class="error"></div>
                </div>

                <script>
                    const vscode = acquireVsCodeApi();
                    const emailStep = document.getElementById('email-step');
                    const otpStep = document.getElementById('otp-step');
                    const emailInput = document.getElementById('email');
                    const otpInput = document.getElementById('otp');
                    const msgDiv = document.getElementById('msg');
                    const btnSend = document.getElementById('btn-send');
                    const btnVerify = document.getElementById('btn-verify');
                    const sentMsg = document.getElementById('sent-msg');
                    const backLink = document.getElementById('back-link');

                    let currentEmail = '';

                    btnSend.addEventListener('click', () => {
                        const email = emailInput.value.trim();
                        if (!email) return;
                        btnSend.disabled = true;
                        msgDiv.textContent = '';
                        vscode.postMessage({ command: 'requestOtp', email });
                    });

                    btnVerify.addEventListener('click', () => {
                        const otp = otpInput.value.trim();
                        if (otp.length !== 6) return;
                        btnVerify.disabled = true;
                        msgDiv.textContent = '';
                        vscode.postMessage({ command: 'verifyOtp', email: currentEmail, otp });
                    });

                    backLink.addEventListener('click', (e) => {
                        e.preventDefault();
                        otpStep.style.display = 'none';
                        emailStep.style.display = 'block';
                        btnSend.disabled = false;
                        msgDiv.textContent = '';
                    });

                    window.addEventListener('message', event => {
                        const message = event.data;
                        switch (message.type) {
                            case 'otpSent':
                                currentEmail = message.email;
                                emailStep.style.display = 'none';
                                otpStep.style.display = 'block';
                                sentMsg.textContent = '验证码已发送至：' + currentEmail;
                                break;
                            case 'error':
                                btnSend.disabled = false;
                                btnVerify.disabled = false;
                                msgDiv.className = 'error';
                                msgDiv.textContent = message.message;
                                break;
                        }
                    });
                </script>
            </body>
            </html>
        `;
    }
}
