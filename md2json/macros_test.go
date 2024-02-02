package md2json_test

import (
	"log"
	"marktome/md2json"
	"os"
	"testing"
)

func TestMacros(t *testing.T) {
	dir, err := os.MkdirTemp(".", "test-macros")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	d1, _ := os.ReadFile("testdata/macros.md")
	d2, _ := os.ReadFile("testdata/macros.yml")
	os.WriteFile(dir+"/macros.md", d1, os.ModePerm)
	os.WriteFile(dir+"/macros.yml", d2, os.ModePerm)
	md2json.SubstituteMacrosFromFile(dir+"/macros.yml", dir)

	d3, _ := os.ReadFile(dir + "/macros.md")
	if string(d3) != "[Link](www.example.com/path)\n" {
		t.Errorf("Macros failure\n%s", d3)
	}
}
