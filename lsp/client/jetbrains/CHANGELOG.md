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

### Fixed
- Issue with paths containing spaces in the previous text field approach
- Better handling of empty and invalid path entries

## [0.1.0]

### Added
- Initial release of gRPC Federation IntelliJ plugin
- Basic Language Server Protocol (LSP) integration
- Proto import paths configuration
- Support for `.proto` files with gRPC Federation annotations
