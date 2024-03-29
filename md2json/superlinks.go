package md2json

import (
	"errors"
	"fmt"
	"strings"
)

func calculateRelativeLocation(origin string, target string) string {
	return strings.TrimPrefix(target, "/") + ".md"
	// origin = "/" + strings.TrimSuffix(strings.TrimPrefix(origin, "/"), "/") + "/"
	// target = "/" + strings.TrimSuffix(strings.TrimPrefix(target, "/"), "/") + "/"
	// rel, _ := filepath.Rel(origin, target)
	// if rel == "." {
	// 	return ""
	// }
	// return rel + "/"
}

func CrosscheckSuperlinks(rootDir string) error {
	headings := map[string]string{}

	for _, fp := range ListAllMd(rootDir) {
		origName := strings.TrimSuffix(strings.TrimPrefix(fp, rootDir), ".md")
		doc, _ := ReadJson(fp)
		var readHeadings func(n Node) error

		readHeadings = func(n Node) error {
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
					e := readHeadings(ch)
					if e != nil {
						return e
					}
				}
			}
			return nil
		}
		err := readHeadings(doc)
		if err != nil {
			return err
		}
	}

	for _, fp := range ListAllMd(rootDir) {
		dirty := false
		doc, _ := ReadJson(fp)
		origName := strings.TrimSuffix(strings.TrimPrefix(fp, rootDir), ".md")

		var checkAnchors func(n *Node) error
		checkAnchors = func(n *Node) error {
			if (n.Type == "HTML" || n.Type == Link) && n.Attributes != nil {
				tag, ok1 := n.Attributes["tag"]
				anchor, ok2 := n.Attributes["anchor"]
				_, ok3 := n.Attributes["href"]
				if ((ok1 && tag == "link") || n.Type == Link) && ok2 {
					n.Type = Link
					delete(n.Attributes, "tag")
					dirty = true
					if !ok3 {
						location, ok := headings[anchor]
						if ok {
							rel := calculateRelativeLocation(origName, location)
							if rel != "" {
								n.Attributes["href"] = rel
							}
						} else {
							return errors.New(fmt.Sprintf("Anchor %s in file %s not found in project", anchor, fp))
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
		err := checkAnchors(&doc)
		if err != nil {
			return err
		}
		if dirty {
			err = WriteJson(&doc, fp)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
