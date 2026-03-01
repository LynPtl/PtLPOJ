import * as vscode from 'vscode';
import axios from 'axios';
import * as fs from 'fs';

function getServerUrl(): string {
    return vscode.workspace.getConfiguration('ptlpoj').get<string>('serverUrl') || 'http://localhost:8080/api';
}

export class AdminViewPanel {
    public static currentPanel: AdminViewPanel | undefined;
    private readonly _panel: vscode.WebviewPanel;
    private readonly _extensionUri: vscode.Uri;
    private _disposables: vscode.Disposable[] = [];
    private _context: vscode.ExtensionContext;
    private _adminToken: string;

    public static async createOrShow(extensionUri: vscode.Uri, context: vscode.ExtensionContext) {
        const column = vscode.window.activeTextEditor ? vscode.window.activeTextEditor.viewColumn : undefined;

        if (AdminViewPanel.currentPanel) {
            AdminViewPanel.currentPanel._panel.reveal(column);
            return;
        }

        const panel = vscode.window.createWebviewPanel(
            'ptlpojAdmin',
            'PtLPOJ Admin Control Panel',
            column || vscode.ViewColumn.One,
            {
                enableScripts: true,
                localResourceRoots: [vscode.Uri.joinPath(extensionUri, 'resources')]
            }
        );

        let adminToken = await context.secrets.get('ptlpoj_admin_token') || '';
        AdminViewPanel.currentPanel = new AdminViewPanel(panel, extensionUri, context, adminToken);
    }

    private constructor(panel: vscode.WebviewPanel, extensionUri: vscode.Uri, context: vscode.ExtensionContext, adminToken: string) {
        this._panel = panel;
        this._extensionUri = extensionUri;
        this._context = context;
        this._adminToken = adminToken;

        this._update();
        this._panel.onDidDispose(() => this.dispose(), null, this._disposables);

        this._panel.webview.onDidReceiveMessage(
            async message => {
                switch (message.type) {
                    case 'fetchUsers':
                        await this._handleFetchUsers();
                        return;
                    case 'addUser':
                        await this._handleAddUser(message.emailInput);
                        return;
                    case 'askDeleteUsers':
                        vscode.window.showWarningMessage(`Are you sure you want to permanently remove these users?`, { modal: true }, 'Yes').then(async res => {
                            if (res === 'Yes') {
                                await this._handleDeleteUsers(message.emailInput);
                                this._panel.webview.postMessage({ type: 'clearEmailInput' });
                            }
                        });
                        return;
                    case 'askRemoveSingleUser':
                        vscode.window.showWarningMessage(`Are you sure you want to remove ${message.email}?`, { modal: true }, 'Yes').then(async res => {
                            if (res === 'Yes') {
                                await this._handleRemoveSingleUser(message.email);
                            }
                        });
                        return;
                    case 'askResetToken':
                        vscode.window.showWarningMessage('Are you sure you want to clear your local Admin Token? You will need to re-enter it.', { modal: true }, 'Yes').then(res => {
                            if (res === 'Yes') {
                                vscode.commands.executeCommand('ptlpoj.resetAdminToken');
                            }
                        });
                        return;
                    case 'selectAndUploadProblems':
                        await this._handleSelectAndUploadProblems();
                        return;
                }
            },
            null,
            this._disposables
        );
    }

    private async _handleFetchUsers() {
        try {
            const currentUserToken = await this._context.secrets.get('ptlpoj_jwt_token') || '';
            const res = await axios.get(`${getServerUrl()}/admin/users`, {
                headers: {
                    'Authorization': `Bearer ${currentUserToken}`,
                    'X-Admin-Token': this._adminToken
                }
            });
            this._panel.webview.postMessage({ type: 'usersData', data: res.data });
        } catch (error: any) {
            const errorMsg = error.response?.data?.error;
            if (error.response?.status === 401) {
                if (errorMsg === 'Invalid Admin Environment Token') {
                    vscode.window.showErrorMessage(`Authentication Failed: Invalid Admin Token.`, 'Reset Admin Token').then(res => {
                        if (res === 'Reset Admin Token') vscode.commands.executeCommand('ptlpoj.resetAdminToken');
                    });
                } else {
                    vscode.window.showErrorMessage(`Authentication Failed: User Session Invalid (${errorMsg}). Please Login to your account first.`, 'Login').then(res => {
                        if (res === 'Login') vscode.commands.executeCommand('ptlpoj.login');
                    });
                }
            } else if (error.response?.status === 403) {
                vscode.window.showErrorMessage(`Permission Denied: Your account role is not ADMIN.`);
            } else {
                vscode.window.showErrorMessage(`Fetch Users Failed: ${errorMsg || error.message}`);
            }
        }
    }

    private async _handleAddUser(emailInput: string) {
        const emails = emailInput.split(/[,;\n ]+/).map(e => e.trim()).filter(e => e.length > 0);
        if (emails.length === 0) return;

        let successCount = 0;
        let failCount = 0;

        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: "Adding Users to Whitelist",
            cancellable: false
        }, async (progress) => {
            const total = emails.length;
            for (let i = 0; i < total; i++) {
                progress.report({ message: `Processing ${i + 1}/${total}: ${emails[i]}` });
                try {
                    const currentUserToken = await this._context.secrets.get('ptlpoj_jwt_token') || '';
                    await axios.post(`${getServerUrl()}/admin/users`, { email: emails[i] }, {
                        headers: {
                            'Authorization': `Bearer ${currentUserToken}`,
                            'X-Admin-Token': this._adminToken
                        }
                    });
                    successCount++;
                } catch (error: any) {
                    const failReason = error.response?.data?.error || error.message;
                    console.error(`Failed to add user ${emails[i]}:`, failReason);
                    failCount++;
                    if (error.response?.status === 401 && failReason === 'Invalid Admin Environment Token') {
                        vscode.window.showErrorMessage(`Invalid Admin Token.`, 'Reset Admin Token').then(res => {
                            if (res === 'Reset Admin Token') vscode.commands.executeCommand('ptlpoj.resetAdminToken');
                        });
                        // Do not return here, continue processing other emails
                    } else if (error.response?.status === 401) {
                        vscode.window.showErrorMessage(`Session Invalid (${failReason}). Please Login first.`, 'Login').then(res => {
                            if (res === 'Login') vscode.commands.executeCommand('ptlpoj.login');
                        });
                        // Do not return here, continue processing other emails
                    } else if (error.response?.status === 403) {
                        vscode.window.showErrorMessage(`Permission Denied: You are not an ADMIN.`);
                        // Do not return here, continue processing other emails
                    }
                }
            }
        });

        if (failCount === 0) {
            vscode.window.showInformationMessage(`Successfully added all ${successCount} user(s) to whitelist!`);
        } else {
            vscode.window.showWarningMessage(`Added ${successCount} user(s), but ${failCount} failed. Check console or if they already exist.`);
        }

        this._handleFetchUsers();
    }

    private async _handleDeleteUsers(emailInput: string) {
        const emails = emailInput.split(/[,;\n ]+/).map(e => e.trim()).filter(e => e.length > 0);
        if (emails.length === 0) return;

        let successCount = 0;
        let failCount = 0;

        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: "Removing Users from Whitelist",
            cancellable: false
        }, async (progress) => {
            const total = emails.length;
            for (let i = 0; i < total; i++) {
                progress.report({ message: `Processing ${i + 1}/${total}: ${emails[i]}` });
                try {
                    const currentUserToken = await this._context.secrets.get('ptlpoj_jwt_token') || '';
                    await axios.delete(`${getServerUrl()}/admin/users?email=${encodeURIComponent(emails[i])}`, {
                        headers: {
                            'Authorization': `Bearer ${currentUserToken}`,
                            'X-Admin-Token': this._adminToken
                        }
                    });
                    successCount++;
                } catch (error: any) {
                    const failReason = error.response?.data?.error || error.message;
                    console.error(`Failed to remove user ${emails[i]}:`, failReason);
                    failCount++;
                    if (error.response?.status === 401 && failReason === 'Invalid Admin Environment Token') {
                        vscode.window.showErrorMessage(`Invalid Admin Token.`, 'Reset Admin Token').then(res => {
                            if (res === 'Reset Admin Token') vscode.commands.executeCommand('ptlpoj.resetAdminToken');
                        });
                        // Do not return here, continue processing other emails
                    } else if (error.response?.status === 401) {
                        vscode.window.showErrorMessage(`Session Invalid (${failReason}). Please Login first.`, 'Login').then(res => {
                            if (res === 'Login') vscode.commands.executeCommand('ptlpoj.login');
                        });
                        // Do not return here, continue processing other emails
                    } else if (error.response?.status === 403) {
                        vscode.window.showErrorMessage(`Permission Denied: You are not an ADMIN.`);
                        // Do not return here, continue processing other emails
                    }
                }
            }
        });

        if (failCount === 0) {
            vscode.window.showInformationMessage(`Successfully removed all ${successCount} user(s) from whitelist!`);
        } else {
            vscode.window.showWarningMessage(`Removed ${successCount} user(s), but ${failCount} failed. They may not exist.`);
        }

        this._handleFetchUsers();
    }

    private async _handleRemoveSingleUser(email: string) {
        try {
            const currentUserToken = await this._context.secrets.get('ptlpoj_jwt_token') || '';
            await axios.delete(`${getServerUrl()}/admin/users?email=${encodeURIComponent(email)}`, {
                headers: {
                    'Authorization': `Bearer ${currentUserToken}`,
                    'X-Admin-Token': this._adminToken
                }
            });
            vscode.window.showInformationMessage(`User ${email} removed.`);
            this._handleFetchUsers();
        } catch (error: any) {
            vscode.window.showErrorMessage(`Delete User Failed: ${error.response?.data?.error || error.message}`);
        }
    }

    private async _handleSelectAndUploadProblems() {
        // Native VS Code file picker
        const uris = await vscode.window.showOpenDialog({
            canSelectMany: true,
            openLabel: 'Select Python Files to Upload',
            filters: { 'Python': ['py'] }
        });

        if (!uris || uris.length === 0) {
            return;
        }

        let successCount = 0;
        let errMsg = '';

        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: "Uploading Problems",
            cancellable: false
        }, async (progress) => {
            const currentUserToken = await this._context.secrets.get('ptlpoj_jwt_token') || '';
            const total = uris.length;
            for (let i = 0; i < total; i++) {
                const fsPath = uris[i].fsPath;
                progress.report({ message: `Parsing and uploading ${i + 1}/${total}...` });

                try {
                    const fileStream = fs.createReadStream(fsPath);
                    const formData = new (require('form-data'))();
                    formData.append('python_file', fileStream);

                    const res = await axios.post(`${getServerUrl()}/admin/problems`, formData, {
                        headers: {
                            ...formData.getHeaders(),
                            'Authorization': `Bearer ${currentUserToken}`,
                            'X-Admin-Token': this._adminToken
                        }
                    });
                    successCount++;
                } catch (error: any) {
                    const failReason = error.response?.data?.error || error.message;
                    errMsg += `\n- ${uris[i].path.split('/').pop()}: ${failReason}`;
                }
            }
        });

        if (errMsg) {
            vscode.window.showErrorMessage(`Uploaded ${successCount}/${uris.length} problems. Failures:${errMsg}`, { modal: true });
        } else {
            vscode.window.showInformationMessage(`🎉 Successfully parsed and uploaded all ${successCount} problem(s). Please refresh problems list.`);
        }

        // Refresh users/problems view if we implement a problem list in admin
    }

    public dispose() {
        AdminViewPanel.currentPanel = undefined;
        this._panel.dispose();
        while (this._disposables.length) {
            const x = this._disposables.pop();
            if (x) {
                x.dispose();
            }
        }
    }

    private _update() {
        const webview = this._panel.webview;
        this._panel.webview.html = this._getHtmlForWebview(webview);
    }

    private _getHtmlForWebview(webview: vscode.Webview): string {
        return `<!DOCTYPE html>
            <html lang="en">
            <head>
                <meta charset="UTF-8">
                <meta name="viewport" content="width=device-width, initial-scale=1.0">
                <title>PtLPOJ Admin</title>
                <style>
                    body { font-family: var(--vscode-font-family); padding: 20px; color: var(--vscode-editor-foreground); background-color: var(--vscode-editor-background); }
                    .tab { overflow: hidden; border-bottom: 1px solid var(--vscode-panel-border); margin-bottom: 15px; }
                    .tab button { background-color: inherit; float: left; border: none; outline: none; cursor: pointer; padding: 10px 20px; color: var(--vscode-editor-foreground); transition: 0.3s; font-size: 14px; font-weight: bold; }
                    .tab button:hover { background-color: var(--vscode-list-hoverBackground); }
                    .tab button.active { border-bottom: 2px solid var(--vscode-button-background); color: var(--vscode-button-background); }
                    .tabcontent { display: none; padding: 10px 0; animation: fadeEffect 0.5s; }
                    @keyframes fadeEffect { from {opacity: 0;} to {opacity: 1;} }
                    
                    /* Tables */
                    table { width: 100%; border-collapse: collapse; margin-top: 15px; }
                    th, td { border-bottom: 1px solid var(--vscode-panel-border); padding: 10px; text-align: left; }
                    th { font-weight: bold; color: var(--vscode-editor-foreground); opacity: 0.8; }
                    
                    /* Forms & Buttons */
                    textarea { width: 100%; padding: 8px; margin: 8px 0; box-sizing: border-box; background: var(--vscode-input-background); color: var(--vscode-input-foreground); border: 1px solid var(--vscode-input-border); border-radius: 4px; resize: vertical; min-height: 80px; font-family: var(--vscode-editor-font-family); }
                    button.primary { background-color: var(--vscode-button-background); color: var(--vscode-button-foreground); border: none; padding: 10px 15px; cursor: pointer; border-radius: 4px; font-weight: bold; margin-top: 10px; }
                    button.primary:hover { background-color: var(--vscode-button-hoverBackground); }
                    button.secondary { background-color: transparent; color: var(--vscode-button-background); border: 1px solid var(--vscode-button-background); padding: 10px 15px; cursor: pointer; border-radius: 4px; font-weight: bold; margin-top: 10px; }
                    button.secondary:hover { background-color: var(--vscode-button-hoverBackground); color: var(--vscode-button-foreground); }
                    button.danger { background-color: var(--vscode-errorForeground); color: white; border: none; padding: 5px 10px; cursor: pointer; border-radius: 3px; font-size: 12px; }
                    button.danger:hover { opacity: 0.8; }
                    button.danger:hover { opacity: 0.8; }
                    
                    .card { background: var(--vscode-editorWidget-background); border: 1px solid var(--vscode-panel-border); border-radius: 6px; padding: 20px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); margin-top: 10px; }
                    .help-text { font-size: 12px; color: var(--vscode-descriptionForeground); margin-top: 5px; display: block; }
                </style>
            </head>
            <body>
                <div style="display: flex; justify-content: space-between; align-items: center;">
                    <h2>👑 PtLPOJ Admin Control Panel</h2>
                    <button class="secondary" style="margin-top: 0; font-size: 11px; padding: 5px 10px;" onclick="resetAdminToken()">🔄 Reset Admin Token</button>
                </div>
                
                <div class="tab">
                  <button class="tablinks active" onclick="openTab(event, 'Whitelist')">👥 Whitelist Manager</button>
                  <button class="tablinks" onclick="openTab(event, 'Problems')">📚 Problem Uploader</button>
                </div>

                <div id="Whitelist" class="tabcontent" style="display:block;">
                    <div class="card">
                        <h3>Manage Users</h3>
                        <p style="opacity: 0.8; font-size: 13px;">Supports bulk operations: enter multiple emails separated by commas, spaces, or newlines (e.g. copy-pasted from Excel column).</p>
                        <textarea id="emailsInput" placeholder="student1@example.com, student2@edu.cn\nstudent3@gmail.com"></textarea>
                        
                        <div style="display: flex; gap: 10px; margin-top: 5px;">
                            <button class="primary" onclick="addUsers()">+ Bulk Add to Whitelist</button>
                            <button class="secondary" onclick="deleteUsers()">- Bulk Delete Users</button>
                        </div>
                    </div>
                    
                    <h3>Authorized Users</h3>
                    <table>
                        <thead>
                            <tr><th>Email</th><th>Role</th><th>Action</th></tr>
                        </thead>
                        <tbody id="usersTableBody">
                            <tr><td colspan="3">Loading...</td></tr>
                        </tbody>
                    </table>
                </div>

                <div id="Problems" class="tabcontent">
                    <div class="card">
                        <h3>🚀 Smart Python Problem Uploader</h3>
                        <p style="opacity: 0.8; font-size: 13px;">Our system automatically utilizes AST Parsing to deeply analyze the native Python code structure you upload.</p>
                        <p style="opacity: 0.8; font-size: 13px;"><strong>Rule:</strong> The file must contain a function <code>def f(...):</code> and rigorous <code>doctests</code>. We will automatically generate the scaffolds and hidden tests.</p>
                        
                        <div style="margin-top: 20px;">
                            <p style="opacity: 0.9; font-weight: bold;">Step 1: Click the button below.</p>
                            <p style="opacity: 0.9;">Step 2: Select one or more <code>.py</code> files using the system dialog.</p>
                        </div>
                        
                        <button class="primary" style="margin-top: 20px; width: 100%;" onclick="uploadProblem()">🖥️ Select .py Files & Upload Bulk</button>
                    </div>
                </div>

                <script>
                    const vscode = acquireVsCodeApi();

                    function openTab(evt, tabName) {
                        var i, tabcontent, tablinks;
                        tabcontent = document.getElementsByClassName("tabcontent");
                        for (i = 0; i < tabcontent.length; i++) {
                            tabcontent[i].style.display = "none";
                        }
                        tablinks = document.getElementsByClassName("tablinks");
                        for (i = 0; i < tablinks.length; i++) {
                            tablinks[i].className = tablinks[i].className.replace(" active", "");
                        }
                        document.getElementById(tabName).style.display = "block";
                        evt.currentTarget.className += " active";
                        
                        if (tabName === 'Whitelist') {
                            vscode.postMessage({ type: 'fetchUsers' });
                        }
                    }

                    // Initial fetch
                    vscode.postMessage({ type: 'fetchUsers' });

                    window.addEventListener('message', event => {
                        const message = event.data;
                        if (message.type === 'clearEmailInput') {
                            document.getElementById('emailsInput').value = '';
                        }
                        if (message.type === 'usersData') {
                            const tbody = document.getElementById('usersTableBody');
                            tbody.innerHTML = '';
                            if (message.data.length === 0) {
                                tbody.innerHTML = '<tr><td colspan="3">No users found.</td></tr>';
                                return;
                            }
                            message.data.forEach(user => {
                                const tr = document.createElement('tr');
                                tr.innerHTML = \`
                                    <td>\${user.email}</td>
                                    <td><span style="background: var(--vscode-badge-background); color: var(--vscode-badge-foreground); padding: 2px 6px; border-radius: 10px; font-size: 11px;">\${user.role}</span></td>
                                    <td><button class="danger" onclick="removeSingleUser('\${user.email}')">Remove</button></td>
                                \`;
                                tbody.appendChild(tr);
                            });
                        }
                    });

                    function addUsers() {
                        const emailInput = document.getElementById('emailsInput');
                        const emails = emailInput.value.trim();
                        if (emails) {
                            vscode.postMessage({ type: 'addUser', emailInput: emails });
                            emailInput.value = '';
                        }
                    }

                    function deleteUsers() {
                        const emailInput = document.getElementById('emailsInput');
                        const emails = emailInput.value.trim();
                        if (emails) {
                            vscode.postMessage({ type: 'askDeleteUsers', emailInput: emails });
                        }
                    }

                    function removeSingleUser(email) {
                        vscode.postMessage({ type: 'askRemoveSingleUser', email });
                    }

                    function resetAdminToken() {
                        vscode.postMessage({ type: 'askResetToken' });
                    }

                    function uploadProblem() {
                        // Directly ask extension host to open the file picker and handle it natively
                        vscode.postMessage({ type: 'selectAndUploadProblems' });
                    }
                </script>
            </body>
            </html>`;
    }
}
