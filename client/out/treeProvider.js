"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.ProblemNode = exports.PtLpoTreeProvider = exports.Problem = void 0;
const vscode = require("vscode");
const axios_1 = require("axios");
const SERVER_URL = 'http://localhost:8080/api';
class Problem {
    constructor(id, title, description, difficulty, userStatus) {
        this.id = id;
        this.title = title;
        this.description = description;
        this.difficulty = difficulty;
        this.userStatus = userStatus;
    }
}
exports.Problem = Problem;
class PtLpoTreeProvider {
    constructor(context) {
        this.context = context;
        this._onDidChangeTreeData = new vscode.EventEmitter();
        this.onDidChangeTreeData = this._onDidChangeTreeData.event;
    }
    refresh() {
        this._onDidChangeTreeData.fire();
    }
    getTreeItem(element) {
        return element;
    }
    async getChildren(element) {
        if (element) {
            return []; // No nested children for now
        }
        const token = await this.context.secrets.get('ptlpoj_jwt_token');
        if (!token) {
            vscode.window.showWarningMessage('PtLPOJ: Please log in to view problems.');
            return [];
        }
        try {
            const res = await axios_1.default.get(`${SERVER_URL}/problems`, {
                headers: { 'Authorization': `Bearer ${token}` }
            });
            return res.data.map((p) => new ProblemNode(`${p.id}: ${p.title} (${p.difficulty})`, p.id, p.difficulty, p.user_status, vscode.TreeItemCollapsibleState.None));
        }
        catch (error) {
            vscode.window.showErrorMessage(`PtLPOJ: Failed to fetch problem list - ${error.message}`);
            return [];
        }
    }
}
exports.PtLpoTreeProvider = PtLpoTreeProvider;
class ProblemNode extends vscode.TreeItem {
    constructor(label, problemId, difficulty, status, collapsibleState) {
        super(label, collapsibleState);
        this.label = label;
        this.problemId = problemId;
        this.difficulty = difficulty;
        this.status = status;
        this.collapsibleState = collapsibleState;
        // Connect node to our "view problem" command
        this.command = {
            title: "View Problem",
            command: "ptlpoj.openProblem",
            arguments: [this]
        };
        this.id = problemId.toString();
        this.tooltip = `Problem ${problemId} - ${difficulty}`;
        this.description = status;
        // Setup Icon and Icon Color based on status
        this.iconPath = new vscode.ThemeIcon(this.getIconForStatus(status), this.getColorForStatus(status));
    }
    getIconForStatus(status) {
        switch (status) {
            case 'AC': return 'pass';
            case 'WA': return 'error';
            case 'TLE': return 'watch';
            case 'RE': return 'warning';
            case 'PENDING':
            case 'RUNNING': return 'sync~spin';
            default: return 'circle-outline'; // UNATTEMPTED
        }
    }
    getColorForStatus(status) {
        switch (status) {
            case 'AC': return new vscode.ThemeColor('testing.iconPassed');
            case 'WA': return new vscode.ThemeColor('testing.iconFailed');
            case 'RE': return new vscode.ThemeColor('testing.iconErrored');
            case 'PENDING':
            case 'RUNNING': return new vscode.ThemeColor('testing.iconQueued');
            default: return undefined;
        }
    }
}
exports.ProblemNode = ProblemNode;
//# sourceMappingURL=treeProvider.js.map