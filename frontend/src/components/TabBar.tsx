import { FileText, X } from 'lucide-react';
import type { EditorTab } from '../types';

interface TabBarProps {
  tabs: EditorTab[];
  activePath: string;
  onSelect: (path: string) => void;
  onClose: (path: string) => void;
}

export function TabBar({ tabs, activePath, onSelect, onClose }: TabBarProps) {
  return (
    <div className="tabbar" role="tablist">
      {tabs.length === 0 ? (
        <span className="tabbar-empty">未打开文件</span>
      ) : (
        tabs.map((tab) => (
          <div
            key={tab.path}
            className={tab.path === activePath ? 'tab active' : 'tab'}
            role="tab"
            aria-selected={tab.path === activePath}
          >
            <button type="button" className="tab-select" onClick={() => onSelect(tab.path)}>
              <FileText size={15} />
              <span>{tab.name}</span>
              {tab.dirty && <i aria-label="Unsaved changes" />}
            </button>
            <button
              type="button"
              className="tab-close"
              onClick={() => onClose(tab.path)}
              title="关闭"
            >
              <X size={14} />
            </button>
          </div>
        ))
      )}
    </div>
  );
}
