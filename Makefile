all:
	go build
	rm -rf stage0 stage1 stage2
	cp -r ../erlydoc/src stage0
	./foli2 macros stage0 ../erlydoc/f2/foliant.flussonic.en.yml
	./foli2 md2json stage0 stage1
	./foli2 planarize stage1/ru
	./foli2 planarize stage1/en
	./foli2 superlinks stage1/ru
	./foli2 superlinks stage1/en
	./foli2 snippets stage1
	./foli2 json2md stage1/en stage2/en/doc
	./foli2 json2md stage1/ru stage2/ru/doc
	./foli2 foliant2mkdocs ../erlydoc/f2/foliant.flussonic.en.yml stage1/en/mkdocs.yml
	./foli2 foliant2mkdocs ../erlydoc/f2/foliant.flussonic.ru.yml stage1/ru/mkdocs.yml
	mv stage1/en/mkdocs.yml stage2/en/mkdocs.yml
	mv stage1/ru/mkdocs.yml stage2/ru/mkdocs.yml
	./mkdocs-clean.py stage2/en/mkdocs.yml
	./mkdocs-clean.py stage2/ru/mkdocs.yml
	cp -r ../erlydoc/f2/overrides stage2/en/overrides
	cp -r ../erlydoc/f2/overrides stage2/ru/overrides
	cp -r ../erlydoc/assets stage2/en/doc/img
	cp -r ../erlydoc/images stage2/en/doc/img/auto
	cp -r ../erlydoc/assets stage2/ru/doc/img
	cp -r ../erlydoc/images stage2/ru/doc/img/auto
	cd stage2/en && mkdocs build
	cd stage2/ru && mkdocs build
	# rm -rf stage1
