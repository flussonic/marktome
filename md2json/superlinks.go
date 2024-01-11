package md2json

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func calculateRelativeLocation(origin string, target string) string {
	origin = "/" + strings.TrimSuffix(strings.TrimPrefix(origin, "/"), "/") + "/"
	target = "/" + strings.TrimSuffix(strings.TrimPrefix(target, "/"), "/") + "/"
	rel, _ := filepath.Rel(origin, target)
	if rel == "." {
		return ""
	}
	return rel + "/"
}

func CrosscheckSuperlinks(rootDir string) error {
	headings := map[string]string{}
	var visitHeadings = func(fp string, fi os.DirEntry, err error) error {
		if fi.IsDir() {
			return nil
		}

		origName := strings.TrimSuffix(strings.TrimPrefix(fp, rootDir), ".md")
		doc, _ := ReadJson(fp)
		var addHeadings func(n Node) error

		addHeadings = func(n Node) error {
			if n.Type == Heading && n.Attributes != nil {
				val, ok := n.Attributes["id"]
				if ok {
					old, ok2 := headings[val]
					if ok2 {
						return errors.New(fmt.Sprintf("Heading %s double declared in %s and %s\n", val, origName, old))
					}
					headings[val] = origName
				}
			}
			if n.Children != nil {
				for _, ch := range n.Children {
					e := addHeadings(ch)
					if e != nil {
						return e
					}
				}
			}
			return nil
		}
		err = addHeadings(doc)
		return err
	}
	filepath.WalkDir(rootDir, visitHeadings)

	var validateAndRewriteAnchors = func(fp string, fi os.DirEntry, err error) error {
		if fi.IsDir() {
			return nil
		}

		dirty := false
		doc, _ := ReadJson(fp)
		origName := strings.TrimSuffix(strings.TrimPrefix(fp, rootDir), ".md")

		var checkAnchors func(n *Node) error
		checkAnchors = func(n *Node) error {
			if n.Type == "HTML" && n.Attributes != nil {
				tag, ok1 := n.Attributes["tag"]
				anchor, ok2 := n.Attributes["anchor"]
				if ok1 && ok2 && tag == "link" {
					n.Type = Link
					delete(n.Attributes, "tag")
					if location, ok := headings[anchor]; ok {
						rel := calculateRelativeLocation(origName, location)
						if rel != "" {
							n.Attributes["href"] = rel
							dirty = true
						}
					}
				}
			}
			if n.Children != nil {
				for i := range n.Children {
					e := checkAnchors(&n.Children[i])
					if e != nil {
						return e
					}
				}
			}
			return nil
		}
		err = checkAnchors(&doc)
		if dirty {
			WriteJson(&doc, fp)
		}
		return err
	}
	err := filepath.WalkDir(rootDir, validateAndRewriteAnchors)
	return err
}
