VERSION ?= $(shell git describe --abbrev=7 --long | sed 's/^v//g'| awk -F '-g' '{print $1}')
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
	# docker build -t latex -f Dockerfile.pandoc .
	../erlydoc/f2/split-sources.sh ../erlydoc/src stage-input
	mkdir -p stage-planar/img stage-out/en/doc stage-out/ru/doc cache
	cp ../erlydoc/f2/*.yml stage-input/

	cp -r ../erlydoc/f2/overrides stage-out/overrides

	cp -r ../erlydoc/assets/* stage-planar/img
	cp -r ../erlydoc/f2/template/flussonic.png stage-planar/img/

	cp ../erlydoc/f2/pdf/* stage-out/en/
	cp ../erlydoc/f2/pdf/* stage-out/ru/


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
	./marktome copy-images stage-planar/en stage-planar/ stage-out/en/
	./marktome copy-images stage-planar/ru stage-planar/ stage-out/ru/

	./marktome json2md stage-planar/en stage-out/en
	./marktome json2md stage-planar/ru stage-out/ru

	cp stage-planar/foliant.flussonic.en.yml stage-out/mkdocs.en.yml
	cp stage-planar/foliant.flussonic.ru.yml stage-out/mkdocs.ru.yml

	../erlydoc/f2/create-tex.py  stage-planar/foliant.flussonic.en.yml stage-out/en/content.tex
	../erlydoc/f2/create-tex.py  stage-planar/foliant.flussonic.ru.yml stage-out/ru/content.tex


	# docker run -i -e COLUMNS="`tput cols`" --rm -w /data -v `pwd`/stage-out/doc:/data -v `pwd`/cache:/data/cache latex pdf.sh
	# docker run -i -e COLUMNS="`tput cols`" --rm -w /data -v `pwd`/stage-out/doc:/data -v `pwd`/cache:/data/cache latex pdf.sh

	cd stage-out && mkdocs build -f mkdocs.en.yml -d flussonic_en
	cd stage-out && mkdocs build -f mkdocs.ru.yml -d flussonic_ru

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
