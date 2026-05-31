package project

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"ViewMD/internal/models"
)

var ignoredDirectories = map[string]struct{}{
	".git":         {},
	".idea":        {},
	".vscode":      {},
	"dist":         {},
	"node_modules": {},
	"vendor":       {},
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Summary(root string) (models.ProjectSummary, error) {
	cleanRoot, err := NormalizeRoot(root)
	if err != nil {
		return models.ProjectSummary{}, err
	}

	name := filepath.Base(cleanRoot)
	if name == "." || name == string(filepath.Separator) {
		name = cleanRoot
	}

	return models.ProjectSummary{
		Name: name,
		Path: cleanRoot,
	}, nil
}

func (s *Service) BuildTree(root string) ([]models.TreeNode, error) {
	cleanRoot, err := NormalizeRoot(root)
	if err != nil {
		return nil, err
	}

	return s.readDir(cleanRoot, "")
}

func NormalizeRoot(root string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", errors.New("project path is empty")
	}

	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	abs = filepath.Clean(abs)

	info, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", errors.New("project path is not a directory")
	}

	return abs, nil
}

func IsMarkdownFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md", ".markdown", ".mdown", ".mkd":
		return true
	default:
		return false
	}
}

func (s *Service) readDir(root string, relative string) ([]models.TreeNode, error) {
	current := filepath.Join(root, relative)
	entries, err := os.ReadDir(current)
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		leftInfo, leftErr := entries[i].Info()
		rightInfo, rightErr := entries[j].Info()
		if leftErr == nil && rightErr == nil && leftInfo.IsDir() != rightInfo.IsDir() {
			return leftInfo.IsDir()
		}
		return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
	})

	nodes := make([]models.TreeNode, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			continue
		}
		if info.IsDir() {
			if shouldSkipDirectory(name) {
				continue
			}
		}

		nodeRelative := filepath.ToSlash(filepath.Join(relative, name))
		node := models.TreeNode{
			Name:         name,
			Path:         nodeRelative,
			IsDir:        info.IsDir(),
			IsMarkdown:   IsMarkdownFile(name),
			Size:         info.Size(),
			ModifiedTime: info.ModTime().Format(time.RFC3339),
		}

		if info.IsDir() {
			children, err := s.readDir(root, filepath.Join(relative, name))
			if err == nil {
				node.Children = children
			}
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

func shouldSkipDirectory(name string) bool {
	_, ignored := ignoredDirectories[strings.ToLower(name)]
	return ignored
}
