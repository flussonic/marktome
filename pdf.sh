#!/bin/bash

pdflatex -synctex=1 -interaction=nonstopmode -file-line-error --shell-escape main.tex
pdflatex -synctex=1 -interaction=nonstopmode -file-line-error --shell-escape main.tex
