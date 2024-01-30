VERSION ?= $(shell git describe --abbrev=7 --long | sed 's/^v//g')
ifeq (,$(BRANCH))
	ifneq (,$(CI_BUILD_REF_SLUG))
		BRANCH=$(CI_BUILD_REF_SLUG)
	else
	  BRANCH=$(shell git rev-parse --abbrev-ref HEAD| sed 's/\//-/')
	endif
endif



all:
	go build
	rm -rf stage*
	docker build -t latex -f Dockerfile.pandoc .
	cp -r ../erlydoc/src stage-input
	mkdir -p stage-planar/img stage-out/en/doc stage-out/ru/doc cache
	cp ../erlydoc/f2/*.yml stage-input/
	sed -i '' 's|/usr/src/app/src/||' stage-input/preprocessors.yml
	sed -i '' 's|src/en|en|' stage-input/foliant.flussonic.en.yml
	sed -i '' 's|src/ru|ru|' stage-input/foliant.flussonic.ru.yml
	sed -i '' 's|src/ru|ru|' stage-input/foliant.watcher.en.yml

	cp -r ../erlydoc/f2/overrides stage-out/en/overrides
	cp -r ../erlydoc/f2/overrides stage-out/ru/overrides

	cp -r ../erlydoc/assets/* stage-planar/img
	cp -r ../erlydoc/images stage-planar/img/auto
	cp -r ../erlydoc/f2/template/flussonic.png stage-planar/img/

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
	./marktome graphviz stage-planar/en stage-planar/img cache
	./marktome graphviz stage-planar/ru stage-planar/img cache
	./marktome copy-images stage-planar/en stage-planar/ stage-out/en/doc/

	exit 4
	./marktome json2md stage-planar/en stage-out/en/doc
	./marktome json2md stage-planar/ru stage-out/ru/doc

	./marktome foliant2mkdocs stage-planar/foliant.flussonic.en.yml stage-out/en/mkdocs.yml
	./marktome foliant2mkdocs stage-planar/foliant.flussonic.ru.yml stage-out/ru/mkdocs.yml

	./mkdocs-clean.py stage-out/en/mkdocs.yml
	./mkdocs-clean.py stage-out/ru/mkdocs.yml

	./create-tex.py  stage-planar/foliant.flussonic.en.yml stage-out/en/doc/content.tex
	./create-tex.py  stage-planar/foliant.flussonic.ru.yml stage-out/ru/doc/content.tex


	# docker run -i -e COLUMNS="`tput cols`" --rm -w /data -v `pwd`/stage-out/en/doc:/data -v `pwd`/cache:/data/cache latex pdf.sh
	# docker run -i -e COLUMNS="`tput cols`" --rm -w /data -v `pwd`/stage-out/ru/doc:/data -v `pwd`/cache:/data/cache latex pdf.sh

	cd stage-out/ru && mkdocs build
	cd stage-out/en && mkdocs build

test:
	go build
	go test -v marktome/md2json

pdf:
	go build
	cp ../erlydoc/f2/pdf/* stage-out/en/doc/
	./marktome json2latex stage-planar/en/mobile-apps-for-accessing-watcher.md stage-out/en/doc/content.tex
	docker run -i -e COLUMNS="`tput cols`" --rm -w /data -v `pwd`/stage-out/en/doc:/data latex pdf.sh


deb:
	docker build -t marktome-build:${BRANCH} --build-arg VERSION=${VERSION} .

upload:
	docker run --rm -e REPOSITORY_SECRET marktome-build:${BRANCH} autodeb.py upload marktome_${VERSION}_all.deb ${REPO}/marktome_${VERSION}_all.deb
