
all:
	go build
	rm -rf stage*
	mkdir -p stage-out/en/doc/img stage-out/ru/doc/img cache
	cp -r ../erlydoc/src stage-input
	cp ../erlydoc/f2/*.yml stage-input/
	sed -i '' 's|/usr/src/app/src/||' stage-input/preprocessors.yml
	sed -i '' 's|src/en|en|' stage-input/foliant.flussonic.en.yml
	sed -i '' 's|src/ru|ru|' stage-input/foliant.flussonic.ru.yml
	sed -i '' 's|src/ru|ru|' stage-input/foliant.watcher.en.yml

	cp -r ../erlydoc/f2/overrides stage-out/en/overrides
	cp -r ../erlydoc/f2/overrides stage-out/ru/overrides

	cp -r ../erlydoc/assets/* stage-out/en/doc/img
	cp -r ../erlydoc/images stage-out/en/doc/img/auto

	cp -r ../erlydoc/assets/* stage-out/ru/doc/img
	cp -r ../erlydoc/images stage-out/ru/doc/img/auto

	cp -r ../erlydoc/f2/template/flussonic.png stage-out/en/doc/img/
	cp -r ../erlydoc/f2/template/flussonic.png stage-out/ru/doc/img/

	cp ../erlydoc/f2/pdf/* stage-out/en/doc/
	cp ../erlydoc/f2/pdf/* stage-out/ru/doc/


	./marktome macros stage-input/foliant.flussonic.en.yml stage-input
	./marktome md2json stage-input stage-json
	cp stage-input/*.yml stage-json/
	./marktome planarize stage-json/foliant.flussonic.en.yml stage-planar/foliant.flussonic.en.yml
	./marktome planarize stage-json/foliant.flussonic.ru.yml stage-planar/foliant.flussonic.ru.yml

	./marktome superlinks stage-planar/en
	./marktome superlinks stage-planar/ru

	./marktome snippets stage-planar
	./marktome graphviz stage-planar/en stage-out/en/doc/img
	./marktome graphviz stage-planar/ru stage-out/ru/doc/img

	./marktome json2md stage-planar/en stage-out/en/doc
	./marktome json2md stage-planar/ru stage-out/ru/doc

	./marktome foliant2mkdocs stage-planar/foliant.flussonic.en.yml stage-out/en/mkdocs.yml
	./marktome foliant2mkdocs stage-planar/foliant.flussonic.ru.yml stage-out/ru/mkdocs.yml

	./mkdocs-clean.py stage-out/en/mkdocs.yml
	./mkdocs-clean.py stage-out/ru/mkdocs.yml

	# ./create-tex.py  stage-planar/foliant.flussonic.en.yml stage-out/en/doc/content.tex
	# docker run -i -e COLUMNS="`tput cols`" --rm -w /data -v `pwd`/stage-out/en/doc:/data -v `pwd`/cache:/data/cache latex pdf.sh

	./create-tex.py  stage-planar/foliant.flussonic.ru.yml stage-out/ru/doc/content.tex
	docker run -i -e COLUMNS="`tput cols`" --rm -w /data -v `pwd`/stage-out/ru/doc:/data -v `pwd`/cache:/data/cache latex pdf.sh

	# cd stage-planar/ru && mkdocs build
	# cd stage-planar/en && mkdocs build

test:
	go build
	go test -v marktome/md2json

pdf:
	go build
	cp ../erlydoc/f2/pdf/* stage-out/en/doc/
	./marktome json2latex stage-planar/en/mobile-apps-for-accessing-watcher.md stage-out/en/doc/content.tex
	docker run -i -e COLUMNS="`tput cols`" --rm -w /data -v `pwd`/stage-out/en/doc:/data latex pdf.sh
