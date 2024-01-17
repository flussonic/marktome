all:
	go build
	rm -rf tmp tmp2
	./foli2 md2json ../erlydoc/src/ tmp2
	./foli2 planarize tmp2/ru
	./foli2 planarize tmp2/en
	./foli2 superlinks tmp2/ru
	./foli2 superlinks tmp2/en
	./foli2 snippets tmp2
	./foli2 json2md tmp2/en tmp/en/doc
	./foli2 json2md tmp2/ru tmp/ru/doc
	ls tmp/
	./foli2 foliant2mkdocs ../erlydoc/f2/foliant.flussonic.en.yml tmp2/en/mkdocs.yml
	./foli2 foliant2mkdocs ../erlydoc/f2/foliant.flussonic.ru.yml tmp2/ru/mkdocs.yml
	mv tmp2/en/mkdocs.yml tmp/en/mkdocs.yml
	mv tmp2/ru/mkdocs.yml tmp/ru/mkdocs.yml
	./mkdocs-clean.py tmp/en/mkdocs.yml
	./mkdocs-clean.py tmp/ru/mkdocs.yml
	cp -r ../erlydoc/f2/overrides tmp/en/overrides
	cp -r ../erlydoc/f2/overrides tmp/ru/overrides
	cd tmp/en && mkdocs build
	rm -rf tmp2
