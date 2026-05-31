package models

type AppState struct {
	LastProject *ProjectSummary `json:"lastProject,omitempty"`
	Workspace   WorkspaceState  `json:"workspace"`
}

type WorkspaceState struct {
	ProjectPath string   `json:"projectPath"`
	OpenFiles   []string `json:"openFiles"`
	ActiveFile  string   `json:"activeFile"`
	ViewMode    string   `json:"viewMode"`
}

type ProjectSummary struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type ProjectSnapshot struct {
	Project       ProjectSummary `json:"project"`
	Tree          []TreeNode     `json:"tree"`
	OpenDocuments []FileDocument `json:"openDocuments"`
	ActiveFile    string         `json:"activeFile"`
	ViewMode      string         `json:"viewMode"`
}

type TreeNode struct {
	Name         string     `json:"name"`
	Path         string     `json:"path"`
	IsDir        bool       `json:"isDir"`
	IsMarkdown   bool       `json:"isMarkdown"`
	Size         int64      `json:"size"`
	ModifiedTime string     `json:"modifiedTime"`
	Children     []TreeNode `json:"children,omitempty"`
}

type FileDocument struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	AbsolutePath string `json:"absolutePath"`
	Content      string `json:"content"`
	Size         int64  `json:"size"`
	ModifiedTime string `json:"modifiedTime"`
}

type CreatedFileResult struct {
	Document FileDocument `json:"document"`
	Tree     []TreeNode   `json:"tree"`
}
