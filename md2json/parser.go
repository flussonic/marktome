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
	Link       Kind = "Link"
	List       Kind = "List"
	ListItem   Kind = "ListItem"
	Admonition Kind = "Admonition"
	Code       Kind = "Code"
	CodeFence  Kind = "CodeFence"
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

		// skip comment
		if st.startsWith("<!--") {
			closing := "-->"
			i := bytes.Index(st.source, []byte(closing))
			st.consumeN(i + len(closing))
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
			node.Children = append(node.Children, parseList(st, []byte{'*', ' '}))
			continue
		}

		if st.startsWith("- ") {
			node.Children = append(node.Children, parseList(st, []byte{'-', ' '}))
			continue
		}

		if st.startsWith("!!! ") {
			node.Children = append(node.Children, parseAdmonition(st))
			continue
		}

		if st.startsWith("```") {
			node.Children = append(node.Children, parseCodeFence(st))
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
	// fmt.Printf("Reading html from %d '%s'\n", st.pos, string(st.source))
	st.flushText()
	st.consumeN(1)
	tagEnd := bytes.Index(st.source, []byte{'>'})
	if tagEnd < 0 {
		panic(fmt.Sprintf("No closing tag for '%s'", string(st.source)))
	}
	nameEnd := bytes.Index(st.source[:tagEnd], []byte{' '})
	if nameEnd < 0 {
		nameEnd = tagEnd
	}
	tagName := st.consumeN(nameEnd)
	attrs := map[string]string{}
	attrs["tag"] = string(tagName)

	tagEnd = bytes.Index(st.source, []byte{'>'})
	if tagEnd < 0 {
		return
	}
	selfClosing := false
	if tagEnd > 0 && st.source[tagEnd-1] == '/' {
		selfClosing = true
		tagEnd -= 1
	}
	tagAttrs := st.consumeN(tagEnd)
	st.consumeN(1) // wipe > symbol
	if selfClosing {
		st.consumeN(1)
	}
	if st.startsWith([]byte{'\n'}) {
		st.consumeN(1)
	}

	attrMatches := tagAttrsRegexp.FindAllSubmatch(tagAttrs, -1)
	for i := range attrMatches {
		attrs[string(attrMatches[i][1])] = string(attrMatches[i][2])
	}
	var text string
	if !selfClosing {
		closure := append([]byte{'<', '/'}, tagName...)
		closure = append(closure, '>')
		finish := bytes.Index(st.source, closure)
		if finish < 0 {
			return
		}
		text = string(st.consumeN(finish))
		st.consumeN(len(closure))
	} else {
		text = ""
	}
	node := Node{
		Type:       HTML,
		Attributes: attrs,
		Literal:    text,
	}
	st.children = append(st.children, node)
}

func (st *InlineParserState) readLink() (Node, bool) {
	if !st.startsWith([]byte{'['}) {
		return Node{}, false
	}
	st.consumeN(1)
	titleEnd := bytes.Index(st.source, []byte{']', '('})
	if titleEnd < 0 {
		return Node{}, false
	}
	title := string(st.consumeN(titleEnd))
	st.consumeN(2)
	urlEnd := bytes.Index(st.source, []byte{')'})
	if urlEnd < 0 {
		return Node{}, false
	}
	url := string(st.consumeN(urlEnd))
	st.consumeN(1)
	node := Node{
		Type:       Link,
		Attributes: map[string]string{"title": title, "src": url},
	}
	return node, true
}

func (st *InlineParserState) parseLink() {
	node, ok := st.readLink()
	if !ok {
		return
	}
	node.Type = Link
	st.children = append(st.children, node)
}

func (st *InlineParserState) parseImage() {
	if !st.startsWith([]byte{'!'}) {
		return
	}
	st.consumeN(1)
	node, ok := st.readLink()
	if !ok {
		return
	}
	node.Type = Image
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
		st.parseLink()
		st.parseImage()
		st.parseHtml()
	}
	st.flushText()
	return st.children
}

func parseSnippet(st *ParserState) Node {
	st.consumeN(len("<snippet "))
	tagEnd := bytes.Index(st.source, []byte{'>'})
	tagAttrs := st.consumeN(tagEnd)
	st.consumeN(1)
	if st.startsWith("\n") {
		st.consumeN(1)
	}
	s1 := st.source
	for !st.eof() {
		line := st.consumeLine()
		if bytes.Index(line, []byte("</snippet>")) >= 0 {
			break
		}
	}
	l := len(s1) - len(st.source)
	node := Node{
		Type:       Snippet,
		Attributes: map[string]string{},
		Literal:    string(s1[:l]),
	}

	attrMatches := tagAttrsRegexp.FindAllSubmatch(tagAttrs, -1)
	for i := range attrMatches {
		node.Attributes[string(attrMatches[i][1])] = string(attrMatches[i][2])
	}
	return node
}

func parseCodeFence(st *ParserState) Node {
	st.consumeN(len("```"))
	language := st.consumeLine()
	s1 := st.source
	for !st.eof() {
		line := st.consumeLine()
		if bytes.Index(line, []byte{'`', '`', '`'}) >= 0 {
			break
		}
	}
	l := len(s1) - len(st.source)
	block := s1[:l]
	node := Node{
		Type:       CodeFence,
		Attributes: map[string]string{},
		Literal:    string(block),
	}
	if len(language) > 0 {
		node.Attributes["lang"] = string(language)
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

func parseList(st *ParserState, symbol []byte) Node {
	node := Node{
		Type:       List,
		Attributes: map[string]string{},
		Children:   []Node{},
	}
	for !st.eof() && st.startsWith(string(symbol)) {
		st.consumeN(2)
		line := st.consumeLine()
		text := parseText(line)
		li := Node{
			Type:     ListItem,
			Children: text,
		}
		node.Children = append(node.Children, li)
		// Обрабатываем лишний перенос строк между элементами списка
		if st.startsWith("\n" + string(symbol)) {
			st.consumeN(1)
		}
	}
	return node
}

func parseAdmonition(st *ParserState) Node {
	st.consumeN(len("!!! "))
	level := string(st.consumeLine())
	for ; st.startsWith(" "); st.consumeN(1) {
	}
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
	if len(s) >= len(st.source) {
		return false
	}
	return bytes.Equal([]byte(s), st.source[:len(s)])
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
