package exporter

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode"

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
	monoFont     string
	codeCJKFont  string
	contentWidth float64
	projectPath  string
}

type pdfInlineSegment struct {
	Text   string
	Bold   bool
	Italic bool
	Code   bool
}

type pdfInlineAtom struct {
	Text    string
	Segment pdfInlineSegment
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
	boldFontBytes := fontBytes
	if data, err := readBoldFont(); err == nil {
		boldFontBytes = data
	}
	monoFontBytes := fontBytes
	if data, err := readMonoFont(); err == nil {
		monoFontBytes = data
	}
	codeCJKFontBytes := fontBytes
	if data, err := readCodeCJKFont(); err == nil {
		codeCJKFontBytes = data
	}

	bodyFont := "ViewMDBody"
	monoFont := "ViewMDMono"
	codeCJKFont := "ViewMDCodeCJK"
	pdf.AddUTF8FontFromBytes(bodyFont, "", fontBytes)
	pdf.AddUTF8FontFromBytes(bodyFont, "B", boldFontBytes)
	pdf.AddUTF8FontFromBytes(bodyFont, "I", fontBytes)
	pdf.AddUTF8FontFromBytes(bodyFont, "BI", boldFontBytes)
	pdf.AddUTF8FontFromBytes(monoFont, "", monoFontBytes)
	pdf.AddUTF8FontFromBytes(codeCJKFont, "", codeCJKFontBytes)
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
		monoFont:     monoFont,
		codeCJKFont:  codeCJKFont,
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
	segments := r.inlineSegments(node)
	if len(segments) == 0 {
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
	r.renderInlineSegments(segments, size, lineHeight, "B", indent)

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

	segments := r.inlineSegments(node)
	if len(segments) == 0 {
		return
	}
	r.renderInlineSegments(segments, 11.5, 6.4, "", indent)
	r.pdf.Ln(2)
}

func (r *markdownPDFRenderer) renderList(node *goldast.List, indent float64) {
	number := node.Start
	if number == 0 {
		number = 1
	}

	for item := node.FirstChild(); item != nil; item = item.NextSibling() {
		marker := "\u00b7"
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
			segments := r.inlineSegments(paragraph)
			if len(segments) > 0 {
				r.renderMarkedInline(marker, segments, indent)
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
	markerWidth := r.listMarkerWidth(marker)
	r.pdf.SetFont(r.bodyFont, "", 11.5)
	r.pdf.SetTextColor(32, 35, 33)
	r.pdf.SetX(pdfMarginLeft + indent)
	r.pdf.CellFormat(markerWidth, 6.4, marker, "", 0, "L", false, 0, "")
	r.pdf.MultiCell(r.contentWidth-indent-markerWidth, 6.4, text, "", "L", false)
}

func (r *markdownPDFRenderer) renderMarkedInline(marker string, segments []pdfInlineSegment, indent float64) {
	markerWidth := r.listMarkerWidth(marker)
	r.pdf.SetFont(r.bodyFont, "", 11.5)
	r.pdf.SetTextColor(32, 35, 33)
	r.pdf.SetX(pdfMarginLeft + indent)
	r.pdf.CellFormat(markerWidth, 6.4, marker, "", 0, "L", false, 0, "")
	r.renderInlineSegmentsAt(segments, 11.5, 6.4, "", pdfMarginLeft+indent+markerWidth, r.contentWidth-indent-markerWidth)
}

func (r *markdownPDFRenderer) listMarkerWidth(marker string) float64 {
	r.pdf.SetFont(r.bodyFont, "", 11.5)
	width := r.pdf.GetStringWidth(marker) + 3
	if width < 8 {
		return 8
	}
	return width
}

func (r *markdownPDFRenderer) renderCodeBlock(code string, indent float64) {
	code = strings.TrimRight(code, "\r\n")
	if code == "" {
		return
	}

	blockWidth := r.contentWidth - indent
	paddingX := 4.0
	paddingY := 3.2
	lineHeight := 5.2
	fontSize := 9.4
	textWidth := blockWidth - paddingX*2

	wrappedLines := make([][]pdfInlineAtom, 0)
	for _, line := range strings.Split(strings.ReplaceAll(code, "\r\n", "\n"), "\n") {
		wrappedLines = append(wrappedLines, r.wrapCodeLine(line, textWidth, fontSize)...)
	}

	x := pdfMarginLeft + indent
	for len(wrappedLines) > 0 {
		_, pageHeight := r.pdf.GetPageSize()
		availableHeight := pageHeight - pdfMarginBottom - r.pdf.GetY()
		maxLines := int((availableHeight - paddingY*2) / lineHeight)
		if maxLines < 1 {
			r.pdf.AddPage()
			continue
		}
		if maxLines > len(wrappedLines) {
			maxLines = len(wrappedLines)
		}

		chunk := wrappedLines[:maxLines]
		blockHeight := float64(len(chunk))*lineHeight + paddingY*2
		y := r.pdf.GetY()

		r.pdf.SetFillColor(17, 23, 27)
		r.pdf.RoundedRect(x, y, blockWidth, blockHeight, 2, "1234", "F")
		r.pdf.SetTextColor(237, 242, 234)

		textY := y + paddingY
		for _, line := range chunk {
			r.drawCodeLine(line, x+paddingX, textY, lineHeight, fontSize)
			textY += lineHeight
		}

		r.pdf.SetY(y + blockHeight)
		wrappedLines = wrappedLines[maxLines:]
		if len(wrappedLines) > 0 {
			r.pdf.AddPage()
		}
	}

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

func (r *markdownPDFRenderer) inlineSegments(node goldast.Node) []pdfInlineSegment {
	segments := make([]pdfInlineSegment, 0)

	var visit func(goldast.Node, pdfInlineSegment)
	visit = func(current goldast.Node, style pdfInlineSegment) {
		switch n := current.(type) {
		case *goldast.Text:
			text := string(n.Text(r.source))
			if n.HardLineBreak() {
				text += "\n"
			} else if n.SoftLineBreak() {
				text += " "
			}
			appendInlineSegment(&segments, pdfInlineSegment{
				Text:   text,
				Bold:   style.Bold,
				Italic: style.Italic,
				Code:   style.Code,
			})
			return
		case *goldast.String:
			appendInlineSegment(&segments, pdfInlineSegment{
				Text:   string(n.Text(r.source)),
				Bold:   style.Bold,
				Italic: style.Italic,
				Code:   style.Code,
			})
			return
		case *goldast.Emphasis:
			next := style
			if n.Level >= 2 {
				next.Bold = true
			} else {
				next.Italic = true
			}
			for child := n.FirstChild(); child != nil; child = child.NextSibling() {
				visit(child, next)
			}
			return
		case *goldast.CodeSpan:
			text := strings.ReplaceAll(r.inlineText(n), "\n", " ")
			appendInlineSegment(&segments, pdfInlineSegment{
				Text: text,
				Code: true,
			})
			return
		}

		for child := current.FirstChild(); child != nil; child = child.NextSibling() {
			visit(child, style)
		}
	}

	visit(node, pdfInlineSegment{})
	return trimInlineSegments(segments)
}

func (r *markdownPDFRenderer) renderInlineSegments(segments []pdfInlineSegment, size float64, lineHeight float64, baseStyle string, indent float64) {
	r.renderInlineSegmentsAt(segments, size, lineHeight, baseStyle, pdfMarginLeft+indent, r.contentWidth-indent)
}

func (r *markdownPDFRenderer) renderInlineSegmentsAt(segments []pdfInlineSegment, size float64, lineHeight float64, baseStyle string, leftX float64, width float64) {
	atoms := inlineAtoms(segments)
	if len(atoms) == 0 || width <= 0 {
		return
	}

	x := leftX
	y := r.pdf.GetY()
	rightX := leftX + width
	drew := false

	for _, atom := range atoms {
		if atom.Text == "\n" {
			y = r.nextInlineLine(y, lineHeight)
			x = leftX
			continue
		}

		parts := []pdfInlineAtom{atom}
		if r.inlineAtomWidth(atom, size, baseStyle) > width && len([]rune(atom.Text)) > 1 {
			parts = splitInlineAtom(atom)
		}

		for _, part := range parts {
			if part.Text == " " && nearSameX(x, leftX) {
				continue
			}

			atomWidth := r.inlineAtomWidth(part, size, baseStyle)
			if !nearSameX(x, leftX) && x+atomWidth > rightX {
				y = r.nextInlineLine(y, lineHeight)
				x = leftX
				if part.Text == " " {
					continue
				}
			}

			r.ensureInlineSpace(&y, lineHeight)
			r.drawInlineAtom(part, x, y, atomWidth, lineHeight, size, baseStyle)
			x += atomWidth
			drew = true
		}
	}

	if drew {
		r.pdf.SetXY(leftX, y+lineHeight)
		r.pdf.SetFont(r.bodyFont, "", 11.5)
		r.pdf.SetTextColor(32, 35, 33)
		r.pdf.SetFillColor(255, 255, 255)
	}
}

func (r *markdownPDFRenderer) nextInlineLine(y float64, lineHeight float64) float64 {
	y += lineHeight
	r.ensureInlineSpace(&y, lineHeight)
	return y
}

func (r *markdownPDFRenderer) ensureInlineSpace(y *float64, lineHeight float64) {
	_, pageHeight := r.pdf.GetPageSize()
	if *y+lineHeight > pageHeight-pdfMarginBottom {
		r.pdf.AddPage()
		*y = r.pdf.GetY()
	}
}

func (r *markdownPDFRenderer) drawInlineAtom(atom pdfInlineAtom, x float64, y float64, width float64, lineHeight float64, size float64, baseStyle string) {
	if atom.Segment.Code {
		r.setCodeFont(atom.Text, inlineCodeFontSize(size))
		r.pdf.SetXY(x, y)
		paddingX := 1.25
		textWidth := width - paddingX*2
		r.pdf.SetFillColor(245, 229, 221)
		r.pdf.RoundedRect(x, y+0.7, width, lineHeight-1.15, 1.2, "1234", "F")
		r.pdf.SetTextColor(141, 57, 34)
		r.pdf.SetXY(x+paddingX, y)
		r.pdf.CellFormat(textWidth, lineHeight, atom.Text, "", 0, "L", false, 0, "")
		return
	}

	r.setInlineFont(atom.Segment, size, baseStyle)
	r.pdf.SetXY(x, y)
	r.pdf.SetTextColor(32, 35, 33)
	r.pdf.CellFormat(width, lineHeight, atom.Text, "", 0, "L", false, 0, "")
}

func (r *markdownPDFRenderer) inlineAtomWidth(atom pdfInlineAtom, size float64, baseStyle string) float64 {
	if atom.Segment.Code {
		r.setCodeFont(atom.Text, inlineCodeFontSize(size))
		return r.pdf.GetStringWidth(atom.Text) + 2.5
	}

	r.setInlineFont(atom.Segment, size, baseStyle)
	return r.pdf.GetStringWidth(atom.Text)
}

func (r *markdownPDFRenderer) setInlineFont(segment pdfInlineSegment, size float64, baseStyle string) {
	if segment.Code {
		r.setCodeFont(segment.Text, inlineCodeFontSize(size))
		return
	}

	bold := segment.Bold || strings.Contains(baseStyle, "B")
	italic := segment.Italic || strings.Contains(baseStyle, "I")
	style := ""
	if bold {
		style += "B"
	}
	if italic {
		style += "I"
	}
	r.pdf.SetFont(r.bodyFont, style, size)
}

func inlineCodeFontSize(size float64) float64 {
	codeSize := size * 0.92
	if codeSize < 8 {
		return 8
	}
	return codeSize
}

func (r *markdownPDFRenderer) setCodeFont(text string, size float64) {
	if needsCJKFont(text) {
		r.pdf.SetFont(r.codeCJKFont, "", size)
		return
	}
	r.pdf.SetFont(r.monoFont, "", size)
}

func (r *markdownPDFRenderer) wrapCodeLine(line string, width float64, size float64) [][]pdfInlineAtom {
	atoms := codeAtoms(strings.ReplaceAll(line, "\t", "    "))
	if len(atoms) == 0 {
		return [][]pdfInlineAtom{{}}
	}

	lines := make([][]pdfInlineAtom, 0)
	current := make([]pdfInlineAtom, 0)
	currentWidth := 0.0

	for _, atom := range atoms {
		parts := []pdfInlineAtom{atom}
		if r.codeAtomWidth(atom, size) > width && len([]rune(atom.Text)) > 1 {
			parts = splitInlineAtom(atom)
		}

		for _, part := range parts {
			partWidth := r.codeAtomWidth(part, size)
			if len(current) > 0 && currentWidth+partWidth > width {
				lines = append(lines, current)
				current = make([]pdfInlineAtom, 0)
				currentWidth = 0
			}
			current = append(current, part)
			currentWidth += partWidth
		}
	}

	if len(current) > 0 {
		lines = append(lines, current)
	}
	return lines
}

func (r *markdownPDFRenderer) drawCodeLine(atoms []pdfInlineAtom, x float64, y float64, lineHeight float64, size float64) {
	currentX := x
	for _, atom := range atoms {
		width := r.codeAtomWidth(atom, size)
		r.setCodeFont(atom.Text, size)
		r.pdf.SetTextColor(237, 242, 234)
		r.pdf.SetXY(currentX, y)
		r.pdf.CellFormat(width, lineHeight, atom.Text, "", 0, "L", false, 0, "")
		currentX += width
	}
}

func (r *markdownPDFRenderer) codeAtomWidth(atom pdfInlineAtom, size float64) float64 {
	r.setCodeFont(atom.Text, size)
	return r.pdf.GetStringWidth(atom.Text)
}

func appendInlineSegment(segments *[]pdfInlineSegment, segment pdfInlineSegment) {
	if segment.Text == "" {
		return
	}

	lastIndex := len(*segments) - 1
	if lastIndex >= 0 && sameInlineStyle((*segments)[lastIndex], segment) {
		(*segments)[lastIndex].Text += segment.Text
		return
	}
	*segments = append(*segments, segment)
}

func sameInlineStyle(left pdfInlineSegment, right pdfInlineSegment) bool {
	return left.Bold == right.Bold && left.Italic == right.Italic && left.Code == right.Code
}

func trimInlineSegments(segments []pdfInlineSegment) []pdfInlineSegment {
	for len(segments) > 0 && !segments[0].Code {
		segments[0].Text = strings.TrimLeftFunc(segments[0].Text, unicode.IsSpace)
		if segments[0].Text != "" {
			break
		}
		segments = segments[1:]
	}

	for len(segments) > 0 {
		last := len(segments) - 1
		if segments[last].Code {
			break
		}
		segments[last].Text = strings.TrimRightFunc(segments[last].Text, unicode.IsSpace)
		if segments[last].Text != "" {
			break
		}
		segments = segments[:last]
	}

	return segments
}

func inlineAtoms(segments []pdfInlineSegment) []pdfInlineAtom {
	atoms := make([]pdfInlineAtom, 0)
	for _, segment := range segments {
		if segment.Text == "" {
			continue
		}
		if segment.Code {
			appendCodeAtoms(&atoms, segment.Text)
			continue
		}
		appendTextAtoms(&atoms, segment)
	}
	return collapseSpaceAtoms(atoms)
}

func appendCodeAtoms(atoms *[]pdfInlineAtom, text string) {
	*atoms = append(*atoms, codeAtoms(text)...)
}

func codeAtoms(text string) []pdfInlineAtom {
	atoms := make([]pdfInlineAtom, 0)
	var builder strings.Builder
	currentNeedsCJK := false
	hasCurrentStyle := false

	flush := func() {
		if builder.Len() == 0 {
			return
		}
		value := builder.String()
		atoms = append(atoms, pdfInlineAtom{
			Text: value,
			Segment: pdfInlineSegment{
				Text: value,
				Code: true,
			},
		})
		builder.Reset()
	}

	for _, char := range text {
		needsCJK := needsCJKRune(char)
		if hasCurrentStyle && needsCJK != currentNeedsCJK {
			flush()
		}
		builder.WriteRune(char)
		currentNeedsCJK = needsCJK
		hasCurrentStyle = true
	}
	flush()
	return atoms
}

func appendTextAtoms(atoms *[]pdfInlineAtom, segment pdfInlineSegment) {
	var builder strings.Builder
	flush := func() {
		if builder.Len() == 0 {
			return
		}
		*atoms = append(*atoms, pdfInlineAtom{Text: builder.String(), Segment: segment})
		builder.Reset()
	}

	for _, char := range strings.ReplaceAll(segment.Text, "\r\n", "\n") {
		switch {
		case char == '\r':
			continue
		case char == '\n':
			flush()
			*atoms = append(*atoms, pdfInlineAtom{Text: "\n", Segment: segment})
		case unicode.IsSpace(char):
			flush()
			*atoms = append(*atoms, pdfInlineAtom{Text: " ", Segment: segment})
		case char < 128:
			builder.WriteRune(char)
		default:
			flush()
			*atoms = append(*atoms, pdfInlineAtom{Text: string(char), Segment: segment})
		}
	}
	flush()
}

func collapseSpaceAtoms(atoms []pdfInlineAtom) []pdfInlineAtom {
	collapsed := make([]pdfInlineAtom, 0, len(atoms))
	for _, atom := range atoms {
		if !atom.Segment.Code && atom.Text == " " {
			if len(collapsed) == 0 || collapsed[len(collapsed)-1].Text == " " || collapsed[len(collapsed)-1].Text == "\n" {
				continue
			}
		}
		if !atom.Segment.Code && atom.Text == "\n" && len(collapsed) > 0 && collapsed[len(collapsed)-1].Text == " " {
			collapsed = collapsed[:len(collapsed)-1]
		}
		collapsed = append(collapsed, atom)
	}
	return collapsed
}

func needsCJKFont(text string) bool {
	for _, char := range text {
		if needsCJKRune(char) {
			return true
		}
	}
	return false
}

func needsCJKRune(char rune) bool {
	return char > unicode.MaxASCII
}

func splitInlineAtom(atom pdfInlineAtom) []pdfInlineAtom {
	parts := make([]pdfInlineAtom, 0, len([]rune(atom.Text)))
	for _, char := range atom.Text {
		parts = append(parts, pdfInlineAtom{Text: string(char), Segment: atom.Segment})
	}
	return parts
}

func nearSameX(left float64, right float64) bool {
	if left > right {
		return left-right < 0.01
	}
	return right-left < 0.01
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
	return readFirstFont(bodyFontCandidates(), "no suitable TrueType font was found for PDF export")
}

func readBoldFont() ([]byte, error) {
	return readFirstFont(boldFontCandidates(), "no suitable bold TrueType font was found for PDF export")
}

func readMonoFont() ([]byte, error) {
	return readFirstFont(monoFontCandidates(), "no suitable monospace TrueType font was found for PDF export")
}

func readCodeCJKFont() ([]byte, error) {
	return readFirstFont(codeCJKFontCandidates(), "no suitable CJK code TrueType font was found for PDF export")
}

func readFirstFont(candidates []string, message string) ([]byte, error) {
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		data, err := os.ReadFile(candidate)
		if err == nil && len(data) > 0 {
			return data, nil
		}
	}
	return nil, errors.New(message)
}

func bodyFontCandidates() []string {
	windowsFonts := windowsFontDir()

	return []string{
		filepath.Join(windowsFonts, "STSONG.TTF"),
		filepath.Join(windowsFonts, "NotoSerifSC-VF.ttf"),
		filepath.Join(windowsFonts, "SimsunExtG.ttf"),
		filepath.Join(windowsFonts, "Deng.ttf"),
		filepath.Join(windowsFonts, "simhei.ttf"),
		filepath.Join(windowsFonts, "NotoSansSC-VF.ttf"),
		filepath.Join(windowsFonts, "malgun.ttf"),
		filepath.Join(windowsFonts, "arial.ttf"),
	}
}

func boldFontCandidates() []string {
	windowsFonts := windowsFontDir()

	return []string{
		filepath.Join(windowsFonts, "Dengb.ttf"),
		filepath.Join(windowsFonts, "simhei.ttf"),
		filepath.Join(windowsFonts, "NotoSansSC-VF.ttf"),
		filepath.Join(windowsFonts, "NotoSerifSC-VF.ttf"),
		filepath.Join(windowsFonts, "simsunb.ttf"),
		filepath.Join(windowsFonts, "arialbd.ttf"),
	}
}

func monoFontCandidates() []string {
	windowsFonts := windowsFontDir()

	return []string{
		filepath.Join(windowsFonts, "CascadiaCode.ttf"),
		filepath.Join(windowsFonts, "CascadiaMono.ttf"),
		filepath.Join(windowsFonts, "consola.ttf"),
		filepath.Join(windowsFonts, "cour.ttf"),
	}
}

func codeCJKFontCandidates() []string {
	windowsFonts := windowsFontDir()

	return []string{
		filepath.Join(windowsFonts, "Deng.ttf"),
		filepath.Join(windowsFonts, "NotoSansSC-VF.ttf"),
		filepath.Join(windowsFonts, "simhei.ttf"),
		filepath.Join(windowsFonts, "STSONG.TTF"),
	}
}

func windowsFontDir() string {
	if windir := os.Getenv("WINDIR"); windir != "" {
		return filepath.Join(windir, "Fonts")
	}
	return `C:\Windows\Fonts`
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
