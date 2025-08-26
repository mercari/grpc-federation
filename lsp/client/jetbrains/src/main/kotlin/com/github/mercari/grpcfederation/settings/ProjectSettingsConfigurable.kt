package com.github.mercari.grpcfederation.settings

import com.github.mercari.grpcfederation.GrpcFederationBundle
import com.intellij.openapi.fileChooser.FileChooserDescriptorFactory
import com.intellij.openapi.options.Configurable
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.TextBrowseFolderListener
import com.intellij.openapi.ui.TextFieldWithBrowseButton
import com.intellij.ui.AnActionButton
import com.intellij.ui.AnActionButtonRunnable
import com.intellij.ui.TableUtil
import com.intellij.ui.ToolbarDecorator
import com.intellij.ui.table.JBTable
import com.intellij.util.ui.JBUI
import java.awt.BorderLayout
import java.awt.Component
import java.io.File
import javax.swing.*
import javax.swing.table.DefaultTableCellRenderer
import javax.swing.table.TableCellRenderer

class ProjectSettingsConfigurable(
        private val project: Project,
) : Configurable {
    private var mainPanel: JPanel? = null
    private val tableModel = ImportPathTableModel()
    private var table: JBTable? = null
    private var isModified = false
    
    override fun getDisplayName(): String = GrpcFederationBundle.message("settings.grpcfederation.title")
    
    override fun createComponent(): JComponent {
        mainPanel = JPanel(BorderLayout())
        
        // Load current settings into the model
        loadCurrentSettings()
        
        // Create the table
        table = JBTable(tableModel).apply {
            setShowGrid(false)
            intercellSpacing = JBUI.emptySize()
            emptyText.text = "No import paths configured"
            
            // Set column widths and configure checkbox for Enabled column
            columnModel.getColumn(0).apply {
                preferredWidth = 60
                maxWidth = 60
                
                // Set up checkbox renderer
                val checkBoxRenderer = object : TableCellRenderer {
                    private val checkBox = JCheckBox()
                    override fun getTableCellRendererComponent(
                        table: JTable,
                        value: Any?,
                        isSelected: Boolean,
                        hasFocus: Boolean,
                        row: Int,
                        column: Int
                    ): Component {
                        checkBox.isSelected = value as? Boolean ?: true
                        checkBox.horizontalAlignment = SwingConstants.CENTER
                        checkBox.background = if (isSelected) table.selectionBackground else table.background
                        return checkBox
                    }
                }
                cellRenderer = checkBoxRenderer
                
                // Set up checkbox editor
                cellEditor = DefaultCellEditor(JCheckBox().apply {
                    horizontalAlignment = SwingConstants.CENTER
                })
            }
            columnModel.getColumn(1).preferredWidth = 400
        }
        
        // Create toolbar with buttons
        val decorator = ToolbarDecorator.createDecorator(table!!)
                .setAddAction { _ ->
                    showAddPathDialog()
                }
                .setRemoveAction { _ ->
                    table?.let { TableUtil.removeSelectedItems(it) }
                    isModified = true
                }
                .setMoveUpAction { _ ->
                    val selectedRow = table!!.selectedRow
                    if (selectedRow > 0) {
                        tableModel.moveUp(selectedRow)
                        table!!.setRowSelectionInterval(selectedRow - 1, selectedRow - 1)
                        isModified = true
                    }
                }
                .setMoveDownAction { _ ->
                    val selectedRow = table!!.selectedRow
                    if (selectedRow >= 0 && selectedRow < tableModel.rowCount - 1) {
                        tableModel.moveDown(selectedRow)
                        table!!.setRowSelectionInterval(selectedRow + 1, selectedRow + 1)
                        isModified = true
                    }
                }
                .addExtraAction(object : AnActionButton("Browse", com.intellij.icons.AllIcons.Actions.Menu_open) {
                    override fun actionPerformed(e: com.intellij.openapi.actionSystem.AnActionEvent) {
                        val selectedRow = table!!.selectedRow
                        if (selectedRow >= 0) {
                            showEditPathDialog(selectedRow)
                        }
                    }
                    
                    override fun isEnabled(): Boolean {
                        return table!!.selectedRow >= 0
                    }
                })
        
        // Add validation button
        val validateButton = JButton("Validate Paths").apply {
            addActionListener {
                validatePaths()
            }
        }
        
        val decoratorPanel = decorator.createPanel()
        
        // Layout
        mainPanel!!.add(decoratorPanel, BorderLayout.CENTER)
        
        val bottomPanel = JPanel().apply {
            layout = BoxLayout(this, BoxLayout.X_AXIS)
            border = JBUI.Borders.emptyTop(8)
            add(validateButton)
            add(Box.createHorizontalGlue())
        }
        mainPanel!!.add(bottomPanel, BorderLayout.SOUTH)
        
        // Add listener for table changes
        tableModel.addTableModelListener {
            isModified = true
        }
        
        return mainPanel!!
    }
    
    private fun loadCurrentSettings() {
        val currentState = project.projectSettings.currentState
        
        // Check if we have new format entries
        if (currentState.importPathEntries.isNotEmpty()) {
            tableModel.setEntries(currentState.importPathEntries.map {
                ProjectSettingsService.ImportPathEntry(it.path, it.enabled)
            })
        } else if (currentState.importPaths.isNotEmpty()) {
            // Migrate from old format
            tableModel.setEntries(currentState.importPaths.map {
                ProjectSettingsService.ImportPathEntry(it, true)
            })
        }
    }
    
    private fun showAddPathDialog() {
        val textField = TextFieldWithBrowseButton()
        val descriptor = FileChooserDescriptorFactory.createSingleFolderDescriptor()
        descriptor.title = "Select Proto Import Path"
        textField.addBrowseFolderListener(TextBrowseFolderListener(descriptor, project))
        
        val panel = JPanel(BorderLayout()).apply {
            add(JLabel("Path:"), BorderLayout.WEST)
            add(textField, BorderLayout.CENTER)
            preferredSize = JBUI.size(400, 30)
        }
        
        val result = JOptionPane.showConfirmDialog(
                mainPanel,
                panel,
                "Add Import Path",
                JOptionPane.OK_CANCEL_OPTION,
                JOptionPane.PLAIN_MESSAGE
        )
        
        if (result == JOptionPane.OK_OPTION && textField.text.isNotEmpty()) {
            tableModel.addEntry(ProjectSettingsService.ImportPathEntry(textField.text, true))
            isModified = true
        }
    }
    
    private fun showEditPathDialog(row: Int) {
        val currentPath = tableModel.getValueAt(row, 1) as String
        val textField = TextFieldWithBrowseButton()
        textField.text = currentPath
        
        val descriptor = FileChooserDescriptorFactory.createSingleFolderDescriptor()
        descriptor.title = "Select Proto Import Path"
        textField.addBrowseFolderListener(TextBrowseFolderListener(descriptor, project))
        
        val panel = JPanel(BorderLayout()).apply {
            add(JLabel("Path:"), BorderLayout.WEST)
            add(textField, BorderLayout.CENTER)
            preferredSize = JBUI.size(400, 30)
        }
        
        val result = JOptionPane.showConfirmDialog(
                mainPanel,
                panel,
                "Edit Import Path",
                JOptionPane.OK_CANCEL_OPTION,
                JOptionPane.PLAIN_MESSAGE
        )
        
        if (result == JOptionPane.OK_OPTION && textField.text.isNotEmpty()) {
            tableModel.setValueAt(textField.text, row, 1)
            isModified = true
        }
    }
    
    private fun validatePaths() {
        val entries = tableModel.getEntries()
        val invalidPaths = mutableListOf<String>()
        
        for (entry in entries) {
            if (entry.enabled) {
                val path = expandPath(entry.path)
                val file = File(path)
                if (!file.exists() || !file.isDirectory) {
                    invalidPaths.add(entry.path)
                }
            }
        }
        
        if (invalidPaths.isEmpty()) {
            JOptionPane.showMessageDialog(
                    mainPanel,
                    "All enabled paths are valid.",
                    "Validation Result",
                    JOptionPane.INFORMATION_MESSAGE
            )
        } else {
            JOptionPane.showMessageDialog(
                    mainPanel,
                    "The following paths are invalid:\n" + invalidPaths.joinToString("\n"),
                    "Validation Result",
                    JOptionPane.WARNING_MESSAGE
            )
        }
    }
    
    private fun expandPath(path: String): String {
        var expanded = path
        if (path.startsWith("~/")) {
            expanded = System.getProperty("user.home") + path.substring(1)
        }
        // Could add more variable expansions here like ${PROJECT_ROOT}
        if (path.contains("\${PROJECT_ROOT}")) {
            expanded = expanded.replace("\${PROJECT_ROOT}", project.basePath ?: "")
        }
        return expanded
    }
    
    override fun isModified(): Boolean {
        if (isModified) return true
        
        val currentEntries = project.projectSettings.currentState.importPathEntries
        val tableEntries = tableModel.getEntries()
        
        if (currentEntries.size != tableEntries.size) return true
        
        return !currentEntries.zip(tableEntries).all { (current, table) ->
            current.path == table.path && current.enabled == table.enabled
        }
    }
    
    override fun apply() {
        val entries = tableModel.getEntries()
        val newState = ProjectSettingsService.State(
                importPaths = entries.filter { it.enabled }.map { it.path }.toMutableList(),
                importPathEntries = entries.map { 
                    ProjectSettingsService.ImportPathEntry(it.path, it.enabled) 
                }.toMutableList()
        )
        project.projectSettings.currentState = newState
        isModified = false
    }
    
    override fun reset() {
        loadCurrentSettings()
        isModified = false
    }
    
    override fun disposeUIResources() {
        mainPanel = null
        table = null
    }
}