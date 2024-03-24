package md2json

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

func Latex(n *Node) ([]byte, error) {
	var text bytes.Buffer
	if n.Children != nil {
		for _, ch := range n.Children {
			text.Write(writeTexNode(&ch))
			text.WriteString("\n")
		}
	}
	return text.Bytes(), nil
}

func writeTexNode(n *Node) []byte {
	switch n.Type {
	case Paragraph:
		return writeTexParagraph(n)
	case Text:
		return writeTexText(n)
	case Image:
		return writeTexImage(n)
	case Link:
		return writeTexLink(n)
	case Emphasis:
		return writeTexEmphasis(n)
	case Bold:
		return writeTexBold(n)
	case Code:
		return writeTexCode(n)
	case Heading:
		return writeTexHeading(n)
	case List:
		return writeTexList(n)
	case ListItem:
		return writeTexListItem(n)
	case Admonition:
		return writeTexAdmonition(n)
	case CodeFence:
		return writeTexCodeFence(n)
	case Table:
		return writeTexTable(n)
	case "NewPage":
		return writeTexNewpage(n)
	// case HTML:
	// 	return writeHTML(n)
	default:
		if n.Attributes != nil {
			tag, _ := n.Attributes["tag"]
			if tag == "if" || tag == "br" {
				return []byte{}
			}
		}
		fmt.Println("Type", n)
	}
	return []byte{}

}

func writeTexNewpage(n *Node) []byte {
	return []byte("\\newpage\n")
}

func writeTexChildren(n *Node) []byte {
	var text bytes.Buffer
	if n.Children != nil {
		for _, ch := range n.Children {
			text.Write(writeTexNode(&ch))
		}
	}
	return text.Bytes()
}

func writeTexParagraph(n *Node) []byte {
	var text bytes.Buffer
	text.Write(writeTexChildren(n))
	text.WriteString("\n")
	return text.Bytes()
}

func writeTexText(n *Node) []byte {
	return []byte(escapeTexText(n.Literal))
}

func writeTexImage(n *Node) []byte {
	src, _ := n.Attributes["src"]
	// TODO: add svg support
	if strings.HasSuffix(src, ".svg") {
		return []byte{}
	}
	return []byte(fmt.Sprintf(
		"\\documentImage{%s}{%s}\n",
		escapeTexText(n.Literal),
		strings.ReplaceAll(src, "_", "\\_")))
}

func writeTexLink(n *Node) []byte {
	src, _ := n.Attributes["href"]
	anchor, hasAnchor := n.Attributes["anchor"]
	if strings.HasPrefix(src, "http") || !hasAnchor || len(anchor) == 0 {
		return []byte(fmt.Sprintf(`\href{%s}{%s}`, src, escapeTexText(n.Literal)))
	} else {
		return []byte(fmt.Sprintf(`\hyperref[%s]{%s}`, escapeTexText(anchor), escapeTexText(n.Literal)))
	}
}

func writeTexEmphasis(n *Node) []byte {
	return []byte(fmt.Sprintf(`\emph{%s}`, escapeTexText(n.Literal)))
}

func writeTexBold(n *Node) []byte {
	return []byte(fmt.Sprintf(`\textbf{%s}`, escapeTexText(n.Literal)))
}

func labelTex(t string) string {
	return strings.ReplaceAll(t, "_", "-")
}

func writeTexHeading(n *Node) []byte {
	var text bytes.Buffer
	text.WriteByte('\\')
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
	switch level {
	case 0:
		text.WriteString("chapter")
	case 1:
		text.WriteString("section")
	case 2:
		text.WriteString("subsection")
	case 3:
		text.WriteString("subsubsection")
	case 4:
		text.WriteString("paragraph")
	default:
		text.WriteString("subparagraph")
	}
	text.WriteString("{")
	text.WriteString(escapeTexText(n.Literal))
	text.WriteString("}")
	id, ok := n.Attributes["id"]
	if ok {
		text.WriteString("\\label{")
		text.WriteString(labelTex(id))
		text.WriteString("}")
	}
	text.WriteString("\n\n")
	return text.Bytes()
}

func writeTexList(n *Node) []byte {
	var text bytes.Buffer
	text.WriteString("\\begin{itemize}\n")
	text.Write(writeTexChildren(n))
	text.WriteString("\\end{itemize}\n\n")
	return text.Bytes()
}

func writeTexListItem(n *Node) []byte {
	var text bytes.Buffer
	text.WriteString("\\item\n")
	// inner := writeTexChildren(n)
	inner := writeTexNode(&n.Children[0])
	rows := bytes.Split(inner, []byte{'\n'})
	if len(rows[len(rows)-1]) == 0 {
		rows = rows[:len(rows)-1]
	}
	for _, r := range rows {
		text.WriteString("  ")
		text.Write(r)
		text.WriteByte('\n')
	}
	return text.Bytes()
}

func escapeTexText(t string) string {
	t = strings.ReplaceAll(t, "#", "\\#")
	t = strings.ReplaceAll(t, "%", "\\%")
	t = strings.ReplaceAll(t, "$", "\\$")
	t = strings.ReplaceAll(t, "&", "\\&")
	t = strings.ReplaceAll(t, "_", "\\_")
	return t
}

func escapeTex(in string) string {
	t := strings.ReplaceAll(in, "|", "\\|")
	t = strings.ReplaceAll(t, "_", "\\_")
	t = strings.ReplaceAll(t, "#", "\\#")
	return t
}

func writeTexCode(n *Node) []byte {
	var text bytes.Buffer
	lang := "c"
	if n.Attributes != nil {
		lang1, ok := n.Attributes["lang"]
		if !ok {
			lang = lang1
		}
	}
	text.WriteString("\\inlineCode{" + lang + "}")
	bracket := "|"
	if strings.Index(n.Literal, "|") >= 0 && strings.Index(n.Literal, "$") < 0 {
		bracket = "$"
	}
	text.WriteString(bracket)
	text.WriteString(n.Literal)
	text.WriteString(bracket)
	return text.Bytes()
}

func writeTexCodeFence(n *Node) []byte {
	var text bytes.Buffer
	text.WriteString("\\begin{multilineCode}\n")
	text.WriteString(n.Literal)
	text.WriteString("\\end{multilineCode}\n")
	return text.Bytes()
}

func writeTexAdmonition(n *Node) []byte {
	var text bytes.Buffer
	level, _ := n.Attributes["level"]
	text.WriteString("\\begin{")
	text.WriteString(level)
	text.WriteString("}\n")
	text.Write(writeTexChildren(n))
	text.WriteString("\n\\end{")
	text.WriteString(level)
	text.WriteString("}\n\n")
	return text.Bytes()
}

func writeTexTable(n *Node) []byte {
	var text bytes.Buffer
	header := n.Children[0]
	body := n.Children[1]
	text.WriteString("\\begin{tabular}{")
	var h bytes.Buffer
	for i, th := range header.Children {
		if i > 0 {
			text.WriteString("|")
			h.WriteString(" & ")
		}
		text.WriteString("c")
		h.WriteString("\\textbf{")
		h.WriteString(escapeTexText(th.Literal))
		h.WriteString("}")
	}
	text.WriteString("}\n")
	h.WriteString("\\\\\n")
	text.Write(h.Bytes())
	text.WriteString("\\hline\n")
	for _, row := range body.Children {
		for i, cell := range row.Children {
			if i > 0 {
				text.WriteString(" & ")
			}
			text.Write(writeTexChildren(&cell))
		}
		text.WriteString(" \\\\\n")
	}
	text.WriteString("\\end{tabular}\n\n")
	return text.Bytes()
}
