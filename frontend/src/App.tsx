import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import gsap from 'gsap';
import { useGSAP } from '@gsap/react';
import { FileText } from 'lucide-react';
import './App.css';
import { backend } from './api';
import { EditorPane } from './components/EditorPane';
import { EmptyEditor } from './components/EmptyEditor';
import { PreviewPane } from './components/PreviewPane';
import { Sidebar } from './components/Sidebar';
import { StartScreen } from './components/StartScreen';
import { StatusBar } from './components/StatusBar';
import { TabBar } from './components/TabBar';
import { Toolbar } from './components/Toolbar';
import type {
  AppState,
  EditorTab,
  FileDocument,
  ProjectSnapshot,
  ProjectSummary,
  TreeNode,
  ViewMode,
} from './types';
import { normalizeViewMode } from './types';

gsap.registerPlugin(useGSAP);

function toEditorTab(document: FileDocument): EditorTab {
  return { ...document, dirty: false };
}

function errorText(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }
  return String(error);
}

function pdfName(fileName: string): string {
  return fileName.replace(/\.(md|markdown|mdown|mkd)$/i, '') + '.pdf';
}

function App() {
  const appRef = useRef<HTMLDivElement>(null);
  const [initialState, setInitialState] = useState<AppState | null>(null);
  const [initialLoading, setInitialLoading] = useState(true);
  const [project, setProject] = useState<ProjectSummary | null>(null);
  const [tree, setTree] = useState<TreeNode[]>([]);
  const [tabs, setTabs] = useState<EditorTab[]>([]);
  const [activePath, setActivePath] = useState('');
  const [viewMode, setViewMode] = useState<ViewMode>('split');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');

  const activeTab = useMemo(
    () => tabs.find((tab) => tab.path === activePath),
    [activePath, tabs],
  );
  const dirty = Boolean(activeTab?.dirty);
  const openPathKey = useMemo(() => tabs.map((tab) => tab.path).join('|'), [tabs]);
  const openFiles = useMemo(() => tabs.map((tab) => tab.path), [openPathKey]);

  useGSAP(() => {
    const mm = gsap.matchMedia();

    mm.add('(prefers-reduced-motion: no-preference)', () => {
      gsap.from('.workspace-animate', {
        autoAlpha: 0,
        y: 14,
        duration: 0.48,
        stagger: 0.04,
        ease: 'power2.out',
      });
    });

    return () => mm.revert();
  }, { scope: appRef, dependencies: [project?.path], revertOnUpdate: true });

  useEffect(() => {
    backend
      .getInitialState()
      .then(setInitialState)
      .catch((caught) => setError(errorText(caught)))
      .finally(() => setInitialLoading(false));
  }, []);

  useEffect(() => {
    if (!project) {
      return;
    }

    const timer = window.setTimeout(() => {
      backend
        .saveWorkspaceState(project.path, openFiles, activePath, viewMode)
        .catch((caught) => console.warn('Failed to persist workspace', caught));
    }, 300);

    return () => window.clearTimeout(timer);
  }, [activePath, openFiles, project, viewMode]);

  useEffect(() => {
    if (!notice) {
      return;
    }

    const timer = window.setTimeout(() => setNotice(''), 2600);
    return () => window.clearTimeout(timer);
  }, [notice]);

  const loadSnapshot = useCallback((snapshot: ProjectSnapshot | null) => {
    if (!snapshot) {
      return;
    }

    const nextTabs = snapshot.openDocuments.map(toEditorTab);
    const nextActive = nextTabs.some((tab) => tab.path === snapshot.activeFile)
      ? snapshot.activeFile
      : nextTabs[0]?.path ?? '';

    setProject(snapshot.project);
    setTree(snapshot.tree ?? []);
    setTabs(nextTabs);
    setActivePath(nextActive);
    setViewMode(normalizeViewMode(snapshot.viewMode));
    setError('');
    setNotice('');
  }, []);

  const runBackendTask = useCallback(async (task: () => Promise<void>) => {
    setBusy(true);
    setError('');
    try {
      await task();
    } catch (caught) {
      setError(errorText(caught));
    } finally {
      setBusy(false);
    }
  }, []);

  const handleChooseFolder = useCallback(() => {
    void runBackendTask(async () => {
      const snapshot = await backend.chooseProjectFolder();
      loadSnapshot(snapshot);
    });
  }, [loadSnapshot, runBackendTask]);

  const handleRestore = useCallback(() => {
    void runBackendTask(async () => {
      const snapshot = await backend.restoreLastProject();
      loadSnapshot(snapshot);
    });
  }, [loadSnapshot, runBackendTask]);

  const handleRefreshTree = useCallback(() => {
    if (!project) {
      return;
    }

    void runBackendTask(async () => {
      setTree(await backend.refreshProjectTree(project.path));
    });
  }, [project, runBackendTask]);

  const handleOpenFile = useCallback((node: TreeNode) => {
    if (!project || !node.isMarkdown) {
      return;
    }

    const existing = tabs.find((tab) => tab.path === node.path);
    if (existing) {
      setActivePath(existing.path);
      return;
    }

    void runBackendTask(async () => {
      const document = await backend.openFile(project.path, node.path);
      setTabs((current) => {
        if (current.some((tab) => tab.path === document.path)) {
          return current;
        }
        return [...current, toEditorTab(document)];
      });
      setActivePath(document.path);
    });
  }, [project, runBackendTask, tabs]);

  const handleCreateFile = useCallback((folderPath: string) => {
    if (!project) {
      return;
    }

    const fileName = window.prompt('文件名', 'untitled.md');
    if (!fileName) {
      return;
    }

    void runBackendTask(async () => {
      const result = await backend.createMarkdownFile(project.path, folderPath, fileName);
      setTree(result.tree);
      setTabs((current) => {
        if (current.some((tab) => tab.path === result.document.path)) {
          return current;
        }
        return [...current, toEditorTab(result.document)];
      });
      setActivePath(result.document.path);
    });
  }, [project, runBackendTask]);

  const handleContentChange = useCallback((value: string) => {
    setTabs((current) =>
      current.map((tab) => (
        tab.path === activePath ? { ...tab, content: value, dirty: true } : tab
      )),
    );
  }, [activePath]);

  const handleSaveActive = useCallback(() => {
    if (!project || !activeTab) {
      return;
    }

    void runBackendTask(async () => {
      const saved = await backend.saveFile(project.path, activeTab.path, activeTab.content);
      setTabs((current) =>
        current.map((tab) => (
          tab.path === saved.path ? { ...toEditorTab(saved), content: saved.content } : tab
        )),
      );
      setNotice('已保存');
    });
  }, [activeTab, project, runBackendTask]);

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 's') {
        event.preventDefault();
        handleSaveActive();
      }
    };

    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [handleSaveActive]);

  const handleCloseTab = useCallback((path: string) => {
    const tab = tabs.find((item) => item.path === path);
    if (tab?.dirty && !window.confirm('文件尚未保存，关闭此标签页？')) {
      return;
    }

    const index = tabs.findIndex((item) => item.path === path);
    const nextTabs = tabs.filter((item) => item.path !== path);
    setTabs(nextTabs);

    if (activePath === path) {
      setActivePath(nextTabs[Math.min(index, nextTabs.length - 1)]?.path ?? '');
    }
  }, [activePath, tabs]);

  const handleExportPDF = useCallback(() => {
    if (!project || !activeTab) {
      return;
    }

    void runBackendTask(async () => {
      const outputPath = await backend.exportMarkdownPDF(
        project.path,
        activeTab.content,
        pdfName(activeTab.name),
      );
      if (outputPath) {
        setNotice(`PDF 已导出：${outputPath}`);
      }
    });
  }, [activeTab, project, runBackendTask]);

  if (initialLoading) {
    return (
      <div className="boot-screen">
        <FileText size={34} />
        <span>ViewMD</span>
      </div>
    );
  }

  return (
    <div id="App" className="app-shell" ref={appRef}>
      {project ? (
        <div className="workspace-layout">
          <div className="workspace-animate">
            <Sidebar
              project={project}
              tree={tree}
              activePath={activePath}
              busy={busy}
              onOpenFile={handleOpenFile}
              onCreateFile={handleCreateFile}
              onRefresh={handleRefreshTree}
            />
          </div>

          <section className="workbench workspace-animate">
            <Toolbar
              viewMode={viewMode}
              dirty={dirty}
              hasActiveFile={Boolean(activeTab)}
              busy={busy}
              onModeChange={setViewMode}
              onSave={handleSaveActive}
              onExportPDF={handleExportPDF}
              onOpenProject={handleChooseFolder}
            />
            <TabBar
              tabs={tabs}
              activePath={activePath}
              onSelect={setActivePath}
              onClose={handleCloseTab}
            />
            <div className={`editor-stage mode-${viewMode}`}>
              {activeTab ? (
                <>
                  {viewMode !== 'preview' && (
                    <EditorPane
                      fileName={activeTab.name}
                      value={activeTab.content}
                      onChange={handleContentChange}
                    />
                  )}
                  {viewMode !== 'code' && (
                    <PreviewPane content={activeTab.content} fileName={activeTab.name} />
                  )}
                </>
              ) : (
                <EmptyEditor />
              )}
            </div>
            <StatusBar activeTab={activeTab} viewMode={viewMode} busy={busy} />
          </section>
        </div>
      ) : (
        <StartScreen
          lastProject={initialState?.lastProject}
          busy={busy}
          onRestore={handleRestore}
          onChooseFolder={handleChooseFolder}
        />
      )}

      {(error || notice) && (
        <div className={error ? 'toast error' : 'toast'}>
          {error || notice}
        </div>
      )}
    </div>
  );
}

export default App;
