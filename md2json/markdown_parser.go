package md2json

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
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

		if st.startsWith("<") && !st.startsWith("<link ") {
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

		n := parseParagraph(st)
		if len(n.Literal) > 0 || (n.Children != nil && len(n.Children) > 0) {
			node.Children = append(node.Children, n)
		}
	}
	return node
}

func parseParagraph(st *ParserState) Node {
	s1 := st.source
	for !st.eof() && !st.startsWith("\n") {
		st.consumeLine()
	}
	l := len(s1) - len(st.source)
	block := bytes.TrimSuffix(s1[:l], []byte{'\n'})
	node := Node{
		Type:       Paragraph,
		Attributes: map[string]string{},
		Children:   parseText(block),
	}
	return node
}

type InlineParserState struct {
	source   []byte
	children []Node
	text     []byte
}

func (st *InlineParserState) flushText() {
	if len(st.text) > 0 {
		node := Node{Type: Text, Literal: string(st.text)}
		st.children = append(st.children, node)
		st.text = []byte{}
	}
}

func (st *InlineParserState) consumeN(n int) []byte {
	if n > len(st.source) {
		n = len(st.source)
	}
	s := st.source[:n]
	st.source = st.source[n:]
	return s
}

func (st *InlineParserState) startsWith(s []byte) bool {
	if len(s) >= len(st.source) {
		return false
	}
	return bytes.Equal(s, st.source[:len(s)])
}

func (st *InlineParserState) parseCode() {
	if !st.startsWith([]byte{'`'}) {
		return
	}
	i := bytes.Index(st.source[1:], []byte{'`'})
	if i < 0 {
		return
	}
	st.flushText()
	st.consumeN(1)
	text := st.consumeN(i)
	st.consumeN(1)
	node := Node{
		Type:    Code,
		Literal: string(text),
	}
	st.children = append(st.children, node)
}

func (st *InlineParserState) parseInliner(symbol []byte, typ Kind) {
	if !st.startsWith(symbol) {
		return
	}
	i := bytes.Index(st.source[len(symbol):], symbol)
	if i > 0 {
		st.flushText()
		st.consumeN(len(symbol))
		node := Node{Type: typ}
		children := parseText(st.consumeN(i))
		if len(children) == 1 && children[0].Type == Text {
			node.Literal = children[0].Literal
		} else {
			node.Children = children
		}
		st.children = append(st.children, node)
		st.consumeN(len(symbol))
	}
}

var tagAttrsRegexp = regexp.MustCompile(`([a-z]+)="([^"]+)"`) //nolint:golint,lll

func (st *InlineParserState) parseHtml() {
	br := "<br>"
	if st.startsWith([]byte(br)) {
		st.flushText()
		st.consumeN(len(br))
		node := Node{
			Type:       HTML,
			Attributes: map[string]string{"tag": "br"},
		}
		st.children = append(st.children, node)
		return
	}
	if st.startsWith([]byte{'<', '/'}) {
		return
	}
	if !st.startsWith([]byte{'<'}) {
		return
	}
	st.flushText()
	st.consumeN(1)
	tagEnd := bytes.Index(st.source, []byte{'>'})
	if tagEnd < 0 {
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
	if !st.startsWith([]byte{'['}) || len(st.source) < 4 {
		return
	}
	titleEnd := bytes.Index(st.source, []byte{']', '('})
	if titleEnd < 0 {
		return
	}
	urlEnd := bytes.Index(st.source[titleEnd:], []byte{')'})
	if urlEnd < 0 {
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
	if !st.startsWith([]byte{'!', '['}) || len(st.source) < 5 {
		return
	}
	titleEnd := bytes.Index(st.source, []byte{']', '('})
	urlEnd := bytes.Index(st.source[titleEnd:], []byte{')'})
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
		text:     []byte{},
	}
	for len(st.source) > 0 {
		l1 := len(st.source)
		st.parseCode()
		st.parseInliner([]byte{'*', '*'}, Bold)
		st.parseInliner([]byte{'*'}, Emphasis)
		st.parseLink()
		st.parseImage()
		st.parseHtml()
		if l1 == len(st.source) {
			st.text = append(st.text, st.source[0])
			st.source = st.source[1:]
		}
	}
	st.flushText()
	return st.children
}

func parseBlockHTML(st *ParserState) Node {
	st1 := InlineParserState{
		source:   st.source,
		children: []Node{},
		text:     []byte{},
	}
	st1.parseHtml()
	if len(st1.children) > 0 {
		st.consumeN(len(st.source) - len(st1.source))
		return st1.children[0]
	}
	line := st.consumeLine()
	return Node{Type: Text, Literal: string(line)}
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
	}
	return node
}

func parseAdmonition(st *ParserState) Node {
	st.consumeN(len("!!! "))
	level := strings.TrimSpace(string(st.consumeLine()))
	starter := "    "
	var text bytes.Buffer
	first := true
	for st.startsWith(starter) {
		line := st.consumeLine()[len(starter):]
		if !first {
			text.WriteString("\n")
		}
		text.Write(line)
		first = false
	}
	children := parseText(text.Bytes())
	node := Node{
		Type:       Admonition,
		Attributes: map[string]string{"level": level},
		Children:   children,
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
