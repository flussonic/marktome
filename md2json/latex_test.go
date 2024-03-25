package md2json_test

import (
	"marktome/md2json"
	"path/filepath"
	"os"
	"testing"
	"strings"
	"bytes"
)

type LatexTest struct {
	name     string
	input    []byte
	expected []byte
}


func TestLatex(t *testing.T) {
	tests := []LatexTest{}

	var visit = func(fp string, fi os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !fi.IsDir() && strings.HasSuffix(fp, ".md") {
			tt := LatexTest{}
			tt.name = strings.TrimSuffix(fp, ".md")
			input, _ := os.ReadFile(fp)
			expected, _ := os.ReadFile(tt.name + ".tex")
			tt.input = input
			tt.expected = expected
			tests = append(tests, tt)
		}
		return nil
	}
	filepath.WalkDir("testdata/latex", visit)

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			doc := md2json.MarkdownParse(tt.input)

			tex, err := md2json.Latex(&doc)
			if err != nil {
				t.Error(err)
			}

			if bytes.Compare(tex, tt.expected) != 0 {
				t.Errorf("Latex()\nactual\n%s\nexpected\n%s", tex, tt.expected)
			}
		})
	}

}