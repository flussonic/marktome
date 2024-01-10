package main

import (
	"fmt"
	"foli2/md2json"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Convert1(input string, output string) error {
	file, err := os.Open(input)
	if err != nil {
		return err
	}
	defer file.Close()

	// Создаем буфер для чтения из файла
	source, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	fmt.Println("Convert", input)
	json := md2json.ParseDocument(source)
	md2json.WriteJson(&json, output)
	return nil
}

func main() {
	if true {
		rootDir := "../erlydoc/src/" // замените на путь к нужному вам каталогу
		var visitFile = func(fp string, fi os.DirEntry, err error) error {
			if fi.IsDir() {
				return nil
			}
			if !strings.HasSuffix(fp, ".md") {
				return nil
			}

			output := "tmp/" + strings.TrimPrefix(fp, rootDir)
			os.MkdirAll(filepath.Dir(output), os.ModePerm)
			err = Convert1(fp, output)
			// output2 := strings.TrimSuffix(output, ".md") + ".json"
			// cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("cat %s | jq > %s", output, output2))
			// cmd.Run()
			// os.Remove(output)
			return err
		}
		err := filepath.WalkDir(rootDir, visitFile)
		if err != nil {
			fmt.Printf("Ошибка при обходе каталога: %v\n", err)
		}
		md2json.CrosscheckSuperlinks("tmp/en")
		md2json.CrosscheckSuperlinks("tmp/ru")
	}
	if true {
		Convert1("live.md", "output.txt")
		// Convert1("../erlydoc/src/en/watcher/authorization-backend.md", "output.txt")
	}
}
