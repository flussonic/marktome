package md2json

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
)

func Json2Md(input string, output string) error {
	doc, err := ReadJson(input)
	if err != nil {
		return err
	}
	text := writeDocument(&doc)
	err = os.WriteFile(output, text, os.ModePerm)
	return err
}

func writeDocument(n *Node) []byte {
	var text bytes.Buffer
	text.Write(writeDocumentMeta(n))
	if n.Children != nil {
		for _, ch := range n.Children {
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
	for k, v := range n.Attributes {
		header.WriteString(k)
		header.WriteString(": ")
		header.WriteString(v)
		header.WriteString("\n")
	}
	header.WriteString("---\n\n")
	return header.Bytes()
}

func writeNode(n *Node) []byte {
	switch n.Type {
	case Paragraph:
		return writeParagraph(n)
	case Text:
		return writeText(n)
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
	case ListItem:
		return writeListItem(n)
	case Admonition:
		return writeAdmonition(n)
	case CodeFence:
		return writeCodeFence(n)
	case HTML:
		return writeHTML(n)
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
	text.WriteString("\n\n")
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
	text.WriteString("```\n\n")
	return text.Bytes()
}

func writeList(n *Node) []byte {
	var text bytes.Buffer
	text.Write(writeChildren(n))
	text.WriteString("\n")
	return text.Bytes()
}

func writeListItem(n *Node) []byte {
	var text bytes.Buffer
	text.WriteString("* ")
	text.Write(writeChildren(n))
	text.WriteString("\n")
	return text.Bytes()
}

func writeAdmonition(n *Node) []byte {
	var text bytes.Buffer
	level, _ := n.Attributes["level"]
	text.WriteString("!!! ")
	text.WriteString(level)
	text.WriteString("\n")
	text.Write(writeParagraph(n))
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
	if tag == "snippet" {
		block = true
	}
	text.WriteString("<")
	text.WriteString(tag)
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
