package md2json

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

func Commmand_md2json(args []string) error {
	if len(args) < 2 {
		return errors.New(fmt.Sprintf("usage: md2json input_dir output_dir"))
	}
	rootDir := args[0]
	outDir := args[1]
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
