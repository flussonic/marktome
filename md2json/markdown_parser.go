package md2json

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
)

func MarkdownParse(source []byte) Node {
	st := ParserState{
		source: source,
	}
	node := parseDocument(&st)
	return node
}

func Md2Json(input string, output string) error {
	source, err := os.ReadFile(input)
	if err != nil {
		return err
	}
	json := MarkdownParse(source)
	err = WriteJson(&json, output)
	return err
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

		if st.startsWith("<") {
			node.Children = append(node.Children, parseBlockHTML(st))
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
	// fmt.Printf("Go into paragraph: '%s'\n", s1[:20])
	for !st.eof() && !st.startsWith("\n") && !st.startsWith("<") && !st.startsWith("```") {
		st.consumeLine()
		// fmt.Printf("Consume line '%s'\n", s)
	}
	// for !st.eof() && (st.source[0] == ' ' || st.source[0] == '\t') {
	// 	fmt.Printf("Consume char %d\n", st.source[0])
	// 	st.consumeN(1)
	// }
	l := len(s1) - len(st.source)
	// fmt.Printf("Paragraph is %d bytes. Flags: %b/%b/%b/%b\n", l,
	// 	st.eof(), st.startsWith("\n"), st.startsWith("<"), st.startsWith("```"))
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
		if st.pos < len(st.source) {
			st.source = st.source[st.pos:]
		} else {
			st.source = []byte{}
		}
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
	// fmt.Printf("Reading html from %d '%s'\n", st.pos, string(st.source[st.pos:]))
	st.flushText()
	st.consumeN(1)
	tagEnd := bytes.Index(st.source, []byte{'>'})
	if tagEnd < 0 {
		// fmt.Printf("No closing tag for '%s'", string(st.source))
		return
	}
	nameEnd := bytes.Index(st.source[:tagEnd], []byte{' '})
	if nameEnd < 0 {
		nameEnd = tagEnd
	}
	tagName := st.consumeN(nameEnd)
	attrs := map[string]string{}
	tag := string(tagName)
	attrs["tag"] = tag

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
		if finish >= 0 {
			text = string(st.consumeN(finish))
			st.consumeN(len(closure))
		} else {
			text = string(st.consumeN(len(st.source)))
		}
	} else {
		text = ""
	}
	node := Node{
		Type:       HTML,
		Attributes: attrs,
	}
	if tag == "if" {
		node.Children = parseText([]byte(text))
	} else if tag == "details" {
		doc := MarkdownParse([]byte(text))
		node.Children = doc.Children
	} else {
		node.Literal = text
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
		Attributes: map[string]string{"href": url},
		Literal:    title,
	}
	return node, true
}

func (st *InlineParserState) parseLink() {
	start := st.pos
	if !st.startsWith([]byte{'['}) || len(st.source)-start < 4 {
		return
	}
	titleEnd := bytes.Index(st.source[start:], []byte{']', '('})
	urlEnd := bytes.Index(st.source[start+titleEnd:], []byte{')'})
	if titleEnd < 0 || urlEnd < 0 {
		return
	}
	st.flushText()
	title := string(st.source[1:titleEnd])
	url := string(st.source[titleEnd+2 : titleEnd+urlEnd])
	st.consumeN(titleEnd + urlEnd + 1)
	node := Node{
		Type:       Link,
		Attributes: map[string]string{"href": url},
		Literal:    title,
	}
	st.children = append(st.children, node)
}

func (st *InlineParserState) parseImage() {
	start := st.pos
	if !st.startsWith([]byte{'!', '['}) || len(st.source)-start < 5 {
		return
	}
	titleEnd := bytes.Index(st.source[start:], []byte{']', '('})
	urlEnd := bytes.Index(st.source[start+titleEnd:], []byte{')'})
	if titleEnd < 0 || urlEnd < 0 {
		return
	}
	st.flushText()
	title := string(st.source[2:titleEnd])
	url := string(st.source[titleEnd+2 : titleEnd+urlEnd])
	st.consumeN(titleEnd + urlEnd + 1)
	node := Node{
		Type:       Image,
		Attributes: map[string]string{"src": url},
		Literal:    title,
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
		st.parseLink()
		st.parseImage()
		st.parseHtml()
	}
	st.flushText()
	return st.children
}

func parseBlockHTML(st *ParserState) Node {
	st1 := InlineParserState{
		source:   st.source,
		children: []Node{},
		pos:      0,
	}
	st1.parseHtml()
	if len(st1.children) > 0 {
		st.consumeN(len(st.source) - len(st1.source))
		return st1.children[0]
	}
	line := st.consumeLine()
	return Node{Type: Text, Literal: string(line)}
}

// func parseSnippet(st *ParserState) Node {
// 	st.consumeN(len("<snippet "))
// 	tagEnd := bytes.Index(st.source, []byte{'>'})
// 	tagAttrs := st.consumeN(tagEnd)
// 	st.consumeN(1)
// 	if st.startsWith("\n") {
// 		st.consumeN(1)
// 	}
// 	s1 := st.source
// 	closing := "</snippet>"
// 	contentEnd := bytes.Index(st.source, []byte(closing))
// 	if contentEnd < 0 {
// 		contentEnd = len(st.source)
// 	}
// 	st.consumeN(contentEnd + len(closing))
// 	if st.startsWith("\n") {
// 		st.consumeN(1)
// 	}
// 	node := Node{
// 		Type:       Snippet,
// 		Attributes: map[string]string{},
// 		Literal:    string(s1[:contentEnd]),
// 	}

// 	attrMatches := tagAttrsRegexp.FindAllSubmatch(tagAttrs, -1)
// 	for i := range attrMatches {
// 		node.Attributes[string(attrMatches[i][1])] = string(attrMatches[i][2])
// 	}
// 	return node
// }

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
	l := len(s1) - len(st.source) - 4
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
	// for ; st.startsWith(" "); st.consumeN(1) {
	// }
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

func (st *ParserState) isEmptyLine() bool {
	for i := 0; i < len(st.source); i++ {
		if st.source[i] == '\n' {
			return true
		}
		if st.source[i] != ' ' && st.source[i] != '\t' {
			return false
		}
	}
	return true
}
