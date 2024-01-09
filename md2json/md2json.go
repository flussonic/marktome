package md2json

import (
	"bytes"
	"fmt"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// NodeData - структура для хранения данных узла Markdown
type NodeData struct {
	Type       string            `json:"type"`
	Children   []NodeData        `json:"children,omitempty"`
	Literal    string            `json:"text,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type VisitFunc func(n ast.Node, d *NodeData, source []byte)

var register = map[ast.NodeKind]VisitFunc{
	ast.KindDocument:        visitDocument,
	ast.KindHeading:         visitHeading,
	ast.KindBlockquote:      visitBlockquote,
	ast.KindCodeBlock:       visitCodeBlock,
	ast.KindFencedCodeBlock: visitFencedCodeBlock,
	ast.KindHTMLBlock:       visitHTMLBlock,
	ast.KindList:            visitList,
	ast.KindListItem:        visitListItem,
	ast.KindParagraph:       visitParagraph,
	ast.KindTextBlock:       visitTextBlock,
	ast.KindThematicBreak:   visitThematicBreak,
	ast.KindAutoLink:        visitAutoLink,
	ast.KindCodeSpan:        visitCodeSpan,
	ast.KindEmphasis:        visitEmphasis,
	ast.KindImage:           visitImage,
	ast.KindLink:            visitLink,
	ast.KindRawHTML:         visitRawHTML,
	ast.KindText:            visitText,
	ast.KindString:          visitString,
}

func visitDocument(n ast.Node, d *NodeData, source []byte) {
	meta := n.(*ast.Document).Meta()
	if meta != nil {
		for k, v := range meta {
			d.Attributes[k] = fmt.Sprintf("%v", v)
		}
	}
}

func visitHeading(n ast.Node, d *NodeData, source []byte) {
	// data.Type = "heading"
	d.Attributes["level"] = fmt.Sprintf("%v", n.(*ast.Heading).Level)
}

func visitBlockquote(n ast.Node, d *NodeData, source []byte) {
}

func visitCodeBlock(n ast.Node, d *NodeData, source []byte) {
	visitHTMLBlock(n, d, source)
}

func visitFencedCodeBlock(n ast.Node, d *NodeData, source []byte) {
	// data.Type = "code"
	if n.(*ast.FencedCodeBlock).Language(source) != nil {
		d.Attributes["lang"] = string(n.(*ast.FencedCodeBlock).Language(source))
	}
	visitHTMLBlock(n, d, source)
}
func visitHTMLBlock(n ast.Node, d *NodeData, source []byte) {
	var b bytes.Buffer
	for i := 0; i < n.Lines().Len(); i++ {
		s := n.Lines().At(i)
		b.WriteString(string(source[s.Start:s.Stop]))
	}
	d.Literal = b.String()
}
func visitList(n ast.Node, d *NodeData, source []byte) {
	// data.Type = "list"
	if n.(*ast.List).IsOrdered() {
		d.Attributes["ordered"] = "true"
		if n.(*ast.List).Start > 0 {
			d.Attributes["start"] = string(n.(*ast.List).Start)
		}
	}
}
func visitListItem(n ast.Node, d *NodeData, source []byte) {
}
func visitParagraph(n ast.Node, d *NodeData, source []byte) {
}
func visitTextBlock(n ast.Node, d *NodeData, source []byte) {
}
func visitThematicBreak(n ast.Node, d *NodeData, source []byte) {
}
func visitAutoLink(n ast.Node, d *NodeData, source []byte) {
}
func visitCodeSpan(n ast.Node, d *NodeData, source []byte) {
}
func visitEmphasis(n ast.Node, d *NodeData, source []byte) {
}
func visitImage(n ast.Node, d *NodeData, source []byte) {
}
func visitLink(n ast.Node, d *NodeData, source []byte) {
	href := string(n.(*ast.Link).Destination)
	if len(href) > 0 {
		d.Attributes["href"] = href
	}
	title := string(n.(*ast.Link).Title)
	if len(title) > 0 {
		d.Attributes["title"] = title
	}
}
func visitRawHTML(n ast.Node, d *NodeData, source []byte) {
	// data.Type = "html"
	l := n.(*ast.RawHTML).Segments.Len()
	var b bytes.Buffer
	for i := 0; i < l; i++ {
		segment := n.(*ast.RawHTML).Segments.At(i)
		b.WriteString(string(segment.Value(source)))
	}
	d.Literal = b.String()
}
func visitText(n ast.Node, d *NodeData, source []byte) {
	// 	// data.Type = "text"
	d.Literal = string(n.Text(source))
}
func visitString(n ast.Node, d *NodeData, source []byte) {
	visitText(n, d, source)
}

func Convert(n ast.Node, source []byte) NodeData {
	data := NodeData{
		Type:       n.Kind().String(),
		Attributes: map[string]string{},
	}

	if n.Attributes() != nil {
		for i, _ := range n.Attributes() {
			data.Attributes[string(n.Attributes()[i].Name)] = fmt.Sprintf("%s", n.Attributes()[i].Value)
		}
	}
	f, ok := register[n.Kind()]
	if ok {
		f(n, &data, source)
		// } else {
		// 	fmt.Println("doesnt have function for", n.Kind().String())
	}
	if n.HasChildren() {
		children := []NodeData{}
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			children = append(children, Convert(child, source))
		}
		data.Children = children
	}
	return data
}

func Parse(source []byte) NodeData {
	blockParsers := []util.PrioritizedValue{
		util.Prioritized(parser.NewSetextHeadingParser(), 100),
		util.Prioritized(parser.NewThematicBreakParser(), 200),
		util.Prioritized(parser.NewListParser(), 300),
		util.Prioritized(parser.NewListItemParser(), 400),
		util.Prioritized(parser.NewCodeBlockParser(), 500),
		util.Prioritized(parser.NewATXHeadingParser(), 600),
		util.Prioritized(parser.NewFencedCodeBlockParser(), 700),
		util.Prioritized(parser.NewBlockquoteParser(), 800),
		util.Prioritized(NewHTMLBlockParser(), 900),
		util.Prioritized(parser.NewParagraphParser(), 1000),
	}
	p := parser.NewParser(parser.WithBlockParsers(blockParsers...),
		parser.WithInlineParsers(parser.DefaultInlineParsers()...),
		parser.WithParagraphTransformers(parser.DefaultParagraphTransformers()...),
	)
	p.AddOptions(parser.WithAttribute())

	md := goldmark.New(
		goldmark.WithParser(p),
		goldmark.WithExtensions(
			extension.GFM,
			meta.New(meta.WithStoresInDocument()),
			extension.NewTable(
				extension.WithTableCellAlignMethod(extension.TableCellAlignDefault),
			),
			Superlinks,
		),
		goldmark.WithParserOptions(
			parser.WithAttribute(),
		),
	)
	reader := text.NewReader(source)
	// fmt.Println("conf:", md.Parser().Config)
	document := md.Parser().Parse(reader)
	root := Convert(document, source)
	return root
}
