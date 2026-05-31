package exporter

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"github.com/yuin/goldmark"
	goldast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	goldtext "github.com/yuin/goldmark/text"
)

const (
	pdfMarginLeft   = 18.0
	pdfMarginTop    = 18.0
	pdfMarginRight  = 18.0
	pdfMarginBottom = 20.0
)

type markdownPDFRenderer struct {
	pdf          *gofpdf.Fpdf
	source       []byte
	bodyFont     string
	contentWidth float64
	projectPath  string
}

func RenderMarkdownPDF(markdownText string, projectPath string, targetPath string) error {
	renderer, err := newMarkdownPDFRenderer(markdownText, projectPath)
	if err != nil {
		return err
	}

	if err := renderer.render(); err != nil {
		return err
	}

	return renderer.pdf.OutputFileAndClose(targetPath)
}

func newMarkdownPDFRenderer(markdownText string, projectPath string) (*markdownPDFRenderer, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(pdfMarginLeft, pdfMarginTop, pdfMarginRight)
	pdf.SetAutoPageBreak(true, pdfMarginBottom)
	pdf.SetTitle("ViewMD Export", true)
	pdf.SetCreator("ViewMD", true)

	fontBytes, err := readBodyFont()
	if err != nil {
		return nil, err
	}

	bodyFont := "ViewMDBody"
	pdf.AddUTF8FontFromBytes(bodyFont, "", fontBytes)
	pdf.AddUTF8FontFromBytes(bodyFont, "B", fontBytes)
	pdf.AddUTF8FontFromBytes(bodyFont, "I", fontBytes)
	pdf.AddUTF8FontFromBytes(bodyFont, "BI", fontBytes)
	if err := pdf.Error(); err != nil {
		return nil, err
	}

	pdf.AliasNbPages("")
	pdf.SetFooterFunc(func() {
		pdf.SetY(-13)
		pdf.SetFont(bodyFont, "", 8.5)
		pdf.SetTextColor(118, 128, 119)
		pdf.CellFormat(0, 5, fmt.Sprintf("%d / {nb}", pdf.PageNo()), "", 0, "C", false, 0, "")
	})

	pdf.AddPage()
	pdf.SetFont(bodyFont, "", 11.5)
	pdf.SetTextColor(32, 35, 33)

	return &markdownPDFRenderer{
		pdf:          pdf,
		source:       []byte(markdownText),
		bodyFont:     bodyFont,
		contentWidth: 210 - pdfMarginLeft - pdfMarginRight,
		projectPath:  projectPath,
	}, nil
}

func (r *markdownPDFRenderer) render() error {
	md := goldmark.New(goldmark.WithExtensions(extension.GFM))
	document := md.Parser().Parse(goldtext.NewReader(r.source))

	for node := document.FirstChild(); node != nil; node = node.NextSibling() {
		r.renderBlock(node, 0)
	}

	return r.pdf.Error()
}

func (r *markdownPDFRenderer) renderBlock(node goldast.Node, indent float64) {
	switch n := node.(type) {
	case *goldast.Heading:
		r.renderHeading(n, indent)
	case *goldast.Paragraph:
		r.renderParagraph(n, indent)
	case *goldast.List:
		r.renderList(n, indent)
	case *goldast.FencedCodeBlock:
		r.renderCodeBlock(linesText(n.Lines(), r.source), indent)
	case *goldast.CodeBlock:
		r.renderCodeBlock(linesText(n.Lines(), r.source), indent)
	case *goldast.Blockquote:
		r.renderBlockquote(n, indent)
	case *goldast.ThematicBreak:
		r.renderRule(indent)
	case *extast.Table:
		r.renderTable(n, indent)
	default:
		text := r.inlineText(node)
		if text != "" {
			r.renderText(text, 11.5, 6.4, "", indent, 32, 35, 33)
		}
	}
}

func (r *markdownPDFRenderer) renderHeading(node *goldast.Heading, indent float64) {
	text := r.inlineText(node)
	if text == "" {
		return
	}

	size := 15.0
	lineHeight := 7.0
	spaceBefore := 4.0
	spaceAfter := 2.2

	switch node.Level {
	case 1:
		size = 24
		lineHeight = 10
		spaceBefore = 0
		spaceAfter = 4
	case 2:
		size = 18
		lineHeight = 8
	case 3:
		size = 14
		lineHeight = 6.8
	}

	if r.pdf.GetY() > pdfMarginTop+1 && spaceBefore > 0 {
		r.pdf.Ln(spaceBefore)
	}
	r.renderText(text, size, lineHeight, "B", indent, 17, 23, 27)

	if node.Level == 1 {
		x := pdfMarginLeft + indent
		y := r.pdf.GetY()
		r.pdf.SetDrawColor(222, 212, 194)
		r.pdf.Line(x, y, x+r.contentWidth-indent, y)
		r.pdf.Ln(spaceAfter)
	} else {
		r.pdf.Ln(spaceAfter)
	}
}

func (r *markdownPDFRenderer) renderParagraph(node *goldast.Paragraph, indent float64) {
	if image := paragraphOnlyImage(node); image != nil {
		if r.renderImage(image, indent) {
			return
		}
	}

	text := r.inlineText(node)
	if text == "" {
		return
	}
	r.renderText(text, 11.5, 6.4, "", indent, 32, 35, 33)
	r.pdf.Ln(2)
}

func (r *markdownPDFRenderer) renderList(node *goldast.List, indent float64) {
	number := node.Start
	if number == 0 {
		number = 1
	}

	for item := node.FirstChild(); item != nil; item = item.NextSibling() {
		marker := "•"
		if node.IsOrdered() {
			marker = fmt.Sprintf("%d.", number)
			number++
		}
		r.renderListItem(marker, item, indent)
	}
	r.pdf.Ln(1)
}

func (r *markdownPDFRenderer) renderListItem(marker string, item goldast.Node, indent float64) {
	markerUsed := false
	for child := item.FirstChild(); child != nil; child = child.NextSibling() {
		if paragraph, ok := child.(*goldast.Paragraph); ok && !markerUsed {
			text := r.inlineText(paragraph)
			if text != "" {
				r.renderMarkedText(marker, text, indent)
				markerUsed = true
			}
			continue
		}
		r.renderBlock(child, indent+8)
	}

	if !markerUsed && item.ChildCount() == 0 {
		r.renderMarkedText(marker, "", indent)
	}
}

func (r *markdownPDFRenderer) renderMarkedText(marker string, text string, indent float64) {
	r.pdf.SetFont(r.bodyFont, "", 11.5)
	r.pdf.SetTextColor(32, 35, 33)
	r.pdf.SetX(pdfMarginLeft + indent)
	r.pdf.CellFormat(8, 6.4, marker, "", 0, "L", false, 0, "")
	r.pdf.MultiCell(r.contentWidth-indent-8, 6.4, text, "", "L", false)
}

func (r *markdownPDFRenderer) renderCodeBlock(code string, indent float64) {
	code = strings.TrimRight(code, "\r\n")
	if code == "" {
		return
	}

	r.pdf.SetFont(r.bodyFont, "", 9.7)
	r.pdf.SetTextColor(237, 242, 234)
	r.pdf.SetFillColor(17, 23, 27)
	r.pdf.SetX(pdfMarginLeft + indent)
	r.pdf.MultiCell(r.contentWidth-indent, 5.2, code, "", "L", true)
	r.pdf.Ln(3)
	r.pdf.SetTextColor(32, 35, 33)
	r.pdf.SetFillColor(255, 255, 255)
}

func (r *markdownPDFRenderer) renderBlockquote(node *goldast.Blockquote, indent float64) {
	text := r.inlineText(node)
	if text == "" {
		return
	}

	r.pdf.SetFont(r.bodyFont, "I", 11.2)
	r.pdf.SetTextColor(79, 89, 79)
	r.pdf.SetDrawColor(213, 154, 45)
	r.pdf.SetFillColor(251, 243, 223)
	r.pdf.SetX(pdfMarginLeft + indent)
	r.pdf.MultiCell(r.contentWidth-indent, 6.2, text, "L", "L", true)
	r.pdf.Ln(3)
	r.pdf.SetTextColor(32, 35, 33)
	r.pdf.SetFillColor(255, 255, 255)
}

func (r *markdownPDFRenderer) renderRule(indent float64) {
	x := pdfMarginLeft + indent
	y := r.pdf.GetY() + 2
	r.pdf.SetDrawColor(207, 196, 177)
	r.pdf.Line(x, y, x+r.contentWidth-indent, y)
	r.pdf.Ln(6)
}

func (r *markdownPDFRenderer) renderTable(node *extast.Table, indent float64) {
	rows := make([][]string, 0)
	for row := node.FirstChild(); row != nil; row = row.NextSibling() {
		cells := make([]string, 0)
		for cell := row.FirstChild(); cell != nil; cell = cell.NextSibling() {
			cells = append(cells, r.inlineText(cell))
		}
		if len(cells) > 0 {
			rows = append(rows, cells)
		}
	}

	maxColumns := 0
	for _, row := range rows {
		if len(row) > maxColumns {
			maxColumns = len(row)
		}
	}
	if maxColumns == 0 {
		return
	}

	tableWidth := r.contentWidth - indent
	columnWidth := tableWidth / float64(maxColumns)
	lineHeight := 5.2

	for rowIndex, row := range rows {
		header := rowIndex == 0
		style := ""
		if header {
			style = "B"
		}

		r.pdf.SetFont(r.bodyFont, style, 9.6)
		rowHeight := lineHeight + 4
		for _, cell := range row {
			lines := r.pdf.SplitText(cell, columnWidth-4)
			if height := float64(max(1, len(lines)))*lineHeight + 4; height > rowHeight {
				rowHeight = height
			}
		}

		r.ensureSpace(rowHeight + 2)
		y := r.pdf.GetY()
		x := pdfMarginLeft + indent

		for column := 0; column < maxColumns; column++ {
			cell := ""
			if column < len(row) {
				cell = row[column]
			}

			if header {
				r.pdf.SetFillColor(241, 236, 226)
			} else {
				r.pdf.SetFillColor(255, 253, 248)
			}
			r.pdf.SetDrawColor(222, 212, 194)
			r.pdf.Rect(x+float64(column)*columnWidth, y, columnWidth, rowHeight, "DF")
			r.pdf.SetXY(x+float64(column)*columnWidth+2, y+2)
			r.pdf.SetTextColor(32, 35, 33)
			r.pdf.MultiCell(columnWidth-4, lineHeight, cell, "", "L", false)
		}

		r.pdf.SetY(y + rowHeight)
	}
	r.pdf.Ln(4)
}

func (r *markdownPDFRenderer) renderText(text string, size float64, lineHeight float64, style string, indent float64, red int, green int, blue int) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}

	r.pdf.SetFont(r.bodyFont, style, size)
	r.pdf.SetTextColor(red, green, blue)
	r.pdf.SetX(pdfMarginLeft + indent)
	r.pdf.MultiCell(r.contentWidth-indent, lineHeight, text, "", "L", false)
}

func (r *markdownPDFRenderer) renderImage(node *goldast.Image, indent float64) bool {
	imagePath := r.resolveImagePath(string(node.Destination))
	if imagePath == "" {
		return false
	}

	info := r.pdf.RegisterImageOptions(imagePath, gofpdf.ImageOptions{ReadDpi: true})
	if r.pdf.Error() != nil || info == nil {
		return false
	}

	width, height := info.Extent()
	maxWidth := r.contentWidth - indent
	maxHeight := 120.0
	if width <= 0 || height <= 0 {
		return false
	}
	if width > maxWidth {
		scale := maxWidth / width
		width *= scale
		height *= scale
	}
	if height > maxHeight {
		scale := maxHeight / height
		width *= scale
		height *= scale
	}

	r.ensureSpace(height + 6)
	x := pdfMarginLeft + indent
	y := r.pdf.GetY()
	r.pdf.ImageOptions(imagePath, x, y, width, height, false, gofpdf.ImageOptions{ReadDpi: true}, 0, "")
	r.pdf.Ln(height + 4)
	return true
}

func (r *markdownPDFRenderer) resolveImagePath(destination string) string {
	if destination == "" {
		return ""
	}
	if strings.HasPrefix(destination, "http://") || strings.HasPrefix(destination, "https://") || strings.HasPrefix(destination, "data:") {
		return ""
	}
	clean, err := url.PathUnescape(destination)
	if err == nil {
		destination = clean
	}
	destination = strings.TrimPrefix(destination, "file:///")
	if filepath.IsAbs(destination) {
		if fileExists(destination) {
			return destination
		}
		return ""
	}
	if r.projectPath == "" {
		return ""
	}
	candidate := filepath.Join(r.projectPath, filepath.FromSlash(destination))
	if fileExists(candidate) {
		return candidate
	}
	return ""
}

func (r *markdownPDFRenderer) ensureSpace(height float64) {
	_, pageHeight := r.pdf.GetPageSize()
	if r.pdf.GetY()+height > pageHeight-pdfMarginBottom {
		r.pdf.AddPage()
	}
}

func (r *markdownPDFRenderer) inlineText(node goldast.Node) string {
	var builder strings.Builder

	_ = goldast.Walk(node, func(child goldast.Node, entering bool) (goldast.WalkStatus, error) {
		if !entering {
			return goldast.WalkContinue, nil
		}

		switch n := child.(type) {
		case *goldast.Text:
			builder.Write(n.Text(r.source))
			if n.HardLineBreak() {
				builder.WriteByte('\n')
			} else if n.SoftLineBreak() {
				builder.WriteByte(' ')
			}
		case *goldast.String:
			builder.Write(n.Text(r.source))
		}

		return goldast.WalkContinue, nil
	})

	return normalizeInlineText(builder.String())
}

func linesText(lines *goldtext.Segments, source []byte) string {
	var builder strings.Builder
	for index := 0; index < lines.Len(); index++ {
		segment := lines.At(index)
		builder.Write(segment.Value(source))
	}
	return builder.String()
}

func normalizeInlineText(text string) string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	for index, line := range lines {
		lines[index] = strings.Join(strings.Fields(line), " ")
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func readBodyFont() ([]byte, error) {
	for _, candidate := range bodyFontCandidates() {
		if candidate == "" {
			continue
		}
		data, err := os.ReadFile(candidate)
		if err == nil && len(data) > 0 {
			return data, nil
		}
	}
	return nil, errors.New("no suitable TrueType font was found for PDF export")
}

func bodyFontCandidates() []string {
	windowsFonts := filepath.Join(os.Getenv("WINDIR"), "Fonts")
	if os.Getenv("WINDIR") == "" {
		windowsFonts = `C:\Windows\Fonts`
	}

	return []string{
		filepath.Join(windowsFonts, "Deng.ttf"),
		filepath.Join(windowsFonts, "simhei.ttf"),
		filepath.Join(windowsFonts, "simsunb.ttf"),
		filepath.Join(windowsFonts, "NotoSansSC-VF.ttf"),
		filepath.Join(windowsFonts, "malgun.ttf"),
		filepath.Join(windowsFonts, "arial.ttf"),
	}
}

func paragraphOnlyImage(node *goldast.Paragraph) *goldast.Image {
	var image *goldast.Image
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		current, ok := child.(*goldast.Image)
		if !ok {
			return nil
		}
		if image != nil {
			return nil
		}
		image = current
	}
	return image
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func max(left int, right int) int {
	if left > right {
		return left
	}
	return right
}
