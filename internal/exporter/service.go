package exporter

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) SaveMarkdownPDF(ctx context.Context, projectPath string, markdownText string, suggestedName string) (string, error) {
	if ctx == nil {
		return "", errors.New("application is not ready")
	}

	defaultName := sanitizePDFName(suggestedName)
	targetPath, err := wailsruntime.SaveFileDialog(ctx, wailsruntime.SaveDialogOptions{
		Title:           "Export PDF",
		DefaultFilename: defaultName,
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "PDF Document (*.pdf)", Pattern: "*.pdf"},
		},
		CanCreateDirectories: true,
	})
	if err != nil {
		return "", err
	}
	if targetPath == "" {
		return "", nil
	}
	if !strings.EqualFold(filepath.Ext(targetPath), ".pdf") {
		targetPath += ".pdf"
	}

	if err := RenderMarkdownPDF(markdownText, projectPath, targetPath); err != nil {
		_ = os.Remove(targetPath)
		return "", err
	}

	return targetPath, nil
}

func sanitizePDFName(name string) string {
	clean := strings.TrimSpace(name)
	if clean == "" {
		clean = "ViewMD-export"
	}
	clean = strings.TrimSuffix(clean, filepath.Ext(clean))

	replacer := strings.NewReplacer(
		"\\", "",
		"/", "",
		":", "",
		"*", "",
		"?", "",
		"\"", "",
		"<", "",
		">", "",
		"|", "",
	)
	clean = replacer.Replace(clean)
	if clean == "" {
		clean = "ViewMD-export"
	}

	return clean + ".pdf"
}
