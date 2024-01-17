package md2json

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ReadJson(input string) (Node, error) {
	file, err := os.Open(input)
	if err != nil {
		return Node{}, err
	}
	defer file.Close()

	// Создаем буфер для чтения из файла
	source, err := io.ReadAll(file)
	if err != nil {
		return Node{}, err
	}

	doc := Node{}
	json.Unmarshal(source, &doc)
	return doc, nil
}

func WriteJson(root *Node, path string) error {
	jsonData, err := json.Marshal(root)
	if err != nil {
		return err
	}

	output, err := os.Create(path)
	if err != nil {
		return err
	}
	defer output.Close()

	_, err = output.Write(jsonData)
	if err != nil {
		return err
	}
	return nil
}

func ListAllMd(rootDir string) []string {
	files := []string{}

	var visit = func(fp string, fi os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !fi.IsDir() && strings.HasSuffix(fp, ".md") {
			files = append(files, fp)
		}
		return nil
	}
	filepath.WalkDir(rootDir, visit)
	return files
}

func Slugify(s string) string {
	s = strings.ReplaceAll(s, "(", "")
	s = strings.ReplaceAll(s, ")", "")
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, ",", "-")
	return s
}
