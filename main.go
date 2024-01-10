package main

import (
	"encoding/json"
	"fmt"
	"foli2/md2json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

func dumpFile(input string) error {
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
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			meta.New(meta.WithTable()),
		),
		goldmark.WithParserOptions(
			parser.WithAttribute(),
		),
	)
	// if err := md.Convert(source, &buf); err != nil {
	// 	return []byte(""), err
	// }

	reader := text.NewReader(source)
	document := md.Parser().Parse(reader)
	document.Dump(source, 0)
	return nil
}

func writeJson(root *md2json.Node, path string) {
	jsonData, err := json.Marshal(root)
	if err != nil {
		return
	}

	output, err := os.Create(path)
	if err != nil {
		fmt.Println("Ошибка при создании файла:", err)
		return
	}
	defer output.Close()

	// Копируем данные из Buffer в файл
	_, err = output.Write(jsonData)
	if err != nil {
		fmt.Println("Ошибка при записи данных в файл:", err)
		return
	}
}

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
	json := md2json.ParseDocument(source)
	writeJson(&json, output)
	return nil
}

func main() {
	if false {
		rootDir := "../erlydoc/src/" // замените на путь к нужному вам каталогу
		var visitFile = func(fp string, fi os.DirEntry, err error) error {
			if fi.IsDir() {
				return nil
			}
			output := "tmp/" + strings.TrimPrefix(fp, rootDir)
			os.MkdirAll(filepath.Dir(output), os.ModePerm)
			err = Convert1(fp, output)
			return err
		}
		err := filepath.WalkDir(rootDir, visitFile)
		if err != nil {
			fmt.Printf("Ошибка при обходе каталога: %v\n", err)
		}
	}
	if true {
		Convert1("live.md", "output.txt")
	}
	if false {
		dumpFile("../erlydoc/src/ru/webrtc.md")
	}

}
