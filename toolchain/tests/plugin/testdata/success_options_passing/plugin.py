#!/usr/bin/env python3
"""
Plugin that validates options are passed correctly from vdl.yaml.
"""
import json
import sys

def main():
    # Read input from stdin
    input_data = json.load(sys.stdin)
    
    ir = input_data.get("ir", {})
    options = input_data.get("options", {})
    
    # Validate options are passed correctly
    assert options is not None, "Options should not be None"
    assert options.get("language") == "python", f"Expected language='python', got {options.get('language')}"
    assert options.get("version") == "3.11", f"Expected version='3.11', got {options.get('version')}"
    
    # Validate nested array
    features = options.get("features", [])
    assert "async" in features, "Expected 'async' in features"
    assert "dataclasses" in features, "Expected 'dataclasses' in features"
    
    # Validate nested object
    config = options.get("config", {})
    assert config.get("indent") == 4, f"Expected indent=4, got {config.get('indent')}"
    assert config.get("use_slots") == True, f"Expected use_slots=True, got {config.get('use_slots')}"
    
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
