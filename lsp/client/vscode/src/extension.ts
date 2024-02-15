import * as path from 'path';
import { workspace, Uri, ExtensionContext, window } from 'vscode';
import { Wasm, ProcessOptions, Stdio } from '@vscode/wasm-wasi';
import { runServerProcess } from './lspServer';

import {
	LanguageClient,
	LanguageClientOptions,
	ServerOptions,
} from 'vscode-languageclient/node';

let client: LanguageClient;
const channel = window.createOutputChannel('gRPC Federation LSP');

export async function activate(context: ExtensionContext) {
	const wasm: Wasm = await Wasm.load();

	const config = workspace.getConfiguration("grpc-federation");
	const cmdPath = config.get<string>("path");
	const importPaths = config.get<string[]>("import-paths");
	const folders = workspace.workspaceFolders;
	const defaultCommand = "grpc-federation-language-server"

	const serverOptions: ServerOptions = async () => {
		const stdio: Stdio = {
			in: {
				kind: 'pipeIn',
			},
			out: {
				kind: 'pipeOut'
			},
			err: {
				kind: 'pipeOut'
			}
		};

		let args = null;
		if (folders.length != 0) {
			const uri = folders[0].uri;
			const command = cmdPath ? path.join(uri.path, cmdPath) : defaultCommand;
			args = importPaths.map((importPath) => { return `-I${path.join(uri.path, importPath)}` });
		}

		const options: ProcessOptions = {
			stdio: stdio,
			mountPoints: [
				{ kind: 'vscodeFileSystem', uri: Uri.parse("/"), mountPoint: "/" },
			],
			args: args,
		};
		const filename = Uri.joinPath(context.extensionUri, 'grpc-federation-language-server.wasm');
		const bits = await workspace.fs.readFile(filename);
		const module = await WebAssembly.compile(bits);
		const process = await wasm.createProcess('grpc-federation-language-server', module, { initial: 160, maximum: 160, shared: true }, options);

		const decoder = new TextDecoder('utf-8');
		process.stderr!.onData((data) => {
		    const v = decoder.decode(data);
			console.debug(v);
			channel.append(decoder.decode(data));
		});

		return runServerProcess(process);
	};

	const clientOptions: LanguageClientOptions = {
		documentSelector: [{ scheme: 'file', language: 'grpc-federation' }],
		synchronize: {
			// Notify the server about file changes to '.clientrc files contained in the workspace
			fileEvents: workspace.createFileSystemWatcher('**/.clientrc')
		},
		outputChannel: channel,
	};

	// Create the language client and start the client.
	client = new LanguageClient(
		'grpc-federation',
		'gRPC Federation',
		serverOptions,
		clientOptions
	);
	try {
		await client.start();
	} catch (error) {
		client.error(`Start failed`, error, 'force');
	}
}

export function deactivate(): Thenable<void> | undefined {
	if (!client) {
		return undefined;
	}
	return client.stop();
}
