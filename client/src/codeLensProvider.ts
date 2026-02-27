import * as vscode from 'vscode';
import * as path from 'path';

export class PtLpoCodeLensProvider implements vscode.CodeLensProvider {
    public provideCodeLenses(document: vscode.TextDocument, token: vscode.CancellationToken): vscode.CodeLens[] | Thenable<vscode.CodeLens[]> {
        const fileName = path.basename(document.uri.fsPath);
        // Only provide lenses for files matching Solution_*.py
        if (fileName.match(/^Solution_\d+\.py$/)) {
            const topOfDocument = new vscode.Range(0, 0, 0, 0);
            return [
                new vscode.CodeLens(topOfDocument, {
                    title: "▶ Run Local Test",
                    command: "ptlpoj.runTest",
                    arguments: [document.uri]
                }),
                new vscode.CodeLens(topOfDocument, {
                    title: "☁ Submit to Sandbox",
                    command: "ptlpoj.submitCode"
                })
            ];
        }
        return [];
    }
}
