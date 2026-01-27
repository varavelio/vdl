#!/usr/bin/env python3
"""
Verification script for generated_code test case.
This script is executed after vdl generate succeeds.
It verifies that the generated Python code is syntactically valid and can be imported.

Exit 0 = verification passed
Exit non-zero = verification failed (message on stderr)
"""
import json
import os
import sys

def main():
    gen_dir = os.path.join(os.path.dirname(__file__), "gen")
    
    # Verify generated Python file exists
    gen_py_file = os.path.join(gen_dir, "generated.py")
    if not os.path.exists(gen_py_file):
        print("ERROR: generated.py not found", file=sys.stderr)
        sys.exit(1)
    
    # Verify manifest exists
    manifest_file = os.path.join(gen_dir, "manifest.json")
    if not os.path.exists(manifest_file):
        print("ERROR: manifest.json not found", file=sys.stderr)
        sys.exit(1)
    
    # Verify Python syntax by compiling
    try:
        with open(gen_py_file, 'r') as f:
            source = f.read()
        compile(source, gen_py_file, 'exec')
    except SyntaxError as e:
        print(f"ERROR: Python syntax error: {e}", file=sys.stderr)
        sys.exit(1)
    
    # Verify manifest content
    with open(manifest_file, 'r') as f:
        manifest = json.load(f)
    
    if not manifest.get("success"):
        print("ERROR: manifest.success is not true", file=sys.stderr)
        sys.exit(1)
    
    stats = manifest.get("statistics", {})
    if stats.get("types", 0) == 0:
        print("ERROR: no types generated", file=sys.stderr)
        sys.exit(1)
    if stats.get("enums", 0) == 0:
        print("ERROR: no enums generated", file=sys.stderr)
        sys.exit(1)
    if stats.get("rpcs", 0) == 0:
        print("ERROR: no rpcs generated", file=sys.stderr)
        sys.exit(1)
    
    # Try to import the generated module
    sys.path.insert(0, gen_dir)
    try:
        import generated
        
        # Verify expected exports exist
        if not hasattr(generated, 'GENERATED_TYPES') or len(generated.GENERATED_TYPES) == 0:
            print("ERROR: GENERATED_TYPES is missing or empty", file=sys.stderr)
            sys.exit(1)
        if not hasattr(generated, 'GENERATED_ENUMS') or len(generated.GENERATED_ENUMS) == 0:
            print("ERROR: GENERATED_ENUMS is missing or empty", file=sys.stderr)
            sys.exit(1)
        if not hasattr(generated, 'GENERATED_RPCS') or len(generated.GENERATED_RPCS) == 0:
            print("ERROR: GENERATED_RPCS is missing or empty", file=sys.stderr)
            sys.exit(1)
        
        print(f"Types: {generated.GENERATED_TYPES}")
        print(f"Enums: {generated.GENERATED_ENUMS}")
        print(f"RPCs: {generated.GENERATED_RPCS}")
        print("SUCCESS: All generated code imports correctly")
        
    except ImportError as e:
        print(f"ERROR: Failed to import generated module: {e}", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"ERROR: Unexpected error importing module: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
