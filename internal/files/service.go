package files

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ViewMD/internal/models"
	"ViewMD/internal/project"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Read(root string, relativePath string) (models.FileDocument, error) {
	absolutePath, cleanRelative, err := ResolveInside(root, relativePath)
	if err != nil {
		return models.FileDocument{}, err
	}

	data, err := os.ReadFile(absolutePath)
	if err != nil {
		return models.FileDocument{}, err
	}

	info, err := os.Stat(absolutePath)
	if err != nil {
		return models.FileDocument{}, err
	}

	return fileDocument(absolutePath, cleanRelative, string(data), info), nil
}

func (s *Service) Save(root string, relativePath string, content string) (models.FileDocument, error) {
	absolutePath, cleanRelative, err := ResolveInside(root, relativePath)
	if err != nil {
		return models.FileDocument{}, err
	}
	if err := os.WriteFile(absolutePath, []byte(content), 0o644); err != nil {
		return models.FileDocument{}, err
	}

	info, err := os.Stat(absolutePath)
	if err != nil {
		return models.FileDocument{}, err
	}

	return fileDocument(absolutePath, cleanRelative, content, info), nil
}

func (s *Service) CreateMarkdown(root string, folderPath string, fileName string) (models.FileDocument, error) {
	cleanName := sanitizeFileName(fileName)
	if cleanName == "" {
		return models.FileDocument{}, errors.New("file name is empty")
	}
	if !project.IsMarkdownFile(cleanName) {
		cleanName += ".md"
	}

	folderAbsolute, cleanFolder, err := ResolveInside(root, folderPath)
	if err != nil {
		return models.FileDocument{}, err
	}
	info, err := os.Stat(folderAbsolute)
	if err != nil {
		return models.FileDocument{}, err
	}
	if !info.IsDir() {
		return models.FileDocument{}, errors.New("target path is not a folder")
	}

	relative := filepath.ToSlash(filepath.Join(cleanFolder, cleanName))
	absolutePath, cleanRelative, err := ResolveInside(root, relative)
	if err != nil {
		return models.FileDocument{}, err
	}

	initialContent := "# " + strings.TrimSuffix(cleanName, filepath.Ext(cleanName)) + "\n\n"
	file, err := os.OpenFile(absolutePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return models.FileDocument{}, err
	}
	if _, err := file.WriteString(initialContent); err != nil {
		_ = file.Close()
		return models.FileDocument{}, err
	}
	if err := file.Close(); err != nil {
		return models.FileDocument{}, err
	}

	info, err = os.Stat(absolutePath)
	if err != nil {
		return models.FileDocument{}, err
	}

	return fileDocument(absolutePath, cleanRelative, initialContent, info), nil
}

func ResolveInside(root string, relativePath string) (string, string, error) {
	cleanRoot, err := project.NormalizeRoot(root)
	if err != nil {
		return "", "", err
	}

	trimmed := strings.TrimSpace(relativePath)
	trimmed = strings.TrimPrefix(trimmed, "/")
	trimmed = strings.TrimPrefix(trimmed, "\\")
	if filepath.IsAbs(trimmed) {
		return "", "", errors.New("absolute file paths are not allowed")
	}

	nativeRelative := filepath.Clean(filepath.FromSlash(trimmed))
	if nativeRelative == "." {
		nativeRelative = ""
	}
	if strings.HasPrefix(nativeRelative, ".."+string(filepath.Separator)) || nativeRelative == ".." {
		return "", "", errors.New("file path escapes the project")
	}

	target := filepath.Clean(filepath.Join(cleanRoot, nativeRelative))
	relToRoot, err := filepath.Rel(cleanRoot, target)
	if err != nil {
		return "", "", err
	}
	if relToRoot == ".." || strings.HasPrefix(relToRoot, ".."+string(filepath.Separator)) || filepath.IsAbs(relToRoot) {
		return "", "", errors.New("file path escapes the project")
	}

	return target, filepath.ToSlash(relToRoot), nil
}

func fileDocument(absolutePath string, cleanRelative string, content string, info os.FileInfo) models.FileDocument {
	return models.FileDocument{
		Name:         filepath.Base(absolutePath),
		Path:         filepath.ToSlash(cleanRelative),
		AbsolutePath: absolutePath,
		Content:      content,
		Size:         info.Size(),
		ModifiedTime: info.ModTime().Format(time.RFC3339),
	}
}

func sanitizeFileName(fileName string) string {
	clean := strings.TrimSpace(fileName)
	clean = strings.ReplaceAll(clean, "\\", "")
	clean = strings.ReplaceAll(clean, "/", "")
	clean = strings.ReplaceAll(clean, ":", "")
	clean = strings.ReplaceAll(clean, "*", "")
	clean = strings.ReplaceAll(clean, "?", "")
	clean = strings.ReplaceAll(clean, "\"", "")
	clean = strings.ReplaceAll(clean, "<", "")
	clean = strings.ReplaceAll(clean, ">", "")
	clean = strings.ReplaceAll(clean, "|", "")
	return clean
}
