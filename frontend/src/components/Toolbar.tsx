import { Code2, Columns2, Eye, FileDown, FolderOpen, Save } from 'lucide-react';
import type { ViewMode } from '../types';

interface ToolbarProps {
  viewMode: ViewMode;
  dirty: boolean;
  hasActiveFile: boolean;
  busy: boolean;
  onModeChange: (mode: ViewMode) => void;
  onSave: () => void;
  onExportPDF: () => void;
  onOpenProject: () => void;
}

const modes: Array<{ value: ViewMode; label: string; icon: typeof Columns2 }> = [
  { value: 'split', label: '双栏', icon: Columns2 },
  { value: 'code', label: '代码', icon: Code2 },
  { value: 'preview', label: '预览', icon: Eye },
];

export function Toolbar({
  viewMode,
  dirty,
  hasActiveFile,
  busy,
  onModeChange,
  onSave,
  onExportPDF,
  onOpenProject,
}: ToolbarProps) {
  return (
    <header className="workspace-toolbar">
      <div className="mode-switch" aria-label="View mode">
        {modes.map(({ value, label, icon: Icon }) => (
          <button
            key={value}
            className={viewMode === value ? 'active' : ''}
            type="button"
            onClick={() => onModeChange(value)}
            title={label}
          >
            <Icon size={16} />
            <span>{label}</span>
          </button>
        ))}
      </div>

      <div className="toolbar-actions">
        <button type="button" onClick={onOpenProject} title="选择文件夹" disabled={busy}>
          <FolderOpen size={17} />
          <span>工程</span>
        </button>
        <button
          type="button"
          className={dirty ? 'attention' : ''}
          onClick={onSave}
          disabled={!hasActiveFile || busy}
          title="保存"
        >
          <Save size={17} />
          <span>保存</span>
        </button>
        <button
          type="button"
          onClick={onExportPDF}
          disabled={!hasActiveFile || busy}
          title="导出 PDF"
        >
          <FileDown size={17} />
          <span>PDF</span>
        </button>
      </div>
    </header>
  );
}
