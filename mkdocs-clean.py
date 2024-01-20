#!/usr/bin/env python3

import yaml
import sys
import os.path

with open(sys.argv[1]) as f:
  mkdocs = yaml.safe_load(f)


if 'plugins' in mkdocs:
  mkdocs['plugins'].remove("redirects")

mkdocs['theme']['custom_dir'] = "overrides"
mkdocs['docs_dir'] = "doc"
mkdocs['markdown_extensions'].append({"toc": {
  "permalink": []
}})
mkdocs['markdown_extensions'].append("attr_list")


with open(sys.argv[1], "w") as f:
  f.write(yaml.dump(mkdocs, default_flow_style=False, sort_keys=False))
  
