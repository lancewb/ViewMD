package exporter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderMarkdownPDF(t *testing.T) {
	target := filepath.Join(t.TempDir(), "viewmd-export.pdf")
	err := RenderMarkdownPDF("# 标题\n\n这是一段中文 Markdown 内容。\n\n- 项目一\n- 项目二\n\n```go\nfmt.Println(\"hello\")\n```\n", "", target)
	if err != nil {
		if strings.Contains(err.Error(), "no suitable TrueType font") {
			t.Skip(err)
		}
		t.Fatal(err)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 1024 {
		t.Fatalf("pdf is unexpectedly small: %d bytes", len(data))
	}
	if string(data[:4]) != "%PDF" {
		t.Fatalf("unexpected pdf header: %q", string(data[:4]))
	}
}
