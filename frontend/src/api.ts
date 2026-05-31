import {
  ChooseProjectFolder,
  CreateMarkdownFile,
  ExportMarkdownPDF,
  GetInitialState,
  OpenFile,
  RefreshProjectTree,
  RestoreLastProject,
  SaveFile,
  SaveWorkspaceState,
} from '../wailsjs/go/main/App';
import type {
  AppState,
  CreatedFileResult,
  FileDocument,
  ProjectSnapshot,
  TreeNode,
  ViewMode,
} from './types';

export const backend = {
  getInitialState: () => GetInitialState() as Promise<AppState>,
  restoreLastProject: () => RestoreLastProject() as Promise<ProjectSnapshot | null>,
  chooseProjectFolder: () => ChooseProjectFolder() as Promise<ProjectSnapshot | null>,
  refreshProjectTree: (projectPath: string) =>
    RefreshProjectTree(projectPath) as Promise<TreeNode[]>,
  openFile: (projectPath: string, filePath: string) =>
    OpenFile(projectPath, filePath) as Promise<FileDocument>,
  saveFile: (projectPath: string, filePath: string, content: string) =>
    SaveFile(projectPath, filePath, content) as Promise<FileDocument>,
  createMarkdownFile: (projectPath: string, folderPath: string, fileName: string) =>
    CreateMarkdownFile(projectPath, folderPath, fileName) as Promise<CreatedFileResult>,
  saveWorkspaceState: (
    projectPath: string,
    openFiles: string[],
    activeFile: string,
    viewMode: ViewMode,
  ) => SaveWorkspaceState(projectPath, openFiles, activeFile, viewMode) as Promise<void>,
  exportMarkdownPDF: (projectPath: string, markdownText: string, suggestedName: string) =>
    ExportMarkdownPDF(projectPath, markdownText, suggestedName) as Promise<string>,
};
