import * as path from 'path';
import { ExtensionContext, workspace } from 'vscode';

import {
	LanguageClient,
	LanguageClientOptions,
	ServerOptions,
} from 'vscode-languageclient/node';
import * as constants from './constants';

let client: LanguageClient;

function getCommand(cmdPath: string, workspacePath: string) {
	if (path.isAbsolute(cmdPath) || cmdPath == constants.defaultCommand)
		return cmdPath;

	return path.join(workspacePath, cmdPath);
}

function getArgs(importPaths: string[], workspacePath: string) {
	const paths: string[] = [];
	for (const importPath of importPaths) {
		if (path.isAbsolute(importPath))
			paths.push(importPath);
		else
			paths.push(path.join(workspacePath, importPath));
	}
	return paths.map((path) => { return `-I${path}`; });
}

export function activate(context: ExtensionContext) {
	const config = workspace.getConfiguration(constants.configName);
	const cmdPath = config.get<string>("path", constants.defaultCommand);
	const importPaths = config.get<string[]>("import-paths", []);
	const folders = workspace.workspaceFolders;

	let serverOptions: ServerOptions;
	if (folders.length != 0) {
		const uri = folders[0].uri;
		const command = getCommand(cmdPath, uri.path);
		const args = getArgs(importPaths, uri.path);
		serverOptions = { command: command, args: args };
	} else {
		serverOptions = { command: constants.defaultCommand };
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
