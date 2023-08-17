import * as path from 'path';
import { workspace, ExtensionContext } from 'vscode';

import {
	LanguageClient,
	LanguageClientOptions,
	ServerOptions,
} from 'vscode-languageclient/node';

let client: LanguageClient;

export function activate(context: ExtensionContext) {
	const config = workspace.getConfiguration("grpc-federation");
	const cmdPath = config.get<string>("path");
	const importPaths = config.get<string[]>("import-paths");
	const folders = workspace.workspaceFolders;
	const defaultCommand = "grpc-federation-language-server"

	let serverOptions: ServerOptions
	if (folders.length != 0) {
		const uri = folders[0].uri;
		const command = cmdPath ? path.join(uri.path, cmdPath) : defaultCommand;
		const args = importPaths.map((importPath) => { return `-I${path.join(uri.path, importPath)}` });
		serverOptions = {command: command, args: args};
	} else {
		serverOptions = {command: defaultCommand }
	}

	const clientOptions: LanguageClientOptions = {
		documentSelector: [{ scheme: 'file', language: 'grpc-federation' }],
		synchronize: {
			// Notify the server about file changes to '.clientrc files contained in the workspace
			fileEvents: workspace.createFileSystemWatcher('**/.clientrc')
		}
	};

	// Create the language client and start the client.
	client = new LanguageClient(
		'grpc-federation',
		'GRPC Federation',
		serverOptions,
		clientOptions
	);

	client.start();
}

export function deactivate(): Thenable<void> | undefined {
	if (!client) {
		return undefined;
	}
	return client.stop();
}
