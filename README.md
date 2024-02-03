# Marktome

**Marktome** (Markdown tome, i.e. large book processor) is a tool for people, who are struggling with multipage, multilanguage documentation generated from markdown files to static HTML and PDF.

Why not mkdocs? Incorrect: why not _only_ mkdocs?

Reason is in preprocessors. All multipage documentation systems like foliant, mkdocs are trying to do something interesting with markdown. If you want to have something like superlinks or xref, then you need to parse whole documentation source code, build index of headings and anchors and the process the links. Mkdocs offers approach with writing all this in python and dealing with objects and classes in Python.

## Workflow

Marktome offers another approach, which is less fragile and much more introspectable: use separate text files.

1. First all markdown files are parsed to well-defined primitive json
2. Now you can run preprocessors one by one
3. You can add your own preprocessor in bash script or maybe even in perl (joke of course, no perl please)
4. Now you have plenty of json files that are still useless
5. Marktome can generate markdown for mkdocs input or generate tex for building PDF

Points 1+ 5 can give you pretty good markdown formatter with stable output.

## Usage manual

```
make test
```

to ensure that all is ok.

```
go build
```

to produce file `marktome`

