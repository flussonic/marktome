package md2json_test

import (
	"encoding/json"
	"marktome/md2json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type MdTest struct {
	name     string
	input    []byte
	expected []byte
}

func TestMarkdownParse(t *testing.T) {
	tests := []MdTest{}

	var visit = func(fp string, fi os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !fi.IsDir() && strings.HasSuffix(fp, ".md") {
			t := MdTest{}
			t.name = strings.TrimSuffix(fp, ".md")
			input, _ := os.ReadFile(fp)
			expected, _ := os.ReadFile(t.name + ".json")
			t.input = input
			t.expected = expected
			tests = append(tests, t)
		}
		return nil
	}
	filepath.WalkDir("testdata/parser", visit)

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			doc := md2json.Node{}
			err := json.Unmarshal(tt.expected, &doc)
			if err != nil {
				t.Error(err)
			}
			expected_, err := json.Marshal(doc)
			expected := string(expected_)
			if err != nil {
				t.Error(err)
			}

			result := md2json.MarkdownParse(tt.input)
			outJson_, err := json.Marshal(result)
			outJson := string(outJson_)
			if err != nil {
				t.Error(err)
			}

			if outJson != expected {
				t.Errorf("Parse()\nactual\n%s\nexpected\n%s", outJson, expected)
			}

			outMd := md2json.WriteDocument(&doc)
			if string(outMd) != string(tt.input) && !strings.Contains(tt.name, "dirty") {
				t.Errorf("WriteMd()\nactual\n%s\nexpected\n%s", outMd, tt.input)
			}

		})
	}

}
