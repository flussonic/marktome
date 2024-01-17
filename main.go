package main

import (
	"fmt"
	"foli2/md2json"
	"os"
	"path/filepath"
	"strings"
)

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

			output := "tmp2/" + strings.TrimPrefix(fp, rootDir)
			os.MkdirAll(filepath.Dir(output), os.ModePerm)
			// fmt.Println("md2json", fp)
			err = md2json.Md2Json(fp, output)
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
		fmt.Println("Planarize ru")
		err = md2json.Planarize("tmp2/ru")
		if err != nil {
			fmt.Printf("Rename tmp2/ru error: %v\n", err)
		}
		fmt.Println("Planarize en")
		err = md2json.Planarize("tmp2/en")
		if err != nil {
			fmt.Printf("Rename tmp2/en error: %v\n", err)
		}
		fmt.Println("Superlinks")
		md2json.CrosscheckSuperlinks("tmp2/en")
		md2json.CrosscheckSuperlinks("tmp2/ru")
		fmt.Println("CopySnippets")
		err = md2json.CopySnippets("tmp2")
		if err != nil {
			fmt.Printf("Snippets tmp2 error: %v\n", err)
		}
		err = md2json.Foliant2Mkdocs("../erlydoc/f2/foliant.flussonic.en.yml", "tmp2/en/mkdocs.yml")
		if err != nil {
			fmt.Println("Failed to write mkdocs.yml", err)
		}
		err = convertJson2Md("tmp2", "tmp")
		if err != nil {
			fmt.Printf("Write json error: %v\n", err)
		}
	}
	if false {
		err := md2json.Foliant2Mkdocs("../erlydoc/f2/foliant.flussonic.en.yml", "tmp/en/mkdocs.yml")
		if err != nil {
			fmt.Println("Failed to write mkdocs.yml", err)
		}

	}
	if false {
		md2json.Md2Json("live.md", "output.txt")
		// Convert1("../erlydoc/src/en/watcher/authorization-backend.md", "output.txt")
	}
	if true {
		md2json.Md2Json("../erlydoc/src/en/transcoder.md", "output.txt")
	}
}

func convertJson2Md(inDir string, outDir string) error {
	paths := md2json.ListAllMd(inDir)
	for _, out := range paths {
		out2 := outDir + "/" + strings.TrimPrefix(out, inDir+"/")
		os.MkdirAll(filepath.Dir(out2), os.ModePerm)
		err := md2json.Json2Md(out, out2)
		if err != nil {
			return err
		}
	}
	return nil
}
