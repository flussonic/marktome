package md2json

import "reflect"

func FoliantMacros(path string) (map[string]string, error) {
	macros := make(map[string]string)

	macrosFile, err := YamlParse(path)
	if err != nil {
		return macros, err
	}
	preprocessors, _ := macrosFile["preprocessors"]
	for _, elem := range preprocessors.([]interface{}) {
		if reflect.TypeOf(elem).Kind() == reflect.Map {
			keys := reflect.ValueOf(elem).MapKeys()
			if len(keys) == 1 && keys[0].Interface().(string) == "macros" {
				macros1, _ := elem.(map[string]interface{})["macros"]
				macros2, _ := macros1.(map[string]interface{})["macros"]
				for k, v := range macros2.(map[string]interface{}) {
					macros[k] = v.(string)
				}
			}
		}
	}
	return macros, nil
}
