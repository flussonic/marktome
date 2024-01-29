package md2json

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func replaceGraphviz(imageDir string, n *Node, path string, cachePath string) (bool, error) {
	dirty := false
	if n.Type == HTML && n.Attributes != nil {
		tag, ok1 := n.Attributes["tag"]
		if ok1 && tag == "graphviz" {
			hash := md5.Sum([]byte(n.Literal))
			id := strings.TrimSuffix(filepath.Base(path), ".md")
			imagePath := id + "-" + hex.EncodeToString(hash[:]) + ".png"
			cachePath := filepath.Join(cachePath, imagePath)
			graphPath := filepath.Join(cachePath, hex.EncodeToString(hash[:])+".vg")

			if _, err := os.Stat(cachePath); err != nil {
				err = os.WriteFile(graphPath, []byte(n.Literal), os.ModePerm)
				if err != nil {
					return false, err
				}
				command := "dot -Tpng -Kdot -o " + cachePath + " " + graphPath
				cmd := exec.Command("/bin/bash", "-c", command)
				_, err := cmd.Output()
				if err != nil {
					return false, err
				}
			} else {
				fmt.Printf("Skip generating %s\n", imagePath)
			}
			if _, err := os.Stat(cachePath); err != nil {
				return false, errors.New(fmt.Sprintf("Failed to create graph from %s %s", path, graphPath))
			}
			img, err := os.ReadFile(cachePath)
			if err != nil {
				return false, err
			}
			fullImagePath := filepath.Join(imageDir, imagePath)
			os.MkdirAll(filepath.Dir(fullImagePath), os.ModePerm)
			err = os.WriteFile(fullImagePath, img, os.ModePerm)
			if err != nil {
				return false, err
			}
			n.Type = Paragraph
			n.Literal = ""
			n.Attributes = map[string]string{}
			n.Children = []Node{
				{
					Type: Image,
					Attributes: map[string]string{
						"src": "img/" + imagePath,
					},
				},
			}
			// n.Type = Image
			// n.Literal = ""
			// delete(n.Attributes, "tag")
			// n.Attributes["src"] = imagePath
			dirty = true
		}
	}
	if n.Children != nil {
		for i := range n.Children {
			d, err := replaceGraphviz(imageDir, &n.Children[i], path, cachePath)
			if err != nil {
				return false, err
			}
			dirty = d || dirty
		}
	}
	return dirty, nil
}

func Graphviz(rootDir string, imageDir string, cachePath string) error {
	for _, fp := range ListAllMd(rootDir) {
		doc, err := ReadJson(fp)
		var dirty bool
		if err != nil {
			return err
		}
		dirty, err = replaceGraphviz(imageDir, &doc, fp, cachePath)
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
