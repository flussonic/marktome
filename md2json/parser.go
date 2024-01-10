package md2json

import (
	"bytes"
	"fmt"
	"regexp"
)

type Kind string
type AttributeMap map[string]string

const (
	Document   Kind = "Document"
	Paragraph  Kind = "Paragraph"
	Text       Kind = "Text"
	Image      Kind = "Image"
	HTML       Kind = "HTML"
	Snippet    Kind = "Snippet"
	Heading    Kind = "Heading"
	List       Kind = "List"
	ListItem   Kind = "ListItem"
	Admonition Kind = "Admonition"
	Code       Kind = "Code"
	Bold       Kind = "Bold"
	Emphasis   Kind = "Emphasis"
)

type Node struct {
	Type       Kind         `json:"type"`
	Children   []Node       `json:"children,omitempty"`
	Literal    string       `json:"text,omitempty"`
	Attributes AttributeMap `json:"attributes,omitempty"`
}

func ParseDocument(source []byte) Node {
	st := ParserState{
		source: source,
	}
	node := parseDocument(&st)
	return node
}

func parseDocument(st *ParserState) Node {
	node := Node{
		Type:       Document,
		Attributes: map[string]string{},
		Children:   []Node{},
	}

	for !st.eof() {
		for !st.eof() && st.startsWith("\n") {
			st.consumeLine()
		}
		if st.eof() {
			break
		}

		if st.startsWith("---\n") {
			parseMeta(&node, st)
			continue
		}

		if st.startsWith("<snippet ") {
			node.Children = append(node.Children, parseSnippet(st))
			continue
		}

		if st.startsWith("#") {
			node.Children = append(node.Children, parseHeading(st))
			continue
		}

		if st.startsWith("* ") {
			node.Children = append(node.Children, parseList(st))
			continue
		}

		if st.startsWith("!!! ") {
			node.Children = append(node.Children, parseAdmonition(st))
			continue
		}

		node.Children = append(node.Children, parseParagraph(st))
	}
	return node
}

func parseParagraph(st *ParserState) Node {
	s1 := st.source
	for !st.eof() && !st.startsWith("\n") {
		st.consumeLine()
	}
	l := len(s1) - len(st.source)
	block := s1[:l]
	node := Node{
		Type:       Paragraph,
		Attributes: map[string]string{},
		Children:   parseText(block),
	}
	return node
}

type InlineParserState struct {
	source   []byte
	pos      int
	children []Node
}

func (st *InlineParserState) flushText() {
	if st.pos > 1 {
		node := Node{Type: Text, Literal: string(st.source[0:st.pos])}
		st.children = append(st.children, node)
		st.source = st.source[st.pos:]
		st.pos = 0
	}
}

func (st *InlineParserState) consumeN(n int) []byte {
	if st.pos+n > len(st.source) {
		n = len(st.source) - st.pos
	}
	st.pos += n
	s := st.source[0:st.pos]
	st.source = st.source[st.pos:]
	st.pos = 0
	return s
}

func (st *InlineParserState) startsWith(s []byte) bool {
	if st.pos+len(s) >= len(st.source) {
		return false
	}
	return bytes.Equal(s, st.source[st.pos:st.pos+len(s)])
}

func (st *InlineParserState) parseInliner(symbol []byte, typ Kind) {
	if !st.startsWith(symbol) {
		return
	}
	st.flushText()
	st.consumeN(len(symbol))
	i := bytes.Index(st.source, symbol)
	if i > 0 {
		st.children = append(st.children, Node{Type: typ, Literal: string(st.consumeN(i))})
		st.consumeN(len(symbol))
	}
}

var tagAttrsRegexp = regexp.MustCompile(`([a-z]+)="([^"]+)"`) //nolint:golint,lll

func (st *InlineParserState) parseHtml() {
	if st.startsWith([]byte{'<', '/'}) {
		return
	}
	if !st.startsWith([]byte{'<'}) {
		return
	}
	st.flushText()
	st.consumeN(1)
	nameEnd := bytes.Index(st.source, []byte{' '})
	if nameEnd < 0 {
		st.flushText()
		return
	}
	tagName := st.consumeN(nameEnd)
	attrs := map[string]string{}
	attrs["tag"] = string(tagName)

	tagEnd := bytes.Index(st.source, []byte{'>'})
	if tagEnd < 0 {
		return
	}
	tagAttrs := st.consumeN(tagEnd)
	st.consumeN(1) // wipe > symbol
	if st.startsWith([]byte{'\n'}) {
		st.consumeN(1)
	}

	attrMatches := tagAttrsRegexp.FindAllSubmatch(tagAttrs, -1)
	for i := range attrMatches {
		attrs[string(attrMatches[i][1])] = string(attrMatches[i][2])
	}
	closure := append([]byte{'<', '/'}, tagName...)
	closure = append(closure, '>')
	finish := bytes.Index(st.source, closure)
	if finish < 0 {
		return
	}
	text := string(st.consumeN(finish))
	st.consumeN(len(closure))
	node := Node{
		Type:       HTML,
		Attributes: attrs,
		Literal:    text,
	}
	st.children = append(st.children, node)
}

func (st *InlineParserState) parseImage() {
	if !st.startsWith([]byte{'!', '['}) {
		return
	}
	st.consumeN(2)
	titleEnd := bytes.Index(st.source, []byte{']', '('})
	if titleEnd < 0 {
		return
	}
	title := string(st.consumeN(titleEnd))
	st.consumeN(2)
	urlEnd := bytes.Index(st.source, []byte{')'})
	if urlEnd < 0 {
		return
	}
	url := string(st.consumeN(urlEnd))
	st.consumeN(1)
	node := Node{
		Type:       Image,
		Attributes: map[string]string{"title": title, "src": url},
	}
	st.children = append(st.children, node)
}

func parseText(source []byte) []Node {
	st := InlineParserState{
		source:   source,
		children: []Node{},
		pos:      0,
	}
	for ; st.pos < len(st.source); st.pos++ {
		st.parseInliner([]byte{'`'}, Code)
		st.parseInliner([]byte{'*', '*'}, Bold)
		st.parseInliner([]byte{'*'}, Emphasis)
		st.parseImage()
		st.parseHtml()
	}
	st.flushText()
	return st.children
}

func parseSnippet(st *ParserState) Node {
	s1 := st.source
	for !st.eof() {
		line := st.consumeLine()
		if bytes.Index(line, []byte{'<', '/', 's', 'n', 'i', 'p', 'p', 'e', 't', '>'}) >= 0 {
			break
		}
	}
	l := len(s1) - len(st.source)
	// TODO: тут надо извлечь сам текст сниппета и его айдишник
	block := s1[:l]
	children := parseText(block)
	html := children[0]
	node := Node{
		Type:       Snippet,
		Attributes: map[string]string{},
		Literal:    html.Literal,
	}
	for k, v := range html.Attributes {
		if k != "tag" {
			node.Attributes[k] = v
		}
	}
	return node
}

var yamlkvRegexp = regexp.MustCompile(`([a-z]+): (.+)`) //nolint:golint,lll

func parseMeta(doc *Node, st *ParserState) {
	st.consumeLine()
	for !st.eof() && !st.startsWith("---") && !st.startsWith("\n") {
		line := st.consumeLine()
		kv := yamlkvRegexp.FindAllSubmatch(line, -1)
		if len(kv) > 0 {
			doc.Attributes[string(kv[0][1])] = string(kv[0][2])
		}
	}
	if !st.eof() && st.startsWith("---") {
		st.consumeLine()
	}
}

func parseHeading(st *ParserState) Node {
	level := 0
	for ; level < len(st.source) && st.source[level] == '#'; level++ {
	}
	st.consumeN(level + 1)
	node := Node{
		Type:       Heading,
		Attributes: map[string]string{},
	}

	h := st.consumeLine()
	if a1 := bytes.Index(h, []byte{' ', '{'}); a1 > 0 {
		a2 := bytes.Index(h[a1:], []byte{'}'})
		if a2 < 0 {
			a2 = len(h) - a1
		}
		attrsLine := h[a1+2 : a1+a2]
		h = h[:a1]
		if len(attrsLine) > 1 && attrsLine[0] == '#' {
			node.Attributes["id"] = string(attrsLine[1:])
		}
	}

	node.Literal = string(h)
	node.Attributes["level"] = fmt.Sprintf("%d", level)
	return node
}

func parseList(st *ParserState) Node {
	node := Node{
		Type:       List,
		Attributes: map[string]string{},
		Children:   []Node{},
	}
	for !st.eof() && st.startsWith("* ") {
		st.consumeN(2)
		line := st.consumeLine()
		text := parseText(line)
		li := Node{
			Type:     ListItem,
			Children: text,
		}
		node.Children = append(node.Children, li)
	}
	return node
}

func parseAdmonition(st *ParserState) Node {
	st.consumeN(len("!!! "))
	level := string(st.consumeLine())
	child := parseParagraph(st)
	node := Node{
		Type:       Admonition,
		Attributes: map[string]string{"level": level},
		Children:   []Node{child},
	}
	return node
}

type ParserState struct {
	source []byte
}

func (st *ParserState) eof() bool {
	return len(st.source) == 0
}

func (st *ParserState) startsWith(s string) bool {
	return st.head(len(s)) == s
}

func (st *ParserState) head(counts ...int) string {
	count := 1
	if len(counts) > 0 {
		count = counts[0]
	}
	if count > len(st.source) {
		count = len(st.source)
	}
	h := st.source[:count]
	if h[len(h)-1] == '\r' {
		h = append(h[0:len(h)-1], '\n')
	}
	return string(h)
}

func (st *ParserState) consumeLine() []byte {
	pos := 0
	for ; pos < len(st.source) && st.source[pos] != '\n'; pos++ {
	}
	end := pos
	if pos < len(st.source) {
		if pos > 0 && st.source[end-1] == '\r' {
			end -= 1
		}
		if st.source[pos] == '\n' {
			pos += 1
		}
	}
	line := st.source[0:end]
	st.source = st.source[pos:]
	return line
}

func (st *ParserState) consumeN(n int) []byte {
	if n > len(st.source) {
		n = len(st.source)
	}
	line := st.source[:n]
	st.source = st.source[n:]
	return line
}
