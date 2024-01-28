#!/usr/bin/env python3

import yaml
import sys
import os.path
import subprocess
from pathlib import Path

with open(sys.argv[1]) as foli:
  foliant = yaml.safe_load(foli)

chapters = foliant['chapters']

out = open(sys.argv[2], 'w')
base_dir = os.path.dirname(sys.argv[1]) + "/" + foliant['src_dir']+"/"
out_dir = os.path.dirname(sys.argv[2])+"/"

def heading(name):
   output = subprocess.check_output(['./marktome','heading',base_dir+name])
   return escape(output.strip().decode('utf-8'))

def escape(text):
   text = text.replace("&","\\&")
   return text

def append_file(name):
    # out.write("\\subsection{%s}\n" % heading(name))
    # out.write("\\include{%s}\n" % Path(name).stem)
    input = base_dir+name
    output = "temp.tex"
    subprocess.check_output(['./marktome','json2latex',input,output,"addheading","1"])
    out.write("\\newpage\n%% MARKDOWN %s\n" % name)
    with open("temp.tex") as temp:
       out.write(temp.read())
    


for ch in chapters:
  if isinstance(ch,str):
    append_file(ch)
  else:
    # if isinstance(ch,dict):
    if len(ch.keys()) != 1:
        print(ch)
        exit(5)
    title = list(ch.keys())[0]
    sections = ch[title]
    if isinstance(sections,list):
        out.write("\\part{%s}\n\n" % escape(title))
        for section0 in sections:
            if len(section0.keys()) != 1:
                print(ch)
                exit(5)
            section_title = list(section0.keys())[0]
            if section_title == "Watcher":
               continue
            out.write("\\chapter{%s}\n\n" % escape(section_title))
            section = section0[section_title]
            for entry in section:
                if isinstance(entry, str):
                    append_file(entry)
                else:
                    if len(entry.keys()) != 1:
                        print(entry)
                        exit(5)
                    subentry_title = list(entry.keys())[0]
                    out.write("\\section{%s}\n\n" % escape(subentry_title))
                    subentry = entry[subentry_title]
                    for se in subentry:
                        if not isinstance(se, str):
                          print(se)
                          exit(6)
                        append_file(se)
            pass
    else:
        # isinstance(sections,str):
        if sections.endswith(".pdf"):
            pass
        else:
            print(ch)
            exit(5)
  
