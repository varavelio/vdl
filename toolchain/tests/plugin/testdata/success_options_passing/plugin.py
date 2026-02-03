#!/usr/bin/env python3
"""
Plugin that validates options are passed correctly from vdl.yaml.
Tests flat key-value options (map<string>).
"""
import json
import sys

def main():
    # Read input from stdin
    input_data = json.load(sys.stdin)
    
    ir = input_data.get("ir", {})
    options = input_data.get("options", {})
    
    # Validate options are passed correctly (flat key-value)
    assert options is not None, "Options should not be None"
    assert options.get("language") == "python", f"Expected language='python', got {options.get('language')}"
    assert options.get("version") == "3.11", f"Expected version='3.11', got {options.get('version')}"
    
    # Generate output reflecting the options received
    output = {
        "files": [
            {
                "path": "options_received.json",
                "content": json.dumps({
                    "options_received": options,
                    "validation": "passed"
                }, indent=2)
            }
        ]
    }
    
    json.dump(output, sys.stdout)

if __name__ == "__main__":
    main()
