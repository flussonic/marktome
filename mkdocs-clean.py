#!/usr/bin/env python3

import yaml
import sys
import os.path

with open(sys.argv[1]) as f:
  mkdocs = yaml.safe_load(f)


del mkdocs['plugins']
mkdocs['theme']['custom_dir'] = "overrides"
mkdocs['docs_dir'] = "doc"
mkdocs['markdown_extensions'] = {
  "toc": {
    "permalink": []
  }
}

with open(sys.argv[1], "w") as f:
  f.write(yaml.dump(mkdocs, default_flow_style=False, sort_keys=False))
  
