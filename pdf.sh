#!/bin/bash

echo "Phase 1 silenced to compile-pdf1.log"
pdflatex -synctex=1 -interaction=nonstopmode -file-line-error --shell-escape main.tex > compile-pdf1.log  2>&1
echo "Phase 2" 
pdflatex -synctex=1 -interaction=nonstopmode -file-line-error --shell-escape main.tex 2>&1 | tee compile-pdf2.log
