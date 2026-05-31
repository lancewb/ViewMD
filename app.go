package main

import (
	"context"
	"errors"
	"os"

	"ViewMD/internal/exporter"
	"ViewMD/internal/files"
	"ViewMD/internal/models"
	"ViewMD/internal/preferences"
	"ViewMD/internal/project"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx        context.Context
	prefs      *preferences.Store
	projects   *project.Service
	files      *files.Service
	pdfExports *exporter.Service
}

func NewApp() *App {
	return &App{
		prefs:      preferences.NewStore(),
		projects:   project.NewService(),
		files:      files.NewService(),
		pdfExports: exporter.NewService(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) GetInitialState() (models.AppState, error) {
	prefs, err := a.prefs.Load()
	if err != nil {
		return models.AppState{}, err
	}

	state := models.AppState{Workspace: prefs.Workspace()}
	if prefs.LastProjectPath != "" {
		if summary, err := a.projects.Summary(prefs.LastProjectPath); err == nil {
			state.LastProject = &summary
		}
	}

	return state, nil
}

func (a *App) RestoreLastProject() (*models.ProjectSnapshot, error) {
	prefs, err := a.prefs.Load()
	if err != nil {
		return nil, err
	}
	if prefs.LastProjectPath == "" {
		return nil, errors.New("there is no previous project")
	}

	return a.loadProjectSnapshot(prefs.LastProjectPath, prefs.Workspace())
}

func (a *App) ChooseProjectFolder() (*models.ProjectSnapshot, error) {
	if a.ctx == nil {
		return nil, errors.New("application is not ready")
	}

	projectPath, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:                "Open Markdown Project",
		CanCreateDirectories: true,
	})
	if err != nil {
		return nil, err
	}
	if projectPath == "" {
		return nil, nil
	}

	workspace := models.WorkspaceState{ProjectPath: projectPath, ViewMode: "split"}
	snapshot, err := a.loadProjectSnapshot(projectPath, workspace)
	if err != nil {
		return nil, err
	}

	if err := a.prefs.Save(preferences.Preferences{
		LastProjectPath: snapshot.Project.Path,
		OpenFiles:       nil,
		ActiveFile:      "",
		ViewMode:        snapshot.ViewMode,
	}); err != nil {
		return nil, err
	}

	return snapshot, nil
}

func (a *App) RefreshProjectTree(projectPath string) ([]models.TreeNode, error) {
	return a.projects.BuildTree(projectPath)
}

func (a *App) OpenFile(projectPath string, filePath string) (models.FileDocument, error) {
	return a.files.Read(projectPath, filePath)
}

func (a *App) SaveFile(projectPath string, filePath string, content string) (models.FileDocument, error) {
	return a.files.Save(projectPath, filePath, content)
}

func (a *App) CreateMarkdownFile(projectPath string, folderPath string, fileName string) (models.CreatedFileResult, error) {
	document, err := a.files.CreateMarkdown(projectPath, folderPath, fileName)
	if err != nil {
		return models.CreatedFileResult{}, err
	}

	tree, err := a.projects.BuildTree(projectPath)
	if err != nil {
		return models.CreatedFileResult{}, err
	}

	return models.CreatedFileResult{
		Document: document,
		Tree:     tree,
	}, nil
}

func (a *App) SaveWorkspaceState(projectPath string, openFiles []string, activeFile string, viewMode string) error {
	summary, err := a.projects.Summary(projectPath)
	if err != nil {
		return err
	}

	return a.prefs.Save(preferences.Preferences{
		LastProjectPath: summary.Path,
		OpenFiles:       openFiles,
		ActiveFile:      activeFile,
		ViewMode:        viewMode,
	})
}

func (a *App) ExportMarkdownPDF(projectPath string, markdownText string, suggestedName string) (string, error) {
	return a.pdfExports.SaveMarkdownPDF(a.ctx, projectPath, markdownText, suggestedName)
}

func (a *App) loadProjectSnapshot(projectPath string, workspace models.WorkspaceState) (*models.ProjectSnapshot, error) {
	summary, err := a.projects.Summary(projectPath)
	if err != nil {
		return nil, err
	}

	tree, err := a.projects.BuildTree(summary.Path)
	if err != nil {
		return nil, err
	}

	openDocuments := make([]models.FileDocument, 0, len(workspace.OpenFiles))
	for _, filePath := range workspace.OpenFiles {
		document, err := a.files.Read(summary.Path, filePath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			continue
		}
		openDocuments = append(openDocuments, document)
	}

	activeFile := workspace.ActiveFile
	if activeFile == "" && len(openDocuments) > 0 {
		activeFile = openDocuments[0].Path
	}

	return &models.ProjectSnapshot{
		Project:       summary,
		Tree:          tree,
		OpenDocuments: openDocuments,
		ActiveFile:    activeFile,
		ViewMode:      preferences.NormalizeViewMode(workspace.ViewMode),
	}, nil
}
