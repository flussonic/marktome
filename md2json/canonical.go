package md2json

import (
	"os"
	"path/filepath"
	"strings"
)

func Transliterate(title string) string {
	title = strings.ToLower(title)
	title = strings.ReplaceAll(title, "а", "a")
	title = strings.ReplaceAll(title, "б", "b")
	title = strings.ReplaceAll(title, "в", "v")
	title = strings.ReplaceAll(title, "г", "g")
	title = strings.ReplaceAll(title, "д", "d")
	title = strings.ReplaceAll(title, "е", "e")
	title = strings.ReplaceAll(title, "ё", "yo")
	title = strings.ReplaceAll(title, "ж", "zh")
	title = strings.ReplaceAll(title, "з", "z")
	title = strings.ReplaceAll(title, "и", "i")
	title = strings.ReplaceAll(title, "й", "ij")
	title = strings.ReplaceAll(title, "к", "k")
	title = strings.ReplaceAll(title, "л", "l")
	title = strings.ReplaceAll(title, "м", "m")
	title = strings.ReplaceAll(title, "н", "n")
	title = strings.ReplaceAll(title, "о", "o")
	title = strings.ReplaceAll(title, "п", "p")
	title = strings.ReplaceAll(title, "р", "r")
	title = strings.ReplaceAll(title, "с", "s")
	title = strings.ReplaceAll(title, "т", "t")
	title = strings.ReplaceAll(title, "у", "u")
	title = strings.ReplaceAll(title, "ф", "f")
	title = strings.ReplaceAll(title, "х", "h")
	title = strings.ReplaceAll(title, "ц", "ts")
	title = strings.ReplaceAll(title, "ч", "ch")
	title = strings.ReplaceAll(title, "ш", "sh")
	title = strings.ReplaceAll(title, "щ", "sh")
	title = strings.ReplaceAll(title, "ы", "i")
	title = strings.ReplaceAll(title, "ь", "")
	title = strings.ReplaceAll(title, "ъ", "")
	title = strings.ReplaceAll(title, "э", "e")
	title = strings.ReplaceAll(title, "ю", "ju")
	title = strings.ReplaceAll(title, "я", "ja")
	// s1 = translit(s0, "ru", reversed=True)
	// title = strings.ReplaceAll(title, "ej", "ey")
	// title = strings.ReplaceAll(title, "oj", "oy")
	// title = strings.ReplaceAll(title, "aj", "ay")
	// title = strings.ReplaceAll(title, "ij", "iy")
	// title = strings.ReplaceAll(title, "ja", "ya")
	// title = strings.ReplaceAll(title, "ju", "yu")
	return title
}

func Slugify(s string) string {
	s = strings.ReplaceAll(s, "(", "")
	s = strings.ReplaceAll(s, ")", "")
	s = strings.ReplaceAll(s, " ", "-")
	return s
}

func AddCanonical(rootDir string) error {
	var transliterate = func(fp string, fi os.DirEntry, err error) error {
		if fi.IsDir() {
			return nil
		}
		doc, err := ReadJson(fp)
		if err != nil {
			return err
		}
		dirty := false

		var searchHeading func(n *Node) (string, bool)

		searchHeading = func(n *Node) (string, bool) {
			if n.Type == Heading && n.Attributes != nil {
				level, ok1 := n.Attributes["level"]
				if ok1 && level == "1" {
					return n.Literal, true
				}
			}
			if n.Children != nil {
				for i := range n.Children {
					val, found := searchHeading(&n.Children[i])
					if found {
						return val, found
					}
				}
			}
			return "", false
		}
		heading, found := searchHeading(&doc)
		if found {
			if doc.Attributes == nil {
				doc.Attributes = map[string]string{}
			}
			doc.Attributes["canonical"] = strings.TrimSuffix(strings.TrimPrefix(fp, rootDir), ".md")
			doc.Attributes["title"] = heading
			doc.Attributes["path"] = Slugify(Transliterate(heading))
			dirty = true
		}
		if dirty {
			WriteJson(&doc, fp)
		}
		return nil
	}
	err := filepath.WalkDir(rootDir, transliterate)
	return err
}
