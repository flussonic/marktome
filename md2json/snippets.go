package md2json

import (
	"errors"
	"fmt"
)

func CopySnippets(rootDir string) error {
	snippets := map[string]string{}

	var loadSnippets func(n *Node, path string) error

	loadSnippets = func(n *Node, path string) error {
		if n.Type == HTML && n.Attributes != nil {
			tag, ok1 := n.Attributes["tag"]
			if ok1 && tag == "snippet" {
				id, ok2 := n.Attributes["id"]
				if ok2 {
					snippets[id] = n.Literal
				} else {
					return errors.New(fmt.Sprintf("File %s has snippet without id", path))
				}
			}
		}
		if n.Children != nil {
			for _, ch := range n.Children {
				err := loadSnippets(&ch, path)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	var replaceSnippets func(n *Node, fp string) (bool, error)

	replaceSnippets = func(n *Node, fp string) (bool, error) {
		dirty := false
		if n.Type == HTML && n.Attributes != nil {
			tag, ok1 := n.Attributes["tag"]
			id, ok2 := n.Attributes["id"]
			if ok1 && ok2 && tag == "snippet" {
				n.Type = CodeFence
				delete(n.Attributes, "tag")
				dirty = true
			}
			if ok1 && ok2 && tag == "include-snippet" {
				text, ok3 := snippets[id]
				if ok3 {
					n.Type = CodeFence
					delete(n.Attributes, "tag")
					n.Literal = text
					dirty = true
				} else {
					return false, errors.New(fmt.Sprintf("failed to find snippet %s for file %s", id, fp))
				}
			} else {
				// fmt.Printf("Strange HTML %v\n", n)
			}
		}
		if n.Children != nil {
			for i := range n.Children {
				d, err := replaceSnippets(&n.Children[i], fp)
				if err != nil {
					return false, err
				}
				dirty = d || dirty
			}
		}
		return dirty, nil
	}

	paths := ListAllMd(rootDir)
	for _, fp := range paths {
		doc, err := ReadJson(fp)
		if err != nil {
			return err
		}
		err = loadSnippets(&doc, fp)
		if err != nil {
			return err
		}
	}

	for _, fp := range paths {
		doc, err := ReadJson(fp)
		var dirty bool
		if err != nil {
			return err
		}
		dirty, err = replaceSnippets(&doc, fp)
		if err != nil {
			return err
		}
		if dirty {
			err = WriteJson(&doc, fp)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
