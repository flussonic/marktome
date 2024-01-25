package md2json

func Foliant2Mkdocs(input string, output string) error {
	foliant, err := YamlParse(input)
	if err != nil {
		return err
	}

	backend_config, _ := foliant["backend_config"]
	mkdocs1, _ := backend_config.(map[string]interface{})["mkdocs"]
	mkdocs_, _ := mkdocs1.(map[string]interface{})["mkdocs.yml"]
	mkdocs := mkdocs_.(map[string]interface{})
	nav, _ := foliant["chapters"]

	mkdocs["nav"] = nav
	title, ok := foliant["title"]
	if ok {
		delete(foliant, "title")
		mkdocs["site_name"] = title
	}
	err = YamlWrite(mkdocs, output)
	return err
}
