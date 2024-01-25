all:
	go build
	rm -rf stage0 stage1 stage2
	cp -r ../erlydoc/src stage0
	cp ../erlydoc/f2/*.yml stage0/
	sed -i '' 's|/usr/src/app/src/||' stage0/preprocessors.yml
	./foli2 macros stage0 stage0/foliant.flussonic.en.yml
	./foli2 md2json stage0 stage1
	./foli2 snippets stage1
	./foli2 graphviz stage1/en stage1/en/img
	./foli2 graphviz stage1/ru stage1/ru/img

	./foli2 planarize stage1/en
	./foli2 planarize stage1/ru

	./foli2 superlinks stage1/en
	./foli2 superlinks stage1/ru

	./foli2 json2md stage1/en stage2/en/doc
	./foli2 json2md stage1/ru stage2/ru/doc

	./foli2 foliant2mkdocs stage0/foliant.flussonic.en.yml stage2/en/mkdocs.yml
	./foli2 foliant2mkdocs stage0/foliant.flussonic.ru.yml stage2/ru/mkdocs.yml

	mv stage0/*.yml stage1/

	sed -i '' 's|src/en|en|' stage1/foliant.flussonic.en.yml
	sed -i '' 's|src/ru|ru|' stage1/foliant.flussonic.ru.yml

	./mkdocs-clean.py stage2/en/mkdocs.yml
	./mkdocs-clean.py stage2/ru/mkdocs.yml

	cp -r ../erlydoc/f2/overrides stage2/en/overrides
	cp -r ../erlydoc/f2/overrides stage2/ru/overrides

	cp -r ../erlydoc/assets stage2/en/doc/img
	cp -r ../erlydoc/images stage2/en/doc/img/auto

	cp -r ../erlydoc/assets stage2/ru/doc/img
	cp -r ../erlydoc/images stage2/ru/doc/img/auto

	cp -r stage1/en/img/* stage2/en/doc/img/
	cp -r stage1/ru/img/* stage2/ru/doc/img/

	cp -r ../erlydoc/f2/template/flussonic.png stage2/en/doc/img/
	cp -r ../erlydoc/f2/template/flussonic.png stage2/ru/doc/img/

	cp ../erlydoc/f2/pdf/* stage2/en/doc/
	./foli2 json2latex stage1/foliant.flussonic.en.yml stage2/en/doc/content.tex
	docker run -i --rm -w /data -v `pwd`/stage2/en/doc:/data latex pdf.sh
		
	# cd stage2/ru && mkdocs build
	# cd stage2/en && mkdocs build
	# rm -rf stage1

	