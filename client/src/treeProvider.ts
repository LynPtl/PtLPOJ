import * as vscode from 'vscode';
import axios from 'axios';

function getServerUrl(): string {
    return vscode.workspace.getConfiguration('ptlpoj').get<string>('serverUrl') || 'http://localhost:8080/api';
}

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

    private filterStatus: string = 'ALL';
    private filterTag: string = 'ALL';
    private searchQuery: string = '';
    private sortBy: 'id' | 'status' | 'difficulty' = 'id';

    constructor(private context: vscode.ExtensionContext) { }

    getFilterSortState() {
        return { status: this.filterStatus, tag: this.filterTag, sort: this.sortBy };
    }

    public async getAllTags(): Promise<string[]> {
        const token = await this.context.secrets.get('ptlpoj_jwt_token');
        if (!token) return [];
        try {
            const res = await axios.get(`${getServerUrl()}/problems`, {
                headers: { 'Authorization': `Bearer ${token}` }
            });
            const tags = new Set<string>();
            res.data.forEach((p: any) => {
                if (p.tags) p.tags.forEach((t: string) => tags.add(t));
            });
            return Array.from(tags).sort();
        } catch {
            return [];
        }
    }

    setFilterSort(status: string, tag: string, sort: 'id' | 'status' | 'difficulty') {
        this.filterStatus = status;
        this.filterTag = tag;
        this.sortBy = sort;
        this.refresh();
    }

    setSearchQuery(query: string) {
        this.searchQuery = query.toLowerCase();
        this.refresh();
    }

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
            return [new ProblemNode(
                "$(sign-in) 点击登录 PtLPOJ",
                -1, // Special ID for onboarding
                "",
                "UNAUTHENTICATED",
                vscode.TreeItemCollapsibleState.None
            )];
        }

        try {
            const res = await axios.get(`${getServerUrl()}/problems`, {
                headers: { 'Authorization': `Bearer ${token}` }
            });

            let problems = res.data;

            // 1. Filter by Status
            if (this.filterStatus !== 'ALL') {
                problems = problems.filter((p: any) => p.user_status === this.filterStatus);
            }

            // 2. Filter by Tag
            if (this.filterTag !== 'ALL') {
                problems = problems.filter((p: any) => p.tags && p.tags.includes(this.filterTag));
            }

            // 2b. Filter by Search Query
            if (this.searchQuery) {
                problems = problems.filter((p: any) =>
                    p.title.toLowerCase().includes(this.searchQuery) ||
                    p.id.toString().includes(this.searchQuery)
                );
            }

            // 3. Sort
            problems.sort((a: any, b: any) => {
                if (this.sortBy === 'id') return a.id - b.id;
                if (this.sortBy === 'difficulty') {
                    const diffScale: any = { 'Easy': 1, 'Medium': 2, 'Hard': 3 };
                    return (diffScale[a.difficulty] || 0) - (diffScale[b.difficulty] || 0);
                }
                if (this.sortBy === 'status') {
                    const statusScale: any = { 'AC': 1, 'PENDING': 2, 'RUNNING': 2, 'WA': 3, 'TLE': 3, 'RE': 3, 'UNATTEMPTED': 4 };
                    return (statusScale[a.user_status] || 99) - (statusScale[b.user_status] || 99);
                }
                return 0;
            });

            return problems.map((p: any) => new ProblemNode(
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

        if (problemId === -1) {
            this.command = {
                title: "Login",
                command: "ptlpoj.login",
                arguments: []
            };
            this.contextValue = 'onboarding';
        }
    }

    // Connect node to our "view problem" command
    command?: vscode.Command = {
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
