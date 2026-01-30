#!/usr/bin/env python3
import sys
import json

data = sys.stdin.read()
# Echo the AST back as a file named "ast.json"
output = {
    "files": [
        {
            "path": "ast.json",
            "content": data
        }
    ]
}
json.dump(output, sys.stdout)
