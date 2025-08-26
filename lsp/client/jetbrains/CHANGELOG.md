<!-- Keep a Changelog guide -> https://keepachangelog.com -->

# gRPC Federation Changelog

## [Unreleased]

## [0.2.0]

### Added
- New list-based UI for Proto Import Paths configuration
- Enable/disable toggle for individual import paths
- File browser dialog for adding and editing paths
- Path validation feature to check if directories exist
- Variable expansion support for `~/` and `${PROJECT_ROOT}`
- Toolbar buttons for add, remove, and reorder operations
- "Validate Paths" button to verify all configured paths

### Changed
- Redesigned settings interface from single text field to table-based UI
- Improved path management with better visual feedback
- Enhanced user experience for managing multiple import paths
- Migrated settings storage format from simple string list to structured entries with enable/disable state

### Fixed
- Issue with paths containing spaces in the previous text field approach
- Better handling of empty and invalid path entries
- Checkbox display issue in the Enabled column with explicit renderer and editor

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
