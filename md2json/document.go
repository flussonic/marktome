package md2json

type Kind string
type AttributeMap map[string]string

const (
	Document   Kind = "Document"
	Paragraph  Kind = "Paragraph"
	Text       Kind = "Text"
	Image      Kind = "Image"
	HTML       Kind = "HTML"
	Heading    Kind = "Heading"
	Link       Kind = "Link"
	List       Kind = "List"
	ListItem   Kind = "ListItem"
	Admonition Kind = "Admonition"
	Code       Kind = "Code"
	CodeFence  Kind = "CodeFence"
	Bold       Kind = "Bold"
	Emphasis   Kind = "Emphasis"
	Table      Kind = "Table"
	TableHead  Kind = "THead"
	TableBody  Kind = "TBody"
	TableRow   Kind = "Row"
	TableCell  Kind = "Cell"
)

type Node struct {
	Type       Kind         `json:"type"`
	Children   []Node       `json:"children,omitempty"`
	Literal    string       `json:"text,omitempty"`
	Attributes AttributeMap `json:"attributes,omitempty"`
}

func (self *Node) Heading() (string, string, bool) {
	var searchHeading func(n *Node) (string, string, bool)

	searchHeading = func(n *Node) (string, string, bool) {
		if n.Type == Heading && n.Attributes != nil {
			level, ok1 := n.Attributes["level"]
			id, ok2 := n.Attributes["id"]
			if ok1 && ok2 && level == "1" {
				return n.Literal, id, true
			}
		}
		if n.Children != nil {
			for i := range n.Children {
				val1, val2, found := searchHeading(&n.Children[i])
				if found {
					return val1, val2, found
				}
			}
		}
		return "", "", false
	}
	return searchHeading(self)
}
