package md2json

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
)

func Foliant2Mkdocs(input string, output string) error {
	foliant, err := YamlParse(input)
	if err != nil {
		return err
	}

	renames, err := loadRenames(filepath.Dir(output))

	backend_config, _ := foliant["backend_config"]
	mkdocs1, _ := backend_config.(map[string]interface{})["mkdocs"]
	mkdocs_, _ := mkdocs1.(map[string]interface{})["mkdocs.yml"]
	mkdocs := mkdocs_.(map[string]interface{})
	nav, _ := foliant["chapters"]

	var renameChapters func(menu interface{})
	renameChapters = func(menu interface{}) {
		if reflect.TypeOf(menu).Kind() == reflect.Slice {
			for i, v := range menu.([]interface{}) {
				if reflect.TypeOf(v).Kind() == reflect.String {
					newPath, ok := renames[v.(string)]
					if ok {
						fmt.Printf("Renaming %d %s to %s\n", i, v, newPath)
						menu.([]interface{})[i] = newPath
					}
				}
				if reflect.TypeOf(v).Kind() == reflect.Map {
					renameChapters(v)
				}
				if reflect.TypeOf(v).Kind() == reflect.Slice {
					renameChapters(v)
				}
			}
		}
		if reflect.TypeOf(menu).Kind() == reflect.Map {
			for k, v := range menu.(map[string]interface{}) {
				if reflect.TypeOf(v).Kind() == reflect.String {
					newPath, ok := renames[v.(string)]
					if ok {
						menu.(map[string]interface{})[k] = newPath
					}
				}
				if reflect.TypeOf(v).Kind() == reflect.Map {
					renameChapters(v)
				}
				if reflect.TypeOf(v).Kind() == reflect.Slice {
					renameChapters(v)
				}
			}
		}
	}
	renameChapters(nav)
	mkdocs["nav"] = nav
	err = YamlWrite(mkdocs, output)
	return err
}

func loadRenames(rootDir string) (map[string]string, error) {
	renames := make(map[string]string)
	paths := ListAllMd(rootDir)
	for _, p := range paths {
		doc, err := ReadJson(p)
		if err != nil {
			return nil, err
		}
		if doc.Attributes != nil {
			original, ok := doc.Attributes["original"]
			if ok {
				renames[original+".md"] = strings.TrimPrefix(p, rootDir+"/")
			}
		}
	}
	return renames, nil
}
