#!/usr/bin/env python3
"""
Plugin that validates an empty schema still produces a valid IR.
Tests edge case: empty schema with no types, enums, rpcs, etc.
"""
import json
import sys

def main():
    input_data = json.load(sys.stdin)
    ir = input_data.get("ir", {})
    
    errors = []
    
    # All IR arrays should exist and be empty (or null/None for empty schema)
    expected_arrays = ["types", "enums", "constants", "patterns", "rpcs", "procedures", "streams", "docs"]
    for arr_name in expected_arrays:
        value = ir.get(arr_name)
        # Accept None/null or empty list for empty schema
        if value is None:
            continue  # None is acceptable for empty schema
        elif not isinstance(value, list):
            errors.append(f"IR.{arr_name} should be array or null, got {type(value)}")
        elif len(value) != 0:
            errors.append(f"IR.{arr_name} should be empty for empty schema, got {len(value)} items")
    
    if errors:
        for err in errors:
            print(f"ERROR: {err}", file=sys.stderr)
        output = {
            "files": [{
                "path": "result.json",
                "content": json.dumps({"success": False, "errors": errors}, indent=2)
            }]
        }
        json.dump(output, sys.stdout)
        sys.exit(1)
    
    output = {
        "files": [{
            "path": "result.json",
            "content": json.dumps({
                "success": True,
                "message": "Empty schema produced valid IR with empty arrays",
                "ir_keys": list(ir.keys())
            }, indent=2)
        }]
    }
    json.dump(output, sys.stdout)

if __name__ == "__main__":
    main()
