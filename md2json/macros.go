package md2json

import (
	"bytes"
	"errors"
	"fmt"
	"os"
)

func replaceMacros(src []byte, macros map[string]string, fp string) ([]byte, bool, error) {
	dirty := false
	var output bytes.Buffer

	open := []byte("<m>")
	close := []byte("</m>")

	for len(src) > 0 {
		i := bytes.Index(src, open)
		if i < 0 {
			output.Write(src)
			return output.Bytes(), dirty, nil
		}
		idx := bytes.Index(src[i:], close)
		if idx < 0 {
			return output.Bytes(), false, errors.New(fmt.Sprintf("Unmatched closing <m> tag in file %s", fp))
		}
		name := src[i+len(open) : i+idx]
		value, ok2 := macros[string(name)]
		if !ok2 {
			return output.Bytes(), false, errors.New(fmt.Sprintf("Not found macro '%s' in file %s", name, fp))
		}
		output.Write(src[:i])
		output.Write([]byte(value))
		dirty = true
		src = src[i+idx+len(close):]
	}
	return output.Bytes(), dirty, nil
}

func SubstituteMacros(rootDir string, macros map[string]string) error {

	paths := ListAllMd(rootDir)
	for _, fp := range paths {
		doc, err := os.ReadFile(fp)
		var dirty bool
		var doc2 []byte
		if err != nil {
			return err
		}
		doc2, dirty, err = replaceMacros(doc, macros, fp)
		if err != nil {
			return err
		}
		if dirty {
			err = os.WriteFile(fp, doc2, os.ModePerm)
		}
		if err != nil {
			return err
		}
	}
	return nil
}