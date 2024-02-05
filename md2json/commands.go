package md2json

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type CommandFunction func([]string) error

var Commands = map[string]CommandFunction{
	"md2json":     Commmand_md2json,
	"planarize":   Command_planarize,
	"superlinks":  Command_superlinks,
	"snippets":    Command_snippets,
	"graphviz":    Command_graphviz,
	"macros":      Command_macros,
	"json2md":     Command_json2md,
	"lint":        Command_lint,
	"json2latex":  Command_json2latex,
	"heading":     Command_heading,
	"copy-images": Command_copyImages,
	"mkdocs":      Command_mkdocs,
}

func Command_mkdocs(args []string) error {
	if len(args) < 2 {
		return errors.New(fmt.Sprintf("usage: mkdocs input.yml output.yml"))
	}
	mkdocs, err := YamlParse(args[0])
	if err != nil {
		return err
	}
	err = YamlWrite(mkdocs, args[1])
	return err
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
	if len(args) < 2 {
		return errors.New(fmt.Sprintf("usage: planarize input_dir|input_mkdocs output_dir|output_mkdocs"))
	}
	if strings.HasSuffix(args[0], ".yml") {
		return PlanarizeMkdocs(args[0], args[1])
	}
	return PlanarizeDirectory(args[0], args[1])
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

func Command_json2md(args []string) error {
	if len(args) < 2 {
		return errors.New(fmt.Sprintf("usage: json2md input_dir output_dir"))
	}
	inDir := args[0]
	outDir := args[1]

	st, err := os.Stat(inDir)
	if err != nil {
		return err
	}
	if !st.IsDir() {
		os.MkdirAll(filepath.Dir(outDir), os.ModePerm)
		err := Json2Md(inDir, outDir)
		return err
	}

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
		return errors.New(fmt.Sprintf("usage: macros foliant.yml dir"))
	}
	return SubstituteMacrosFromFile(args[0], args[1])
}

func Command_graphviz(args []string) error {
	if len(args) < 3 {
		return errors.New(fmt.Sprintf("usage: grapviz srcDir imageDir /tmp"))
	}
	return Graphviz(args[0], args[1], args[2])
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
	input := args[0]
	output := args[1]
	args = args[2:]
	level := -1
	for len(args) > 0 {
		if args[0] == "addheading" {
			if len(args) < 2 {
				return errors.New(fmt.Sprintf("addheading level"))
			}
			level0, err := strconv.Atoi(args[1])
			if err != nil {
				return err
			}
			level = level0
			args = args[2:]
			continue
		}
		return errors.New(fmt.Sprintf("Unknown json2latex args %v", args))
	}

	doc, err := ReadJson(input)
	if err != nil {
		return err
	}
	if level != -1 {
		for _, n := range doc.Children {
			if n.Type == Heading {
				lvl, _ := n.Attributes["level"]
				lvl0, _ := strconv.Atoi(lvl)
				n.Attributes["level"] = fmt.Sprintf("%d", lvl0+level)
			}
		}
	}
	tex, err := Latex(&doc)
	if err != nil {
		return err
	}
	err = os.WriteFile(output, tex, os.ModePerm)
	return err
}

func Command_heading(args []string) error {
	if len(args) < 1 {
		return errors.New("usage: heading input")
	}
	doc, err := ReadJson(args[0])
	if err != nil {
		return err
	}
	title, _, found := doc.Heading()
	if !found {
		return errors.New(fmt.Sprintf("No heading in document: %s", args[0]))
	}
	fmt.Printf("%s\n", title)
	return nil
}

func Command_copyImages(args []string) error {
	if len(args) < 3 {
		return errors.New("usage: copy-images inputDoc inputImg outputImg")
	}
	err := CopyImages(args[0], args[1], args[2])
	return err
}
