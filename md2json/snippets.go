package md2json

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func CopySnippets(rootDir string) error {
	snippets := map[string]string{}

	var visitSnippets = func(fp string, fi os.DirEntry, err error) error {
		if fi.IsDir() {
			return nil
		}
		doc, err := ReadJson(fp)
		if err != nil {
			return err
		}
		var loadSnippets func(n Node)

		loadSnippets = func(n Node) {
			if n.Type == Snippet {
				snippets[n.Attributes["id"]] = n.Literal
			}
			if n.Children != nil {
				for _, ch := range n.Children {
					loadSnippets(ch)
				}
			}
		}
		loadSnippets(doc)
		return nil
	}
	err := filepath.WalkDir(rootDir, visitSnippets)
	if err != nil {
		return err
	}

	var includeSnippets = func(fp string, fi os.DirEntry, err error) error {
		if fi.IsDir() {
			return nil
		}
		doc, err := ReadJson(fp)
		dirty := false
		if err != nil {
			return err
		}
		var replaceSnippets func(n *Node) error

		replaceSnippets = func(n *Node) error {
			if n.Type == HTML && n.Attributes != nil {
				val, ok1 := n.Attributes["tag"]
				id, ok2 := n.Attributes["id"]
				if ok1 && ok2 && val == "include-snippet" {
					text, ok3 := snippets[id]
					if ok3 {
						n.Type = Snippet
						delete(n.Attributes, "tag")
						n.Literal = text
						dirty = true
					} else {
						return errors.New(fmt.Sprintf("failed to find snippet %s for file %s", id, fp))
					}
				}
			}
			if n.Children != nil {
				for i := range n.Children {
					err := replaceSnippets(&n.Children[i])
					if err != nil {
						return err
					}
				}
			}
			return nil
		}
		err = replaceSnippets(&doc)
		if err != nil {
			return err
		}
		if dirty {
			err = WriteJson(&doc, fp)
		}
		return err
	}
	err = filepath.WalkDir(rootDir, includeSnippets)
	return err
}
