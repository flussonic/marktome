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

	renames, err := loadRenames(srcDir)
	if err != nil {
		return nil, err
	}

	full := []Node{}

	var mergeChapters func(menu interface{}) error
	mergeChapters = func(menu interface{}) error {
		if reflect.TypeOf(menu).Kind() == reflect.Slice {
			for _, v := range menu.([]interface{}) {
				if reflect.TypeOf(v).Kind() == reflect.String {
					fp1, _ := renames[v.(string)]
					if !strings.HasSuffix(fp1, ".md") {
						continue
					}
					fp := filepath.Join(srcDir, fp1)
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
					fp1, _ := renames[v.(string)]
					if !strings.HasSuffix(fp1, ".md") {
						continue
					}
					fp := filepath.Join(srcDir, fp1)
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
	// case Code:
	// 	return writeCode(n)
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
	// case HTML:
	// 	return writeHTML(n)
	default:
		fmt.Println("Type", n.Type)
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
	return []byte(n.Literal)
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
		return []byte(fmt.Sprintf(`\href{%s}{%s}`, anchor, n.Literal))
	} else {
		return []byte(fmt.Sprintf(`\href{%s}{%s}`, src, n.Literal))
	}
}

func writeTexEmphasis(n *Node) []byte {
	return []byte(fmt.Sprintf(`\emph{%s}`, n.Literal))
}

func writeTexBold(n *Node) []byte {
	return []byte(fmt.Sprintf(`\textbf{%s}`, n.Literal))
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
	for i := 1; i < level; i++ {
		text.WriteString("sub")
	}
	text.WriteString("section{")
	text.WriteString(n.Literal)
	text.WriteString("}")
	id, ok := n.Attributes["id"]
	if ok {
		text.WriteString("\\label{")
		text.WriteString(id)
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

func writeTexCodeFence(n *Node) []byte {
	var text bytes.Buffer
	text.WriteString(`\begin{codesnippet}
\begin{minted}[frame=single,breaklines]{c}
`)
	text.WriteString(n.Literal)
	text.WriteString(`\end{minted}
\end{codesnippet}

`)
	// это можно добавить для подписи
	// \caption{My func}\label{lst:my_func}

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
