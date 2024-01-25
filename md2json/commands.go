package md2json

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type CommandFunction func([]string) error

var Commands = map[string]CommandFunction{
	"md2json":        Commmand_md2json,
	"planarize":      Command_planarize,
	"superlinks":     Command_superlinks,
	"snippets":       Command_snippets,
	"graphviz":       Command_graphviz,
	"macros":         Command_macros,
	"foliant2mkdocs": Commmand_foliant2mkdocs,
	"json2md":        Command_json2md,
	"lint":           Command_lint,
	"json2latex":     Command_json2latex,
}

func Commmand_md2json(args []string) error {
	if len(args) < 2 {
		return errors.New(fmt.Sprintf("usage: md2json input_dir output_dir"))
	}
	rootDir := args[0]
	outDir := args[1]
	st, err := os.Stat(rootDir)
	if err != nil {
		return err
	}
	if !st.IsDir() {
		os.MkdirAll(filepath.Dir(outDir), os.ModePerm)
		err := Md2Json(rootDir, outDir)
		return err
	}
	paths := ListAllMd(rootDir)
	for _, fp := range paths {
		output := outDir + "/" + strings.TrimPrefix(fp, rootDir)
		os.MkdirAll(filepath.Dir(output), os.ModePerm)
		err := Md2Json(fp, output)
		if err != nil {
			return err
		}
	}
	return nil
}

func Command_planarize(args []string) error {
	if len(args) < 1 {
		return errors.New(fmt.Sprintf("usage: planarize dir"))
	}
	return Planarize(args[0])
}

func Command_superlinks(args []string) error {
	if len(args) < 1 {
		return errors.New(fmt.Sprintf("usage: superlinks dir"))
	}
	return CrosscheckSuperlinks(args[0])
}

func Command_snippets(args []string) error {
	if len(args) < 1 {
		return errors.New(fmt.Sprintf("usage: snippets dir"))
	}
	return CopySnippets(args[0])
}

func Commmand_foliant2mkdocs(args []string) error {
	if len(args) < 2 {
		return errors.New(fmt.Sprintf("usage: foliant2mkdocs input_dir output_dir"))
	}
	return Foliant2Mkdocs(args[0], args[1])
}

func Command_json2md(args []string) error {
	if len(args) < 2 {
		return errors.New(fmt.Sprintf("usage: json2md input_dir output_dir"))
	}
	inDir := args[0]
	outDir := args[1]
	for _, out := range ListAllMd(inDir) {
		out2 := outDir + "/" + strings.TrimPrefix(out, inDir+"/")
		os.MkdirAll(filepath.Dir(out2), os.ModePerm)
		err := Json2Md(out, out2)
		if err != nil {
			return err
		}
	}
	return nil
}

func Command_macros(args []string) error {
	if len(args) < 2 {
		return errors.New(fmt.Sprintf("usage: macros dir path-to-macros.yml"))
	}
	rootDir := args[0]
	macros := make(map[string]string)

	macrosFile, err := YamlParse(args[1])
	if err != nil {
		return err
	}
	preprocessors, _ := macrosFile["preprocessors"]
	for _, elem := range preprocessors.([]interface{}) {
		if reflect.TypeOf(elem).Kind() == reflect.Map {
			keys := reflect.ValueOf(elem).MapKeys()
			if len(keys) == 1 && keys[0].Interface().(string) == "macros" {
				macros1, _ := elem.(map[string]interface{})["macros"]
				macros2, _ := macros1.(map[string]interface{})["macros"]
				for k, v := range macros2.(map[string]interface{}) {
					macros[k] = v.(string)
				}
			}
		}
	}
	return SubstituteMacros(rootDir, macros)
}

func Command_graphviz(args []string) error {
	if len(args) < 2 {
		return errors.New(fmt.Sprintf("usage: grapviz srcDir imageDir"))
	}
	return Graphviz(args[0], args[1])
}

func Command_lint(args []string) error {
	if len(args) < 1 {
		return errors.New(fmt.Sprintf("usage: lint file.md"))
	}
	f, err := os.CreateTemp("/tmp", "md2json-")
	if err != nil {
		return err
	}
	defer f.Close()
	err = Md2Json(args[0], f.Name())
	if err != nil {
		return err
	}
	err = Json2Md(f.Name(), args[0])
	return err
}

func Command_json2latex(args []string) error {
	if len(args) < 2 {
		return errors.New(fmt.Sprintf("usage: json2latex input_dir output.tex"))
	}

	doc, err := MergeDocument(args[0])
	if err != nil {
		return err
	}
	tex, err := Latex(doc)
	if err != nil {
		return err
	}
	err = os.WriteFile(args[1], tex, os.ModePerm)
	return err
}
