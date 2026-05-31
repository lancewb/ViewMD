package exporter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderMarkdownPDF(t *testing.T) {
	target := filepath.Join(t.TempDir(), "viewmd-export.pdf")
	markdown := "# \u6807\u9898\n\n\u8fd9\u662f **\u52a0\u7c97\u4e2d\u6587** \u548c `\u884c\u5185\u4ee3\u7801\u4e2d\u6587` \u3002\n\n- \u65e0\u5e8f\u5217\u8868\n* \u661f\u53f7\u5217\u8868\n1. \u6709\u5e8f\u5217\u8868\n2. \u7b2c\u4e8c\u9879\n\n```go\nfmt.Println(\"\u4ee3\u7801\u5757\u4e2d\u6587\")\n```\n"
	err := RenderMarkdownPDF(markdown, "", target)
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
