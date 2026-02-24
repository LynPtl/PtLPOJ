import * as vscode from 'vscode';
import axios from 'axios';

const SERVER_URL = 'http://localhost:8080/api';

export class Problem {
    constructor(
        public readonly id: number,
        public readonly title: string,
        public readonly description: string,
        public readonly difficulty: string,
        public readonly userStatus: string
    ) { }
}

export class PtLpoTreeProvider implements vscode.TreeDataProvider<ProblemNode> {
    private _onDidChangeTreeData: vscode.EventEmitter<ProblemNode | undefined | null | void> = new vscode.EventEmitter<ProblemNode | undefined | null | void>();
    readonly onDidChangeTreeData: vscode.Event<ProblemNode | undefined | null | void> = this._onDidChangeTreeData.event;

    constructor(private context: vscode.ExtensionContext) { }

    refresh(): void {
        this._onDidChangeTreeData.fire();
    }

    getTreeItem(element: ProblemNode): vscode.TreeItem {
        return element;
    }

    async getChildren(element?: ProblemNode): Promise<ProblemNode[]> {
        if (element) {
            return []; // No nested children for now
        }

        const token = await this.context.secrets.get('ptlpoj_jwt_token');
        if (!token) {
            vscode.window.showWarningMessage('PtLPOJ: Please log in to view problems.');
            return [];
        }

        try {
            const res = await axios.get(`${SERVER_URL}/problems`, {
                headers: { 'Authorization': `Bearer ${token}` }
            });

            return res.data.map((p: any) => new ProblemNode(
                `${p.id}: ${p.title} (${p.difficulty})`,
                p.id,
                p.difficulty,
                p.user_status,
                vscode.TreeItemCollapsibleState.None
            ));
        } catch (error: any) {
            vscode.window.showErrorMessage(`PtLPOJ: Failed to fetch problem list - ${error.message}`);
            return [];
        }
    }
}

export class ProblemNode extends vscode.TreeItem {
    constructor(
        public readonly label: string,
        public readonly problemId: number,
        public readonly difficulty: string,
        public readonly status: string,
        public readonly collapsibleState: vscode.TreeItemCollapsibleState
    ) {
        super(label, collapsibleState);
        this.id = problemId.toString();
        this.tooltip = `Problem ${problemId} - ${difficulty}`;
        this.description = status;

        // Setup Icon and Icon Color based on status
        this.iconPath = new vscode.ThemeIcon(this.getIconForStatus(status), this.getColorForStatus(status));
    }

    // Connect node to our "view problem" command
    command = {
        title: "View Problem",
        command: "ptlpoj.openProblem",
        arguments: [this]
    };

    private getIconForStatus(status: string): string {
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

    private getColorForStatus(status: string): vscode.ThemeColor | undefined {
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
