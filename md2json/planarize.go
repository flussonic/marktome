package md2json

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

func PlanarizeMkdocs(input string, output string) error {
	foliant, err := YamlParse(input)
	if err != nil {
		return err
	}
	srcDir, ok := foliant["docs_dir"]
	if !ok {
		return errors.New(fmt.Sprintf("No docs_dir in foliant config %s", input))
	}
	nav, ok := foliant["nav"]
	if !ok {
		return errors.New(fmt.Sprintf("No nav in foliant config %s", input))
	}
	inDir := filepath.Join(filepath.Dir(input), srcDir.(string))
	paths := ListAllMd(inDir)
	outDir := filepath.Join(filepath.Dir(output), srcDir.(string))
	renames, err := Planarize(inDir, outDir, paths)

	err = renameChapters(nav, renames)
	if err != nil {
		return err
	}
	err = YamlWrite(foliant, output)
	return err
}

// var renameChapters func(menu interface{}, renames map[string]string) error

func renameChapters(menu interface{}, renames map[string]string) error {
	if reflect.TypeOf(menu).Kind() == reflect.Slice {
		for i, v := range menu.([]interface{}) {
			if reflect.TypeOf(v).Kind() == reflect.String {
				if !strings.HasSuffix(v.(string), ".md") {
					continue
				}
				newPath, ok := renames[v.(string)]
				if ok {
					menu.([]interface{})[i] = newPath
				} else {
					return errors.New(fmt.Sprintf("Unknown menu item %d %s", i, v))
				}
			}
			if reflect.TypeOf(v).Kind() == reflect.Map || reflect.TypeOf(v).Kind() == reflect.Slice {
				err := renameChapters(v, renames)
				if err != nil {
					return err
				}
			}
		}
	}
	if reflect.TypeOf(menu).Kind() == reflect.Map {
		for k, v := range menu.(map[string]interface{}) {
			if reflect.TypeOf(v).Kind() == reflect.String {
				if !strings.HasSuffix(v.(string), ".md") {
					continue
				}
				newPath, ok := renames[v.(string)]
				if ok {
					menu.(map[string]interface{})[k] = newPath
				} else {
					return errors.New(fmt.Sprintf("Unknown menu item '%s' -> '%s'", k, v))
				}
			}
			if reflect.TypeOf(v).Kind() == reflect.Map || reflect.TypeOf(v).Kind() == reflect.Slice {
				err := renameChapters(v, renames)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func PlanarizeDirectory(inDir string, outDir string) error {
	paths := ListAllMd(inDir)
	_, err := Planarize(inDir, outDir, paths)
	return err
}

func Planarize(inDir string, outDir string, paths []string) (map[string]string, error) {
	renames := make(map[string]string)

	for _, fp := range paths {
		doc, err := ReadJson(fp)
		if err != nil {
			return renames, err
		}

		heading, id, found := doc.Heading()
		if !found {
			return renames, errors.New(fmt.Sprintf("No heading and title in %s", fp))
		}
		if doc.Attributes == nil {
			doc.Attributes = map[string]string{}
		}
		original := strings.TrimSuffix(strings.TrimPrefix(fp, inDir+"/"), ".md")
		doc.Attributes["original"] = original
		doc.Attributes["title"] = heading

		output := fmt.Sprintf("%s/%s.md", outDir, id)
		renames[original+".md"] = id + ".md"

		os.MkdirAll(filepath.Dir(output), os.ModePerm)
		err = WriteJson(&doc, output)
		if err != nil {
			return renames, err
		}
	}
	return renames, nil
}
