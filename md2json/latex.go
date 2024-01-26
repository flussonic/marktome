package md2json

import (
	"bytes"
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

func MergeDocument(input string) (*Node, error) {
	if strings.HasSuffix(input, ".md") {
		doc, err := ReadJson(input)
		return &doc, err
	}

	foliant, err := YamlParse(input)
	if err != nil {
		return nil, err
	}
	nav, _ := foliant["chapters"]
	srcDir0, _ := foliant["src_dir"]
	srcDir := filepath.Join(filepath.Dir(input), srcDir0.(string))

	full := []Node{}

	var mergeChapters func(menu interface{}) error
	mergeChapters = func(menu interface{}) error {
		if reflect.TypeOf(menu).Kind() == reflect.Slice {
			for _, v := range menu.([]interface{}) {
				if reflect.TypeOf(v).Kind() == reflect.String {
					if !strings.HasSuffix(v.(string), ".md") {
						continue
					}
					fp := filepath.Join(srcDir, v.(string))
					doc, err := ReadJson(fp)
					if err != nil {
						return err
					}
					full = append(full, doc.Children...)
				}
				if reflect.TypeOf(v).Kind() == reflect.Map {
					err := mergeChapters(v)
					if err != nil {
						return err
					}
				}
				if reflect.TypeOf(v).Kind() == reflect.Slice {
					err := mergeChapters(v)
					if err != nil {
						return err
					}
				}
			}
		}
		if reflect.TypeOf(menu).Kind() == reflect.Map {
			for _, v := range menu.(map[string]interface{}) {
				if reflect.TypeOf(v).Kind() == reflect.String {
					if !strings.HasSuffix(v.(string), ".md") {
						continue
					}
					fp := filepath.Join(srcDir, v.(string))
					doc, err := ReadJson(fp)
					if err != nil {
						return err
					}
					full = append(full, doc.Children...)
				}
				if reflect.TypeOf(v).Kind() == reflect.Map {
					// full = append(full, Node{
					// 	Type:       Heading,
					// 	Attributes: map[string]string{"level": "0"},
					// 	Literal:    k,
					// })
					err := mergeChapters(v)
					if err != nil {
						return err
					}
				}
				if reflect.TypeOf(v).Kind() == reflect.Slice {
					// full = append(full, Node{
					// 	Type:       Heading,
					// 	Attributes: map[string]string{"level": "0"},
					// 	Literal:    k,
					// })
					err := mergeChapters(v)
					if err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
	err = mergeChapters(nav)
	if err != nil {
		return nil, err
	}
	document := Node{
		Type:       Document,
		Attributes: make(AttributeMap),
		Children:   full,
	}
	return &document, nil
}

func Latex(n *Node) ([]byte, error) {
	var text bytes.Buffer
	if n.Children != nil {
		for _, ch := range n.Children {
			text.Write(writeTexNode(&ch))
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
	text.WriteString("\n")
	return text.Bytes()
}

func writeTexText(n *Node) []byte {
	return []byte(escapeTexText(n.Literal))
}

func writeTexImage(n *Node) []byte {
	src, _ := n.Attributes["src"]
	t := `\begin{figure}[hbt!]
	\centering
	\includegraphics[width=0.9\textwidth]{%s}
	\caption{%s}
 \end{figure}`
	return []byte(fmt.Sprintf(t, strings.ReplaceAll(src, "_", "\\_"), n.Literal))
}

func writeTexLink(n *Node) []byte {
	src, _ := n.Attributes["href"]
	anchor, _ := n.Attributes["anchor"]
	if strings.HasSuffix(src, ".md") {
		return []byte(fmt.Sprintf(`\href{%s}{%s}`, labelTex(anchor), escapeTex(n.Literal)))
	} else {
		return []byte(fmt.Sprintf(`\href{%s}{%s}`, src, escapeTex(n.Literal)))
	}
}

func writeTexEmphasis(n *Node) []byte {
	return []byte(fmt.Sprintf(`\emph{%s}`, escapeTex(n.Literal)))
}

func writeTexBold(n *Node) []byte {
	return []byte(fmt.Sprintf(`\textbf{%s}`, escapeTex(n.Literal)))
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
	case 5:
		text.WriteString("subparagraph")
	}
	text.WriteString("{")
	text.WriteString(n.Literal)
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
	inner := writeTexChildren(n)
	rows := bytes.Split(inner, []byte{'\n'})
	for _, r := range rows {
		text.WriteString("  ")
		text.Write(r)
		text.WriteByte('\n')
	}
	return text.Bytes()
}

func escapeTexText(t string) string {
	t = strings.ReplaceAll(t, "#", "\\#")
	t = strings.ReplaceAll(t, "$", "\\$")
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
	text.WriteString("\\mintinline{" + lang + "}|")
	text.WriteString(escapeTex(n.Literal))
	text.WriteString("|")
	return text.Bytes()
}

func writeTexCodeFence(n *Node) []byte {
	var text bytes.Buffer

	// \begin{codesnippet}
	text.WriteString(`\begin{minted}[frame=single,breaklines]`)
	lang := "c"
	if n.Attributes != nil {
		lang1, ok := n.Attributes["lang"]
		if ok {
			lang = lang1
		}
	}
	text.WriteString(fmt.Sprintf("{%s}", lang))
	text.WriteString("\n")
	text.WriteString(n.Literal)
	text.WriteString("\\end{minted}\n\n")
	// это можно добавить для подписи
	// \caption{My func}\label{lst:my_func}
	// \end{codesnippet}

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
		h.WriteString(th.Literal)
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
		text.WriteString("\\\\\n")
	}
	text.WriteString("\\end{tabular}\n\n")
	return text.Bytes()
}
