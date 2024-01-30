package md2json

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func CopyImages(srcDir string, imageDir string, outDir string) error {
	for _, fp := range ListAllMd(srcDir) {
		doc, err := ReadJson(fp)
		if err != nil {
			return err
		}
		err = copyRequiredImages(&doc, fp, imageDir, outDir)
		if err != nil {
			return err
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func copyRequiredImages(n *Node, path string, imageDir string, outDir string) error {
	if n.Type == Image {
		src, ok := n.Attributes["src"]
		if !ok {
			return errors.New(fmt.Sprintf("File %s has image without src", path))
		}
		sourceFile := filepath.Join(imageDir, src)
		destFile := filepath.Join(outDir, src)
		imageBody, err := os.ReadFile(sourceFile)
		if err != nil {
			return errors.New(fmt.Sprintf("File %s has invalid link to image %s", path, src))
		}
		os.MkdirAll(filepath.Dir(destFile), os.ModePerm)
		err = os.WriteFile(destFile, imageBody, os.ModePerm)
		if err != nil {
			return err
		}
	}
	if n.Children != nil {
		for i := range n.Children {
			err := copyRequiredImages(&n.Children[i], path, imageDir, outDir)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
