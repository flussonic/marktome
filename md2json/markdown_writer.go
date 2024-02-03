package md2json

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strconv"
)

func Json2Md(input string, output string) error {
	doc, err := ReadJson(input)
	if err != nil {
		return err
	}
	text := WriteDocument(&doc)
	err = os.WriteFile(output, text, os.ModePerm)
	return err
}

func WriteDocument(n *Node) []byte {
	var text bytes.Buffer
	text.Write(writeDocumentMeta(n))
	if n.Children != nil {
		for _, ch := range n.Children {
			if len(text.Bytes()) > 0 {
				if !bytes.HasSuffix(text.Bytes(), []byte{'\n'}) {
					text.WriteByte('\n')
				}
				text.WriteByte('\n')
			}
			text.Write(writeNode(&ch))
		}
	}
	return text.Bytes()
}

func writeDocumentMeta(n *Node) []byte {
	if n.Attributes == nil {
		return []byte{}
	}
	var header bytes.Buffer
	header.WriteString("---\n")
	keys := make([]string, 0, len(n.Attributes))
	for k := range n.Attributes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v, _ := n.Attributes[k]
		header.WriteString(k)
		header.WriteString(": ")
		header.WriteString(v)
		header.WriteString("\n")
	}
	header.WriteString("---\n")
	return header.Bytes()
}

func writeNode(n *Node) []byte {
	switch n.Type {
	case Paragraph:
		return writeParagraph(n)
	case Text:
		return writeText(n)
	case Comment:
		return writeComment(n)
	case Image:
		return writeImage(n)
	case Link:
		return writeLink(n)
	case Emphasis:
		return writeEmphasis(n)
	case Bold:
		return writeBold(n)
	case Code:
		return writeCode(n)
	case Heading:
		return writeHeading(n)
	case List:
		return writeList(n)
	// case ListItem:
	// 	return writeListItem(n)
	case Admonition:
		return writeAdmonition(n)
	case CodeFence:
		return writeCodeFence(n)
	case HTML:
		return writeHTML(n)
	case Table:
		return writeTable(n)
	default:
		fmt.Println("Type", n.Type)
	}
	return []byte{}
}

func writeHeading(n *Node) []byte {
	var text bytes.Buffer
	level_, ok := n.Attributes["level"]
	var level int
	var err error
	if ok {
		level, err = strconv.Atoi(level_)
		if err != nil {
			level = 3
		}
	} else {
		level = 3
	}
	for i := 0; i < level; i++ {
		text.WriteString("#")
	}
	text.WriteString(" ")
	text.WriteString(n.Literal)
	id, ok := n.Attributes["id"]
	if ok {
		text.WriteString(" {#")
		text.WriteString(id)
		text.WriteString("}")
	}
	text.WriteString("\n")
	return text.Bytes()
}

func writeChildren(n *Node) []byte {
	var text bytes.Buffer
	if n.Children != nil {
		for _, ch := range n.Children {
			text.Write(writeNode(&ch))
		}
	}
	return text.Bytes()
}

func writeParagraph(n *Node) []byte {
	var text bytes.Buffer
	text.Write(writeChildren(n))
	text.WriteString("\n")
	return text.Bytes()
}

func writeText(n *Node) []byte {
	return []byte(n.Literal)
}

func writeComment(n *Node) []byte {
	var text bytes.Buffer
	text.WriteString("<!--")
	text.WriteString(n.Literal)
	text.WriteString("-->")
	return text.Bytes()
}

func writeInliner(n *Node) []byte {
	if n.Children == nil {
		return []byte(n.Literal)
	} else {
		return writeChildren(n)
	}
}

func writeEmphasis(n *Node) []byte {
	var text bytes.Buffer
	text.WriteString("*")
	text.Write(writeInliner(n))
	text.WriteString("*")
	return text.Bytes()
}
func writeBold(n *Node) []byte {
	var text bytes.Buffer
	text.WriteString("**")
	text.Write(writeInliner(n))
	text.WriteString("**")
	return text.Bytes()
}

func writeCode(n *Node) []byte {
	var text bytes.Buffer
	text.WriteString("`")
	text.WriteString(n.Literal)
	text.WriteString("`")
	return text.Bytes()
}

func writeImage(n *Node) []byte {
	var text bytes.Buffer
	src, _ := n.Attributes["src"]
	text.WriteString("![")
	text.WriteString(n.Literal)
	text.WriteString("](")
	text.WriteString(src)
	text.WriteString(")")
	return text.Bytes()
}

func writeLink(n *Node) []byte {
	var text bytes.Buffer
	src, _ := n.Attributes["href"]
	text.WriteString("[")
	text.WriteString(n.Literal)
	text.WriteString("](")
	text.WriteString(src)
	anchor, ok2 := n.Attributes["anchor"]
	if ok2 {
		text.WriteString("#")
		text.WriteString(anchor)
	}
	text.WriteString(")")
	return text.Bytes()
}

func writeCodeFence(n *Node) []byte {
	var text bytes.Buffer
	text.WriteString("```")
	if n.Attributes != nil {
		lang, ok := n.Attributes["lang"]
		if ok {
			text.WriteString(lang)
		}
	}
	text.WriteString("\n")
	text.WriteString(n.Literal)
	text.WriteString("```\n")
	return text.Bytes()
}

func writeList(n *Node) []byte {
	var text bytes.Buffer
	ordered := false
	if n.Children != nil {
		_, ok := n.Attributes["ordered"]
		if ok {
			ordered = true
		}
	}
	for i, ch := range n.Children {
		j := i + 1
		if !ordered {
			j = -1
		}
		if i > 0 && n.Children[i-1].Children != nil && len(n.Children[i-1].Children) > 1 {
			text.WriteString("\n")
		}
		text.Write(writeListItem(j, &ch))
	}
	return text.Bytes()
}

func writeListItem(i int, n *Node) []byte {
	var text bytes.Buffer
	if i < 0 {
		text.WriteString("* ")
	} else {
		text.WriteString(fmt.Sprintf("%d. ", i))
	}
	if len(n.Children) == 0 {
		text.WriteString("\n")
		return text.Bytes()
	}
	text.Write(writeNode(&n.Children[0]))
	if len(n.Children) == 1 {
		return text.Bytes()
	}
	text.WriteString("\n")
	var nested bytes.Buffer
	for _, ch := range n.Children[1:] {
		nested.Write(writeNode(&ch))
	}
	nestedBytes := nested.Bytes()
	if nestedBytes[len(nestedBytes)-1] == '\n' {
		nestedBytes = nestedBytes[:len(nestedBytes)-1]
	}
	rows := bytes.Split(nestedBytes, []byte("\n"))
	for _, r := range rows {
		text.WriteString("    ")
		text.Write(r)
		text.WriteString("\n")
	}
	return text.Bytes()
}

func writeAdmonition(n *Node) []byte {
	var text bytes.Buffer
	level, _ := n.Attributes["level"]
	text.WriteString("!!! ")
	text.WriteString(level)
	text.WriteString("\n")
	inner := writeChildren(n)
	rows := bytes.Split(inner, []byte{'\n'})
	for _, r := range rows {
		text.WriteString("    ")
		text.Write(r)
		text.WriteByte('\n')
	}
	return text.Bytes()
}

func writeHTML(n *Node) []byte {
	var text bytes.Buffer
	block := false
	tag, _ := n.Attributes["tag"]
	if tag == "if" {
		text.Write(writeChildren(n))
		text.WriteString("\n\n")
		return text.Bytes()
	}
	switch tag {
	case
		"graphviz",
		"snippet",
		"include-snippet":
		block = true
	}
	text.WriteString("<")
	text.WriteString(tag)
	if tag == "br" {
		text.WriteString(">")
		return text.Bytes()
	}
	for k, v := range n.Attributes {
		if k != "tag" {
			text.WriteString(" ")
			text.WriteString(k)
			text.WriteString("=\"")
			text.WriteString(v)
			text.WriteString("\"")
		}
	}
	if len(n.Literal) > 0 || (n.Children != nil && len(n.Children) > 0) {
		text.WriteString(">")
		if block {
			text.WriteString("\n")
		}
		if n.Children == nil {
			text.WriteString(n.Literal)
		} else {
			text.Write(writeChildren(n))
		}
		text.WriteString("</")
		text.WriteString(tag)
		text.WriteString(">")
	} else {
		text.WriteString("/>")
	}
	if block {
		text.WriteString("\n")
	}
	return text.Bytes()
}

func writeTable(node *Node) []byte {
	var text bytes.Buffer
	header := node.Children[0]
	body := node.Children[1]
	text.WriteString("|")

	var second bytes.Buffer
	second.WriteString("|")
	for _, h := range header.Children {
		text.WriteString(" ")
		text.Write(writeNode(&h))
		text.WriteString(" |")
		second.WriteString("---|")
	}
	text.WriteString("\n")
	text.Write(second.Bytes())
	text.WriteString("\n")
	for _, row := range body.Children {
		text.WriteString("|")
		for _, cell := range row.Children {
			text.WriteString(" ")
			text.Write(writeChildren(&cell))
			text.WriteString(" |")
		}
		text.WriteString("\n")
	}
	return text.Bytes()
}
