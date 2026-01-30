#!/usr/bin/env python3
import json
import os
import sys

# Read generated AST
try:
    with open("gen/ast.json", "r") as f:
        data = json.load(f)
except FileNotFoundError:
    print("gen/ast.json not found")
    sys.exit(1)

ir = data.get("ir", {})
consts = ir.get("constants", [])

found = False
for c in consts:
    if c.get("name") == "OVERRIDE_WORKS":
        found = True
        break

if not found:
    print(f"OverrideWorks constant not found in AST. Keys in data: {list(data.keys())}. IR keys: {list(ir.keys())}")
    # Print const names found
    names = [c.get("name") for c in consts]
    print(f"Found consts: {names}")
    sys.exit(1)
