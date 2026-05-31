export type ViewMode = 'split' | 'code' | 'preview';

export interface ProjectSummary {
  name: string;
  path: string;
}

export interface WorkspaceState {
  projectPath: string;
  openFiles: string[];
  activeFile: string;
  viewMode: ViewMode;
}

export interface AppState {
  lastProject?: ProjectSummary | null;
  workspace: WorkspaceState;
}

export interface TreeNode {
  name: string;
  path: string;
  isDir: boolean;
  isMarkdown: boolean;
  size: number;
  modifiedTime: string;
  children?: TreeNode[];
}

export interface FileDocument {
  name: string;
  path: string;
  absolutePath: string;
  content: string;
  size: number;
  modifiedTime: string;
}

export interface EditorTab extends FileDocument {
  dirty: boolean;
}

export interface ProjectSnapshot {
  project: ProjectSummary;
  tree: TreeNode[];
  openDocuments: FileDocument[];
  activeFile: string;
  viewMode: ViewMode;
}

export interface CreatedFileResult {
  document: FileDocument;
  tree: TreeNode[];
}

export function normalizeViewMode(value: string | undefined | null): ViewMode {
  if (value === 'code' || value === 'preview' || value === 'split') {
    return value;
  }
  return 'split';
}
