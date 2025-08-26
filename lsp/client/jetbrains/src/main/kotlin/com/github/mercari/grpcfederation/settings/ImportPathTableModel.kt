package com.github.mercari.grpcfederation.settings

import com.intellij.util.ui.ItemRemovable
import javax.swing.table.AbstractTableModel

class ImportPathTableModel : AbstractTableModel(), ItemRemovable {
    private val entries = mutableListOf<ProjectSettingsService.ImportPathEntry>()
    
    companion object {
        private const val COLUMN_ENABLED = 0
        private const val COLUMN_PATH = 1
        private val COLUMN_NAMES = arrayOf("Enabled", "Import Path")
        private val COLUMN_CLASSES = arrayOf(Boolean::class.java, String::class.java)
    }

    fun getEntries(): List<ProjectSettingsService.ImportPathEntry> = entries.toList()
    
    fun setEntries(newEntries: List<ProjectSettingsService.ImportPathEntry>) {
        entries.clear()
        entries.addAll(newEntries)
        fireTableDataChanged()
    }
    
    fun addEntry(entry: ProjectSettingsService.ImportPathEntry) {
        entries.add(entry)
        fireTableRowsInserted(entries.size - 1, entries.size - 1)
    }
    
    fun removeEntry(index: Int) {
        if (index >= 0 && index < entries.size) {
            entries.removeAt(index)
            fireTableRowsDeleted(index, index)
        }
    }
    
    fun moveUp(index: Int) {
        if (index > 0 && index < entries.size) {
            val entry = entries.removeAt(index)
            entries.add(index - 1, entry)
            fireTableRowsUpdated(index - 1, index)
        }
    }
    
    fun moveDown(index: Int) {
        if (index >= 0 && index < entries.size - 1) {
            val entry = entries.removeAt(index)
            entries.add(index + 1, entry)
            fireTableRowsUpdated(index, index + 1)
        }
    }
    
    override fun getRowCount(): Int = entries.size
    
    override fun getColumnCount(): Int = COLUMN_NAMES.size
    
    override fun getColumnName(column: Int): String = COLUMN_NAMES[column]
    
    override fun getColumnClass(column: Int): Class<*> = COLUMN_CLASSES[column]
    
    override fun isCellEditable(rowIndex: Int, columnIndex: Int): Boolean = true
    
    override fun getValueAt(rowIndex: Int, columnIndex: Int): Any? {
        if (rowIndex >= entries.size) return null
        
        val entry = entries[rowIndex]
        return when (columnIndex) {
            COLUMN_ENABLED -> entry.enabled
            COLUMN_PATH -> entry.path
            else -> null
        }
    }
    
    override fun setValueAt(value: Any?, rowIndex: Int, columnIndex: Int) {
        if (rowIndex >= entries.size) return
        
        val entry = entries[rowIndex]
        when (columnIndex) {
            COLUMN_ENABLED -> entry.enabled = value as? Boolean ?: true
            COLUMN_PATH -> entry.path = value as? String ?: ""
        }
        fireTableCellUpdated(rowIndex, columnIndex)
    }
    
    override fun removeRow(idx: Int) {
        removeEntry(idx)
    }
}