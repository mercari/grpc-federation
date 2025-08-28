<!-- Plugin description -->
# gRPC Federation Language Server for JetBrains

The gRPC Federation Language Server for JetBrains is a plugin designed for JetBrains IDEs to integrate with the gRPC Federation language server, providing enhanced development experience for Protocol Buffer files with gRPC Federation annotations.

## Features

- **Go to Definition** - Navigate to proto message and field definitions
- **Diagnostics** - Real-time error detection and validation
- **Code Completion** - Intelligent suggestions for gRPC Federation annotations

## Requirements

- **IntelliJ IDEA Ultimate** 2023.3 or newer (This plugin requires Ultimate Edition)
- **grpc-federation-language-server** installed and available in PATH

## Installation

### Step 1: Install the Language Server
The plugin requires `grpc-federation-language-server` to be accessible in the system PATH. If Go is available, install it with:

```console
$ go install github.com/mercari/grpc-federation/cmd/grpc-federation-language-server@latest
```

Verify the installation:
```console
$ which grpc-federation-language-server
```

### Step 2: Install the Plugin
Install the plugin from JetBrains Marketplace:

1. Open **Settings/Preferences** → **Plugins**
2. Select **Marketplace** tab
3. Search for **"gRPC Federation"**
4. Click **Install** and restart the IDE

## Configuration

### Setting Import Paths
Configure proto import paths for the language server to resolve dependencies:

1. Navigate to **Settings/Preferences** → **Tools** → **gRPC Federation**
2. In the **Proto Import Paths** section:
   - Click **+** to add a new path
   - Use the file browser to select directories
   - Enable/disable paths using checkboxes
   - Reorder paths using ↑/↓ buttons
   - Click **Validate Paths** to verify all directories exist

### Supported Path Variables

- `~/` - User's home directory
- `${PROJECT_ROOT}` - Current project root directory

Example paths:
```
/usr/local/include
${PROJECT_ROOT}/proto
~/workspace/shared-protos
```

## Troubleshooting

### Language Server Not Found
If you see "grpc-federation-language-server not found" error:
1. Ensure the language server is installed (see Installation section)
2. Check if it's in your PATH: `which grpc-federation-language-server`
3. Restart the IDE after installation

### Proto Files Not Recognized
If proto files are not being analyzed:
1. Ensure the file extension is `.proto`
2. Check that import paths are correctly configured
3. Validate paths using the **Validate Paths** button
4. Check the IDE logs: **Help** → **Show Log in Finder/Explorer**

### Import Resolution Issues
If imports are not being resolved:
1. Add all necessary proto directories to import paths
2. Ensure paths are in the correct order (more specific paths first)
3. Check that all proto dependencies are available

## Support

- **Issues**: [GitHub Issues](https://github.com/mercari/grpc-federation/issues)
- **Documentation**: [gRPC Federation Documentation](https://github.com/mercari/grpc-federation)

<!-- Plugin description end -->
