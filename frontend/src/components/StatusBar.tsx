import { Circle, FileText } from 'lucide-react';
import type { EditorTab, ViewMode } from '../types';

interface StatusBarProps {
  activeTab?: EditorTab;
  viewMode: ViewMode;
  busy: boolean;
}

const modeLabel: Record<ViewMode, string> = {
  split: '双栏',
  code: '代码',
  preview: '预览',
};

export function StatusBar({ activeTab, viewMode, busy }: StatusBarProps) {
  return (
    <footer className="statusbar">
      <span>
        <Circle size={9} className={busy ? 'pulse' : ''} />
        {busy ? '处理中' : '就绪'}
      </span>
      <span>{modeLabel[viewMode]}</span>
      {activeTab && (
        <>
          <span>
            <FileText size={13} />
            {activeTab.path}
          </span>
          <span>{activeTab.content.length.toLocaleString()} chars</span>
        </>
      )}
    </footer>
  );
}
