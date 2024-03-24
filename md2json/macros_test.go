package md2json_test

import (
	"log"
	"marktome/md2json"
	"os"
	"testing"
)

func TestMacros(t *testing.T) {
	dir1, err := os.MkdirTemp(".", "test-macros-src")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir1)

	dir2, err := os.MkdirTemp(".", "test-macros-dest")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir2)

	d1, _ := os.ReadFile("testdata/macros.md")
	d2, _ := os.ReadFile("testdata/macros.yml")
	os.WriteFile(dir1+"/macros.md", d1, os.ModePerm)
	os.WriteFile(dir1+"/macros.yml", d2, os.ModePerm)
	md2json.SubstituteMacrosFromFile(dir1+"/macros.yml", dir1, dir2)

	d3, _ := os.ReadFile(dir2 + "/macros.md")
	if string(d3) != "[Link](www.example.com/path)\n" {
		t.Errorf("Macros failure: %s", d3)
	}
}
