<!-- Keep a Changelog guide -> https://keepachangelog.com -->

# gRPC Federation Changelog

## [Unreleased]

### Added
- New list-based UI for Proto Import Paths configuration
- Enable/disable toggle for individual import paths
- File browser dialog for adding and editing paths
- Path validation feature to check if directories exist
- Variable expansion support for `~/` and `${PROJECT_ROOT}`
- Toolbar buttons for add, remove, and reorder operations
- "Validate Paths" button to verify all configured paths
- Semantic tokens support for enhanced syntax highlighting of gRPC Federation annotations
  - Token type mappings for proper colorization of proto elements
  - Explicit semantic tokens request for proto files

### Changed
- Redesigned settings interface from single text field to table-based UI
- Improved path management with better visual feedback
- Enhanced user experience for managing multiple import paths
- Migrated settings storage format from simple string list to structured entries with enable/disable state
- Updated to IntelliJ Platform Gradle Plugin 2.7.2 (latest)
- Updated to Platform Version 2024.3.3 with Kotlin 2.1.0
- Updated JVM toolchain from Java 17 to Java 21 (recommended for 2024.3.3)
- Changed LSP server descriptor from ProjectWideLspServerDescriptor to LspServerDescriptor for proper semantic tokens support

### Fixed
- Issue with paths containing spaces in the previous text field approach
- Better handling of empty and invalid path entries
- Checkbox display issue in the Enabled column with explicit renderer and editor
- Diagnostic messages containing HTML tags (like `<input>`) are now properly escaped to prevent UI rendering issues

### Deprecated
- `importPaths` field in settings storage (will be removed in v0.3.0)
  - Automatic migration: Existing configurations will be automatically converted to the new format
  - The plugin maintains backward compatibility by reading from both old and new formats
  - Only the new `importPathEntries` format will be supported from v0.3.0 onwards

## [0.1.0]

### Added
- Initial release of gRPC Federation IntelliJ plugin
- Basic Language Server Protocol (LSP) integration
- Proto import paths configuration
- Support for `.proto` files with gRPC Federation annotations
