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

        let token = await context.secrets.get('ptlpoj_admin_token') || '';
        AdminViewPanel.currentPanel = new AdminViewPanel(panel, extensionUri, context, token);
    }

    private constructor(panel: vscode.WebviewPanel, extensionUri: vscode.Uri, context: vscode.ExtensionContext, token: string) {
        this._panel = panel;
        this._extensionUri = extensionUri;
        this._context = context;
        this._adminToken = token;

        this._update();

        this._panel.onDidDispose(() => this.dispose(), null, this._disposables);

        this._panel.webview.onDidReceiveMessage(
            async message => {
                switch (message.type) {
                    case 'fetchUsers':
                        await this._handleFetchUsers();
                        return;
                    case 'addUser':
                        await this._handleAddUser(message.email);
                        return;
                    case 'deleteUser':
                        await this._handleDeleteUser(message.email);
                        return;
                    case 'uploadProblem':
                        await this._handleUploadProblem(message.filePath);
                        return;
                }
            },
            null,
            this._disposables
        );
    }

    private async _handleFetchUsers() {
        try {
            const res = await axios.get(`${getServerUrl()}/admin/users`, {
                headers: { 'Authorization': `Bearer ${this._adminToken}` }
            });
            this._panel.webview.postMessage({ type: 'usersData', data: res.data });
        } catch (error: any) {
            vscode.window.showErrorMessage(`Fetch Users Failed: ${error.response?.data?.error || error.message}`);
        }
    }

    private async _handleAddUser(email: string) {
        try {
            await axios.post(`${getServerUrl()}/admin/users`, { email: email }, {
                headers: { 'Authorization': `Bearer ${this._adminToken}` }
            });
            vscode.window.showInformationMessage(`User ${email} added to whitelist!`);
            this._handleFetchUsers(); // Refresh
        } catch (error: any) {
            vscode.window.showErrorMessage(`Add User Failed: ${error.response?.data?.error || error.message}`);
        }
    }

    private async _handleDeleteUser(email: string) {
        try {
            await axios.delete(`${getServerUrl()}/admin/users?email=${encodeURIComponent(email)}`, {
                headers: { 'Authorization': `Bearer ${this._adminToken}` }
            });
            vscode.window.showInformationMessage(`User ${email} removed.`);
            this._handleFetchUsers(); // Refresh
        } catch (error: any) {
            vscode.window.showErrorMessage(`Delete User Failed: ${error.response?.data?.error || error.message}`);
        }
    }

    private async _handleUploadProblem(filePath: string) {
        try {
            const fileStream = fs.createReadStream(filePath);
            const formData = new (require('form-data'))();
            formData.append('python_file', fileStream);

            vscode.window.showInformationMessage('AST Parsing and Uploading Problem...');

            const res = await axios.post(`${getServerUrl()}/admin/problems`, formData, {
                headers: {
                    ...formData.getHeaders(),
                    'Authorization': `Bearer ${this._adminToken}`
                }
            });
            vscode.window.showInformationMessage(`✅ ${res.data.message}`);
        } catch (error: any) {
            vscode.window.showErrorMessage(`Upload Failed: ${error.response?.data?.error || error.message}`);
        }
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
                    input[type="text"], input[type="file"] { width: 100%; padding: 8px; margin: 8px 0; box-sizing: border-box; background: var(--vscode-input-background); color: var(--vscode-input-foreground); border: 1px solid var(--vscode-input-border); border-radius: 4px; }
                    button.primary { background-color: var(--vscode-button-background); color: var(--vscode-button-foreground); border: none; padding: 10px 15px; cursor: pointer; border-radius: 4px; font-weight: bold; margin-top: 10px; }
                    button.primary:hover { background-color: var(--vscode-button-hoverBackground); }
                    button.danger { background-color: var(--vscode-errorForeground); color: white; border: none; padding: 5px 10px; cursor: pointer; border-radius: 3px; font-size: 12px; }
                    button.danger:hover { opacity: 0.8; }
                    
                    .card { background: var(--vscode-editorWidget-background); border: 1px solid var(--vscode-panel-border); border-radius: 6px; padding: 20px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); margin-top: 10px; }
                    .help-text { font-size: 12px; color: var(--vscode-descriptionForeground); margin-top: 5px; display: block; }
                </style>
            </head>
            <body>
                <h2>👑 PtLPOJ Admin Control Panel</h2>
                
                <div class="tab">
                  <button class="tablinks active" onclick="openTab(event, 'Whitelist')">👥 Whitelist Manager</button>
                  <button class="tablinks" onclick="openTab(event, 'Problems')">📚 Problem Uploader</button>
                </div>

                <div id="Whitelist" class="tabcontent" style="display:block;">
                    <div class="card">
                        <h3>Add New User</h3>
                        <div style="display: flex; gap: 10px; align-items: center;">
                            <input type="text" id="newEmail" placeholder="student@example.com" style="flex: 1;" />
                            <button class="primary" style="margin-top: 0;" onclick="addUser()">Add to Whitelist</button>
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
                            <label for="problemFile" style="font-weight: bold;">Select .py Source File:</label>
                            <input type="file" id="problemFile" accept=".py" />
                            <span class="help-text">Example: "01_Add_Two_Numbers.py"</span>
                        </div>
                        
                        <button class="primary" style="margin-top: 20px; width: 100%;" onclick="uploadProblem()">Execute Smart Parse & Upload</button>
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
                                    <td><button class="danger" onclick="deleteUser('\${user.email}')">Remove</button></td>
                                \`;
                                tbody.appendChild(tr);
                            });
                        }
                    });

                    function addUser() {
                        const emailInput = document.getElementById('newEmail');
                        const email = emailInput.value.trim();
                        if (email) {
                            vscode.postMessage({ type: 'addUser', email });
                            emailInput.value = '';
                        }
                    }

                    function deleteUser(email) {
                        if (confirm('Are you sure you want to remove ' + email + '?')) {
                            vscode.postMessage({ type: 'deleteUser', email });
                        }
                    }

                    function uploadProblem() {
                        const fileInput = document.getElementById('problemFile');
                        if (fileInput.files.length > 0) {
                            // Since Webview cannot send raw files easily over postMessage without base64 bloat,
                            // we send the absolute path back to the extension host to read it directly.
                            // However, file.path is available in electron browsers for absolute path!
                            const file = fileInput.files[0];
                            if (file.path) {
                                vscode.postMessage({ type: 'uploadProblem', filePath: file.path });
                            } else {
                                alert("Failed to resolve file path.");
                            }
                        } else {
                            alert('Please select a .py file first.');
                        }
                    }
                </script>
            </body>
            </html>`;
    }
}
