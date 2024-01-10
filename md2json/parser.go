package md2json

import (
	"bytes"
	"fmt"
	"regexp"
)

func ParseDocument(source []byte) NodeData {
	st := ParserState{
		source: source,
	}
	node := parseDocument(&st)
	return node
}

func parseDocument(st *ParserState) NodeData {
	node := NodeData{
		Type:       "Document",
		Attributes: map[string]string{},
		Children:   []NodeData{},
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

		child := parseParagraph(st)
		node.Children = append(node.Children, child)
	}
	// switch source[pos] {
	// case '#':

	// }
	return node
}

func parseParagraph(st *ParserState) NodeData {
	s1 := st.source
	for !st.eof() && !st.startsWith("\n") {
		st.consumeLine()
	}
	l := len(s1) - len(st.source)
	block := s1[:l]
	node := NodeData{
		Type:       "Paragraph",
		Attributes: map[string]string{},
		Children:   parseText(block),
	}
	return node
}

func parseText(source []byte) []NodeData {
	node := NodeData{
		Type:       "Text",
		Attributes: map[string]string{},
		Literal:    string(source),
	}
	return []NodeData{node}
}

func parseSnippet(st *ParserState) NodeData {
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
	node := NodeData{
		Type:       "Snippet",
		Attributes: map[string]string{},
		Literal:    string(block),
	}
	return node
}

var yamlkvRegexp = regexp.MustCompile(`([a-z]+): (.+)`) //nolint:golint,lll

func parseMeta(doc *NodeData, st *ParserState) {
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

func parseHeading(st *ParserState) NodeData {
	level := 0
	for ; level < len(st.source) && st.source[level] == '#'; level++ {
	}
	st.consumeN(level + 1)
	node := NodeData{
		Type:       "Heading",
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

func parseList(st *ParserState) NodeData {
	node := NodeData{
		Type:       "List",
		Attributes: map[string]string{},
		Children:   []NodeData{},
	}
	for !st.eof() && st.startsWith("* ") {
		st.consumeN(2)
		line := st.consumeLine()
		text := parseText(line)
		li := NodeData{
			Type:     "ListItem",
			Children: text,
		}
		node.Children = append(node.Children, li)
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
