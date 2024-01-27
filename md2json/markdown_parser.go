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

	if st.startsWith("---\n") {
		parseMeta(st, &node)
	}

	for !st.eof() {
		for !st.eof() && st.startsWith("\n") {
			st.consumeLine()
		}
		if st.eof() {
			break
		}

		if parseComment(st, &node) {
			continue
		}
		if parseBlockHTML(st, &node) {
			continue
		}
		if parseHeading(st, &node) {
			continue
		}
		if parseList(st, &node) {
			continue
		}
		if parseAdmonition(st, &node) {
			continue
		}
		if parseCodeFence(st, &node) {
			continue
		}
		if parseTable(st, &node) {
			continue
		}
		// goes last
		parseParagraph(st, &node)
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
	j := bytes.Index(st.source[len(symbol):], []byte{'\n'})
	if j > 0 && j < i {
		return
	}
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

func parseParagraph(st *ParserState, node *Node) bool {
	s1 := st.source
	for !st.eof() && !st.startsWith("\n") {
		st.consumeLine()
	}
	l := len(s1) - len(st.source)
	block := bytes.TrimSuffix(s1[:l], []byte{'\n'})
	children := parseText(block)
	if len(children) > 0 {
		n := Node{
			Type:       Paragraph,
			Attributes: map[string]string{},
			Children:   children,
		}
		node.Children = append(node.Children, n)
		return true
	}
	return false
}

func parseTable(st *ParserState, node *Node) bool {
	if !st.startsWith("|") {
		return false
	}
	source := st.source
	headerText := st.consumeLine()
	if !st.startsWith("|-") {
		st.source = source
		return false
	}
	st.consumeLine() // Line with |----|---|
	if !st.startsWith("|") {
		st.source = source
		return false
	}

	header := Node{Type: TableHead, Children: make([]Node, 0)}
	headerParts := bytes.Split(headerText, []byte{'|'})
	for i, p := range headerParts {
		p_ := strings.Trim(string(p), " \t")
		if i > 0 && i < len(headerParts)-1 {
			header.Children = append(header.Children, Node{Type: Text, Literal: p_})
		}
	}

	body := Node{Type: TableBody, Children: make([]Node, 0)}
	for st.startsWith("|") {
		row := Node{Type: TableRow, Children: make([]Node, 0)}
		line := st.consumeLine()
		rowParts := bytes.Split(line, []byte{'|'})
		for i, p := range rowParts {
			p_ := strings.Trim(string(p), " \t")
			cell := parseText([]byte(p_))
			if i > 0 && i < len(rowParts)-1 {
				row.Children = append(row.Children, Node{Type: TableCell, Children: cell})
			}
		}
		body.Children = append(body.Children, row)
	}

	table := Node{Type: Table, Children: []Node{header, body}}
	node.Children = append(node.Children, table)
	return true
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

func parseComment(st *ParserState, node *Node) bool {
	opening := "<!--"
	closing := "-->"
	if st.startsWith(opening) {
		i := bytes.Index(st.source[len(opening):], []byte(closing))
		if i >= 0 {
			st.consumeN(len(opening))
			text := st.consumeN(i)
			st.consumeN(len(closing))
			node.Children = append(node.Children, Node{Type: Comment, Literal: string(text)})
			return true
		}
		return false
	}
	return false
}

func parseBlockHTML(st *ParserState, node *Node) bool {
	if !st.startsWith("<") {
		return false
	}
	// FIXME: Dirty hack with knowledge about inline/block html elements
	if st.startsWith("<link ") {
		return false
	}
	st1 := InlineParserState{
		source:   st.source,
		children: []Node{},
		text:     []byte{},
	}
	st1.parseHtml()
	n1 := Node{}
	if len(st1.children) > 0 {
		st.consumeN(len(st.source) - len(st1.source))
		n1 = st1.children[0]
	} else {
		line := st.consumeLine()
		n1 = Node{Type: Text, Literal: string(line)}
	}
	node.Children = append(node.Children, n1)
	return true
}

func parseCodeFence(st *ParserState, node *Node) bool {
	if !st.startsWith("```") {
		return false
	}
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
	n1 := Node{
		Type:       CodeFence,
		Attributes: map[string]string{},
		Literal:    string(block),
	}
	if len(language) > 0 {
		n1.Attributes["lang"] = string(language)
	}
	node.Children = append(node.Children, n1)
	return true
}

var yamlkvRegexp = regexp.MustCompile(`(\w+): (.+)`) //nolint:golint,lll

func parseMeta(st *ParserState, doc *Node) {
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

func parseHeading(st *ParserState, node *Node) bool {
	if !st.startsWith("#") {
		return false
	}
	level := 0
	for ; level < len(st.source) && st.source[level] == '#'; level++ {
	}
	st.consumeN(level + 1)
	n1 := Node{
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
			n1.Attributes["id"] = string(attrsLine[1:])
		}
	}

	n1.Literal = string(h)
	n1.Attributes["level"] = fmt.Sprintf("%d", level)
	node.Children = append(node.Children, n1)
	return true
}

func parseList(st *ParserState, node *Node) bool {
	symbol := []byte{}
	if st.startsWith("* ") {
		symbol = []byte("* ")
	} else if st.startsWith("- ") {
		symbol = []byte("- ")
	} else {
		return false
	}

	n1 := Node{
		Type:       List,
		Attributes: map[string]string{},
		Children:   []Node{},
	}
	for !st.eof() && st.startsWith(string(symbol)) {
		st.consumeN(len(symbol))
		line := st.consumeLine()
		if st.startsWith("\n") {
			st.consumeLine()
		}
		text := parseText(line)
		liContent := Node{
			Type:     Paragraph,
			Children: text,
		}
		li := Node{
			Type:     ListItem,
			Children: []Node{liContent},
		}
		nestedPrefix := "    "
		if st.startsWith(nestedPrefix) {
			var nested bytes.Buffer
			for st.startsWith(nestedPrefix) {
				l := bytes.TrimPrefix(st.consumeLine(), []byte(nestedPrefix))
				nested.Write(l)
				nested.WriteString("\n")
			}
			nestedDoc := MarkdownParse(nested.Bytes())
			li.Children = append(li.Children, nestedDoc.Children...)
		}
		n1.Children = append(n1.Children, li)
	}
	node.Children = append(node.Children, n1)
	return true
}

func parseAdmonition(st *ParserState, node *Node) bool {
	if !st.startsWith("!!! ") {
		return false
	}
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
	n1 := Node{
		Type:       Admonition,
		Attributes: map[string]string{"level": level},
		Children:   children,
	}
	node.Children = append(node.Children, n1)
	return true
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
