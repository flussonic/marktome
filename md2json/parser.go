package md2json

import (
	"bytes"
	"fmt"
	"regexp"
)

func ParseDocument(source []byte) NodeData {
	st := ParserState{
		pos:    0,
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
		for !st.eof() && st.head() == "\n" {
			st.consumeLine()
		}
		if st.eof() {
			break
		}

		if st.head(4) == "---\n" {
			parseMeta(&node, st)
			continue
		}

		if st.head() == "#" {
			node.Children = append(node.Children, parseHeading(st))
			continue
		}

		child := parseBlock(st)
		node.Children = append(node.Children, child)
	}
	// switch source[pos] {
	// case '#':

	// }
	return node
}

func parseBlock(st *ParserState) NodeData {
	begin := st.pos
	for !st.eof() && st.head() != "\n" {
		st.consumeLine()
	}
	end := st.pos
	block := st.source[begin:end]
	node := NodeData{
		Type:       "Paragraph",
		Attributes: map[string]string{},
		Literal:    string(block),
	}
	return node
}

var yamlkvRegexp = regexp.MustCompile(`([a-z]+): (.+)`) //nolint:golint,lll

func parseMeta(doc *NodeData, st *ParserState) {
	st.consumeLine()
	for !st.eof() && st.head(3) != "---" && st.head() != "\n" {
		line := st.consumeLine()
		kv := yamlkvRegexp.FindAllSubmatch(line, -1)
		if len(kv) > 0 {
			doc.Attributes[string(kv[0][1])] = string(kv[0][2])
		}
	}
	if !st.eof() && st.head(3) == "---" {
		st.consumeLine()
	}
}

func parseHeading(st *ParserState) NodeData {
	level := 0
	for ; st.pos+level < len(st.source) && st.source[st.pos+level] == '#'; level++ {
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

type ParserState struct {
	pos    int
	source []byte
}

func (st *ParserState) eof() bool {
	return st.pos >= len(st.source)
}

func (st *ParserState) head(counts ...int) string {
	count := 1
	if len(counts) > 0 {
		count = counts[0]
	}
	if st.pos+count > len(st.source) {
		count = len(st.source) - st.pos
	}
	h := st.source[st.pos : st.pos+count]
	if h[len(h)-1] == '\r' {
		h = append(h[0:len(h)-1], '\n')
	}
	return string(h)
}

func (st *ParserState) consumeLine() []byte {
	pos := st.pos
	begin := st.pos
	for ; pos < len(st.source) && st.source[pos] != '\n'; pos++ {
	}
	end := pos
	if pos < len(st.source) {
		if pos > begin && st.source[end-1] == '\r' {
			end -= 1
		}
		if st.source[pos] == '\n' {
			pos += 1
		}
	}
	st.pos = pos
	return st.source[begin:end]
}

func (st *ParserState) consumeN(n int) []byte {
	begin := st.pos
	if begin+n > len(st.source) {
		n = len(st.source) - begin
	}
	st.pos += n
	return st.source[begin : begin+n]
}
