package md2json

import "fmt"

func Mkdocs(path string) error {
	obj, err := YamlParse(path)
	fmt.Println(obj)
	return err
}
