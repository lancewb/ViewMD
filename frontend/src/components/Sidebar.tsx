import { useEffect, useState } from 'react';
import {
  ChevronDown,
  ChevronRight,
  FileCode2,
  FilePlus2,
  FileText,
  Folder,
  FolderOpen,
  RefreshCw,
} from 'lucide-react';
import type { ProjectSummary, TreeNode } from '../types';

interface SidebarProps {
  project: ProjectSummary;
  tree: TreeNode[];
  activePath: string;
  busy: boolean;
  onOpenFile: (node: TreeNode) => void;
  onCreateFile: (folderPath: string) => void;
  onRefresh: () => void;
}

export function Sidebar({
  project,
  tree,
  activePath,
  busy,
  onOpenFile,
  onCreateFile,
  onRefresh,
}: SidebarProps) {
  const [expanded, setExpanded] = useState<Set<string>>(new Set());

  useEffect(() => {
    setExpanded(new Set(tree.filter((node) => node.isDir).map((node) => node.path)));
  }, [project.path, tree]);

  const toggle = (path: string) => {
    setExpanded((current) => {
      const next = new Set(current);
      if (next.has(path)) {
        next.delete(path);
      } else {
        next.add(path);
      }
      return next;
    });
  };

  return (
    <aside className="sidebar">
      <div className="project-head">
        <div>
          <p>工程</p>
          <h2>{project.name}</h2>
        </div>
        <div className="sidebar-actions">
          <button type="button" title="新建 Markdown" onClick={() => onCreateFile('')}>
            <FilePlus2 size={17} />
          </button>
          <button type="button" title="刷新" onClick={onRefresh} disabled={busy}>
            <RefreshCw size={16} />
          </button>
        </div>
      </div>

      <div className="project-path" title={project.path}>
        {project.path}
      </div>

      <nav className="file-tree" aria-label="Project tree">
        {tree.length === 0 ? (
          <div className="tree-empty">空工程</div>
        ) : (
          tree.map((node) => (
            <TreeItem
              key={node.path}
              node={node}
              level={0}
              expanded={expanded}
              activePath={activePath}
              onToggle={toggle}
              onOpenFile={onOpenFile}
              onCreateFile={onCreateFile}
            />
          ))
        )}
      </nav>
    </aside>
  );
}

interface TreeItemProps {
  node: TreeNode;
  level: number;
  expanded: Set<string>;
  activePath: string;
  onToggle: (path: string) => void;
  onOpenFile: (node: TreeNode) => void;
  onCreateFile: (folderPath: string) => void;
}

function TreeItem({
  node,
  level,
  expanded,
  activePath,
  onToggle,
  onOpenFile,
  onCreateFile,
}: TreeItemProps) {
  const isExpanded = expanded.has(node.path);
  const isActive = activePath === node.path;
  const left = 12 + level * 15;

  if (node.isDir) {
    return (
      <div className="tree-branch">
        <div className="tree-row tree-row-folder" style={{ paddingLeft: left }}>
          <button className="tree-main" type="button" onClick={() => onToggle(node.path)}>
            {isExpanded ? <ChevronDown size={15} /> : <ChevronRight size={15} />}
            {isExpanded ? <FolderOpen size={16} /> : <Folder size={16} />}
            <span>{node.name}</span>
          </button>
          <button
            className="tree-inline-action"
            type="button"
            title="在此新建 Markdown"
            onClick={(event) => {
              event.stopPropagation();
              onCreateFile(node.path);
            }}
          >
            <FilePlus2 size={13} />
          </button>
        </div>
        {isExpanded && node.children?.map((child) => (
          <TreeItem
            key={child.path}
            node={child}
            level={level + 1}
            expanded={expanded}
            activePath={activePath}
            onToggle={onToggle}
            onOpenFile={onOpenFile}
            onCreateFile={onCreateFile}
          />
        ))}
      </div>
    );
  }

  return (
    <button
      className={isActive ? 'tree-row tree-row-file active' : 'tree-row tree-row-file'}
      type="button"
      style={{ paddingLeft: left + 30 }}
      onClick={() => node.isMarkdown && onOpenFile(node)}
      disabled={!node.isMarkdown}
      title={node.path}
    >
      {node.isMarkdown ? <FileText size={15} /> : <FileCode2 size={15} />}
      <span>{node.name}</span>
    </button>
  );
}
