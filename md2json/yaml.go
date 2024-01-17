package md2json

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type YamlFragment struct {
	content *yaml.Node
}

func (i *YamlFragment) UnmarshalYAML(value *yaml.Node) error {
	i.content = value
	return nil
}

type YamlIncludeProcessor struct {
	target interface{}
	dir    string
}

func (i *YamlIncludeProcessor) UnmarshalYAML(value *yaml.Node) error {
	resolved, err := resolveIncludes(value, i.dir)
	if err != nil {
		return err
	}
	return resolved.Decode(i.target)
}

func resolveIncludes(node *yaml.Node, dir string) (*yaml.Node, error) {
	if node.Tag == "!include" {
		if node.Kind != yaml.ScalarNode {
			return nil, errors.New("!include on a non-scalar node")
		}
		path := filepath.Join(dir, node.Value)
		file, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var f YamlFragment
		err = yaml.Unmarshal(file, &f)
		if err != nil {
			fmt.Printf("Failed to unmarshall %s: %v\n", path, err)
		}
		return f.content, err
	}
	if node.Kind == yaml.SequenceNode || node.Kind == yaml.MappingNode {
		var err error
		for i := range node.Content {
			node.Content[i], err = resolveIncludes(node.Content[i], dir)
			if err != nil {
				return nil, err
			}
		}
	}
	return node, nil
}

func YamlParse(path string) (map[string]interface{}, error) {
	obj := make(map[string]interface{})

	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, &YamlIncludeProcessor{target: &obj, dir: filepath.Dir(path)})
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func YamlWrite(obj interface{}, path string) error {
	yamlFile, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}
	os.MkdirAll(filepath.Dir(path), os.ModePerm)
	err = os.WriteFile(path, yamlFile, os.ModePerm)
	return err
}
