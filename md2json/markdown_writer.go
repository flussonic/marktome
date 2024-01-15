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
	case CodeFence:
	case Snippet:
		return writeCodeFence(n)
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

func writeParagraph(n *Node) []byte {
	var text bytes.Buffer
	if n.Children != nil {
		for _, ch := range n.Children {
			text.Write(writeNode(&ch))
		}
	}
	text.WriteString("\n\n")
	return text.Bytes()
}

func writeText(n *Node) []byte {
	return []byte(n.Literal)
}

func writeEmphasis(n *Node) []byte {
	var text bytes.Buffer
	text.WriteString("*")
	text.WriteString(n.Literal)
	text.WriteString("*")
	return text.Bytes()
}
func writeBold(n *Node) []byte {
	var text bytes.Buffer
	text.WriteString("**")
	text.WriteString(n.Literal)
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
	title, _ := n.Attributes["title"]
	text.WriteString("![")
	text.WriteString(title)
	text.WriteString("](")
	text.WriteString(src)
	text.WriteString(")")
	return text.Bytes()
}

func writeLink(n *Node) []byte {
	var text bytes.Buffer
	src, _ := n.Attributes["src"]
	title, _ := n.Attributes["title"]
	text.WriteString("[")
	text.WriteString(title)
	text.WriteString("](")
	text.WriteString(src)
	text.WriteString(")")
	return text.Bytes()
}

func writeCodeFence(n *Node) []byte {
	var text bytes.Buffer
	text.WriteString("```\n")
	text.WriteString(n.Literal)
	text.WriteString("```\n\n")
	return text.Bytes()
}
