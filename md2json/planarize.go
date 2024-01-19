package md2json

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Planarize(rootDir string) error {
	paths := ListAllMd(rootDir)

	for _, fp := range paths {
		doc, err := ReadJson(fp)
		if err != nil {
			return err
		}

		heading, id, found := doc.Heading()
		if !found {
			return errors.New(fmt.Sprintf("No heading and title in %s", fp))
		}
		if doc.Attributes == nil {
			doc.Attributes = map[string]string{}
		}
		original := strings.TrimSuffix(strings.TrimPrefix(fp, rootDir+"/"), ".md")
		if original == "index" {
			continue
		}
		doc.Attributes["original"] = original
		doc.Attributes["title"] = heading

		output := fmt.Sprintf("%s/%s.md", rootDir, id)

		os.MkdirAll(filepath.Dir(output), os.ModePerm)
		err = WriteJson(&doc, output)
		if err != nil {
			return err
		}
		if fp != output {
			os.Remove(fp)
			os.Remove(filepath.Dir(fp))
		}
	}
	return nil
}
